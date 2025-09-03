package wzlib

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/encoding"
)

/*type WzDirectory struct {
	Name    string
	Entries []*WzEntry
}*/

type WzFile struct {
	FileName      string
	FileStream    *WzBinaryReader
	Header        *WzHeader
	Directories   []*WzDirectory
	MergedWzFiles []*WzFile
	OwnerWzFile   *WzFile
	Node          *WzNode
	Loaded        bool
	WzStructure   *WzStructure
	TextEncoding  encoding.Encoding // 文件的文本编码方式
}

func NewWzFile(fileName string) (*WzFile, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	wzFile := &WzFile{
		FileName:    fileName,
		FileStream:  NewWzBinaryReader(file),
		Directories: []*WzDirectory{},
	}

	// Parse header
	err = wzFile.GetHeader()
	if err != nil {
		file.Close()
		return nil, err
	}

	wzFile.Loaded = true
	return wzFile, nil
}

func (wf *WzFile) GetHeader() error {
	wf.FileStream.Seek(0, io.SeekStart)

	// Read signature
	signature := make([]byte, 4)
	_, err := wf.FileStream.Read(signature)
	if err != nil {
		return err
	}
	if string(signature) != "PKG1" {
		return errors.New("invalid file signature")
	}

	dataSize, err := wf.FileStream.ReadInt64()
	if err != nil {
		return err
	}
	headerSize, err := wf.FileStream.ReadInt32()
	if err != nil {
		return err
	}
	copyright := make([]byte, headerSize-int32(wf.FileStream.Pos()))

	_, err = wf.FileStream.Read(copyright)
	if err != nil {
		return err
	}
	var encverMissing = false
	var encver = int16(-1)
	if dataSize >= 2 {
		wf.FileStream.Seek(int64(headerSize), io.SeekStart)
		encver, err = wf.FileStream.ReadInt16()
		if err != nil {
			return err
		}
		if encver > 0xFF {
			encverMissing = true
		} else if encver == 0x80 {
			if dataSize >= 5 {
				wf.FileStream.Seek(int64(headerSize), io.SeekStart)
				propCount, err := wf.FileStream.ReadCompressedInt32()
				if err != nil {
					return err
				}
				if propCount > 0 && (propCount&0xFF) == 0 && propCount <= 0xffff {
					encverMissing = true
				}
			}
		}
	} else {
		encverMissing = true
	}
	dataStartPos := headerSize + func() int32 {
		if encverMissing {
			return 0
		} else {
			return 2
		}
	}()

	// Read additional header fields
	header := &WzHeader{}
	header.DataSize = dataSize
	header.HeaderSize = headerSize
	header.DataStartPosition = int64(dataStartPos)
	if encverMissing {
	} else {
		header.SetOrdinalVersionDetector(int(encver))
		header.VersionDetector.TryGetNextVersion()
	}
	wf.Header = header
	return nil
}

// GetDirTree parses the directory tree structure from the WZ file
func (wf *WzFile) GetDirTree2(reader *WzBinaryReader, parent *WzNode, useBaseWz bool, loadWzAsFolder bool) error {
	dirs := []string{}
	count, err := reader.ReadCompressedInt32()
	if err != nil {
		return fmt.Errorf("failed to read directory count: %v", err)
	}
	cryptoKey := wf.WzStructure.Encryption.Keys

	for i := 0; i < int(count); i++ {
		nodeType, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("failed to read node type: %v", err)
		}

		var name string
		switch nodeType {
		case 0x02:
			stringOffAdd := -1
			if wf.Header.HasCapabilities(WzCapabilitiesEncverMissing) {
				stringOffAdd = 2
			}
			offset, err := reader.ReadInt32()
			if err != nil {
				return fmt.Errorf("failed to read string offset: %v", err)
			}
			name, err = reader.ReadStringAt(int64(int(offset)+stringOffAdd), cryptoKey)
			if err != nil {
				return fmt.Errorf("failed to read string at offset: %v", err)
			}
		case 0x04, 0x03:
			name, err = reader.ReadString(cryptoKey)
			if err != nil {
				return fmt.Errorf("failed to read string: %v", err)
			}
		default:
			return fmt.Errorf("unknown node type: %x", nodeType)
		}

		size, err := reader.ReadCompressedInt32()
		if err != nil {
			return fmt.Errorf("failed to read size: %v", err)
		}

		cs32, err := reader.ReadCompressedInt32()
		if err != nil {
			return fmt.Errorf("failed to read checksum: %v", err)
		}

		pos := reader.Pos()
		hashOffset, err := reader.ReadUInt32()
		if err != nil {
			return fmt.Errorf("failed to read hash offset: %v", err)
		}

		switch nodeType {
		case 0x02, 0x04:
			img := NewWzImage(name, int(size), int(cs32), uint32(hashOffset), uint32(pos), wf)
			childNode := parent.AddChild(img.Node)
			childNode.Value = img
			//img.OwnerNode = childNode
		case 0x03:
			dir := NewWzDirectory(name, int(size), int(cs32), uint32(hashOffset), uint32(pos), wf)
			wf.Directories = append(wf.Directories, dir)
			dirs = append(dirs, name)
		}
	}
	dirCount := len(dirs)
	willLoadBaseWz := useBaseWz && strings.EqualFold(parent.Text, "base.wz")

	baseFolder := filepath.Dir(wf.Header.FileName)

	if willLoadBaseWz && wf.WzStructure.AutoDetectExtFiles {
		alphaRegex := regexp.MustCompile(`^[A-Za-z]+$`)
		for i := 0; i < dirCount; i++ {
			dir := dirs[i]
			if alphaRegex.MatchString(dir) {
				wzTypeName := dir

				// Detect extended WZ files
				for fileID := 2; ; fileID++ {
					extDirName := fmt.Sprintf("%s%d", wzTypeName, fileID)
					extWzFile := filepath.Join(baseFolder, extDirName+".wz")
					if _, err := os.Stat(extWzFile); os.IsNotExist(err) {
						break
					}
					if !containsIgnoreCase(dirs[:dirCount], extDirName) {
						dirs = append(dirs, extDirName)
					}
				}

				// Detect KMST1058-style WZ files
				for fileID := 1; ; fileID++ {
					extDirName := fmt.Sprintf("%s%03d", wzTypeName, fileID)
					extWzFile := filepath.Join(baseFolder, extDirName+".wz")
					if _, err := os.Stat(extWzFile); os.IsNotExist(err) {
						break
					}
					if !containsIgnoreCase(dirs[:dirCount], extDirName) {
						dirs = append(dirs, extDirName)
					}
				}
			}
		}
	}

	for i, dir := range dirs {
		node := parent.AddChild(NewWzNode(dir))
		if i < dirCount {
			if err := wf.GetDirTree(node); err != nil {
				return fmt.Errorf("failed to get directory tree for %s: %v", dir, err)
			}
		}

		if len(node.Nodes) == 0 {
			wf.WzStructure.HasBaseWz = wf.WzStructure.HasBaseWz || willLoadBaseWz
			if loadWzAsFolder {
				//todo
				wzFolder := filepath.Join(baseFolder, dir)
				if willLoadBaseWz {
					wzFolder = filepath.Join(filepath.Dir(baseFolder), dir)
				}
				if _, err := os.Stat(wzFolder); err == nil {
					if err := wf.WzStructure.LoadWzFolder(wzFolder, node, false); err != nil {
						return fmt.Errorf("failed to load WZ folder: %v", err)
					}
				}
			} else if willLoadBaseWz {
				//todo
				/*filePath := filepath.Join(baseFolder, dir+".wz")
				if _, err := os.Stat(filePath); err == nil {
					if err := wf.WzStructure.LoadFile(filePath, node, false, loadWzAsFolder); err != nil {
						return fmt.Errorf("failed to load WZ file: %v", err)
					}
				}*/

			}
		}
	}

	//parent.TrimChildren()
	return nil
}

func getStreamLength(stream io.ReadSeeker) (int64, error) {
	current, err := stream.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	end, err := stream.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	_, err = stream.Seek(current, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return end, nil
}

func (wf *WzFile) GetDirTree(parent *WzNode) error {
	length, err := getStreamLength(wf.FileStream)
	if err != nil {
		return err
	}
	reader, err := NewPartialStream(wf.FileStream.File(), wf.Header.DataStartPosition, length-wf.Header.DataStartPosition)
	if err != nil {
		return err
	}
	return wf.getDirTree(NewWzBinaryReader(reader), parent, false, false)
}

func (wf *WzFile) getDirTree(reader *WzBinaryReader, parent *WzNode, useBaseWz bool, loadWzAsFolder bool) error {
	dirs := []*WzDirectory{}
	count, err := reader.ReadCompressedInt32()
	if err != nil {
		return fmt.Errorf("failed to read directory count: %v", err)
	}
	cryptoKey := wf.WzStructure.Encryption.Keys

	for i := 0; i < int(count); i++ {
		nodeType, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("failed to read node type: %v", err)
		}

		var name string
		switch nodeType {
		case 0x02:
			stringOffAdd := -1
			if wf.Header.HasCapabilities(WzCapabilitiesEncverMissing) {
				stringOffAdd = 2
			}
			offset, err := reader.ReadInt32()
			if err != nil {
				return fmt.Errorf("failed to read string offset: %v", err)
			}
			name, err = reader.ReadStringAt(int64(offset)+int64(stringOffAdd), cryptoKey)
			if err != nil {
				return fmt.Errorf("failed to read string at offset: %v", err)
			}
		case 0x03, 0x04:
			name, err = reader.ReadString(cryptoKey)
			if err != nil {
				return fmt.Errorf("failed to read string: %v", err)
			}
		default:
			return fmt.Errorf("unknown node type: %x", nodeType)
		}

		size, err := reader.ReadCompressedInt32()
		if err != nil {
			return err
		}

		cs32, err := reader.ReadCompressedInt32()
		if err != nil {
			return err
		}

		offset := reader.Pos() + wf.FileStream.Pos()

		hashOffset, err := reader.ReadUInt32()
		if err != nil {
			return err
		}

		switch nodeType {
		case 0x02, 0x04:

			img := NewWzImage(name, int(size), int(cs32), hashOffset, uint32(offset), wf)
			child := parent.AddChild(img.Node)
			child.Value = img

		case 0x03:
			dir := &WzDirectory{
				Name:         name,
				Offset:       int(offset),
				Size:         int(size),
				Checksum:     int(cs32),
				HashedOffset: hashOffset,
				WzFile:       wf,
			}
			wf.Directories = append(wf.Directories, dir)
			dirs = append(dirs, dir)
		}
	}

	for _, dir := range dirs {
		child := parent.AddChild(NewWzNode(dir.Name))

		// Seek 到子目录偏移位置，构造新 reader
		/*_, err := wf.FileStream.Seek(int64(dir.Offset), io.SeekStart)
		if err != nil {
			return fmt.Errorf("seek to dir %s offset failed: %v", dir.Name, err)
		}*/
		//subReader := NewWzBinaryReader(wf.FileStream)

		err = wf.getDirTree(reader, child, useBaseWz, loadWzAsFolder)
		if err != nil {
			return fmt.Errorf("read subdir %s failed: %v", dir.Name, err)
		}
	}

	return nil
}

func (wf *WzFile) MergeWzFile(other *WzFile) {
	// 将 other 的所有子节点移动到当前文件的根节点
	children := make([]*WzNode, len(other.Node.Nodes))
	copy(children, other.Node.Nodes)

	// 清空 other 的节点列表
	other.Node.Nodes = nil

	// 添加到当前 WzFile 的根节点下
	for _, child := range children {
		wf.Node.AddChild(child)
	}

	// 初始化 mergedWzFiles（延迟）
	if wf.MergedWzFiles == nil {
		wf.MergedWzFiles = []*WzFile{}
	}
	wf.MergedWzFiles = append(wf.MergedWzFiles, other)

	// 设置 owner
	other.OwnerWzFile = wf
}

// containsIgnoreCase checks if a slice contains a string (case-insensitive)
func containsIgnoreCase(slice []string, str string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, str) {
			return true
		}
	}
	return false
}

/*func (wf *WzFile) GetDirTree(reader *WzBinaryReader, parent *WzNode, useBaseWz bool, loadWzAsFolder bool) error {
	var dirs []string
	count, err := reader.ReadCompressedInt32()
	if err != nil {
		return err
	}

	// Get the encryption key if needed
	cryptoKey := wf.WzStructure.encryption.keys

	for i := int32(0); i < count; i++ {
		nodeType, err := reader.ReadByte()
		if err != nil {
			return err
		}

		var name string
		switch nodeType {
		case 0x02:
			stringOffAdd := 0
			if wf.Header.HasCapabilities(WzCapabilitiesEncverMissing) {
				stringOffAdd = 2
			} else {
				stringOffAdd = -1
			}
			offset, err := reader.ReadInt32()
			if err != nil {
				return err
			}
			name, err = reader.ReadStringAt(int64(offset+int32(stringOffAdd)), cryptoKey)
		case 0x04, 0x03:
			name, err = reader.ReadString(cryptoKey)
		default:
			return fmt.Errorf("unknown type %x in WzDirTree", nodeType)
		}

		if err != nil {
			return err
		}

		// Read size and checksum
		size, err := reader.ReadCompressedInt32()
		if err != nil {
			return err
		}
		cs32, err := reader.ReadCompressedInt32()
		if err != nil {
			return err
		}

		pos := uint32(wf.FileStream.Position)
		hashOffset, err := reader.ReadUInt32()
		if err != nil {
			return err
		}

		// Handle based on node type
		switch nodeType {
		case 0x02, 0x04:
			img := NewWzImage(name, size, cs32, hashOffset, pos, wf)
			childNode := parent.Nodes.Add(name)
			// Further logic for handling image or directory...
		}
	}
	return nil
}

func (wf *WzFile) ParseDirectory() error {
	for _, dir := range wf.Directories {
		err := dir.Parse()
		if err != nil {
			return err
		}
	}
	return nil
}

func (wf *WzFile) Close() error {
	if wf.FileStream != nil {
		return wf.FileStream.Close()
	}
	return nil
}
*/
