package wzlib

import (
	"errors"
	"fmt"
	"golang.org/x/text/encoding"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type WzStructure struct {
	WzFiles             []*WzFile         // 已加载的 Wz 文件列表
	Encryption          *WzCrypto         // 加密处理逻辑
	WzNode              *WzNode           // Wz 文件的目录树根节点
	ImgNumber           int               // 加载的图像数量
	HasBaseWz           bool              // 是否存在基础 Wz 文件
	TextEncoding        encoding.Encoding // 文件的文本编码方式
	AutoDetectExtFiles  bool              // 是否自动检测扩展文件
	ImgCheckDisabled    bool              // 是否禁用图像校验
	WzVersionVerifyMode int               // 版本验证模式
}

// LoadWzFile loads a WZ file into the structure
func (ws *WzStructure) LoadWzFile(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("fileName cannot be empty")
	}

	wzFile, err := NewWzFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to create WzFile: %v", err)
	}

	if wzFile.Loaded {
		ws.WzFiles = append(ws.WzFiles, wzFile)
	} else {
		return errors.New("not a wz file")
	}

	if ws.Encryption == nil {
		ws.Encryption = NewWzCrypto()
		err = ws.Encryption.DetectEncryption(wzFile)
		if err != nil {
			return err
		}
	}

	ws.WzNode = NewWzNode(filepath.Base(fileName))
	_, err = wzFile.FileStream.Seek(wzFile.Header.DataStartPosition, io.SeekStart)
	if err != nil {
		return err
	}
	wzFile.WzStructure = ws
	err = wzFile.GetDirTree(ws.WzNode)
	if err != nil {
		return err
	}
	return nil
}

// ResetEncryption resets the encryption state
func (ws *WzStructure) ResetEncryption() {
	ws.Encryption.Reset()
}

// DetectEncryption detects the encryption type for a WZ file
func (ws *WzStructure) DetectEncryption(wzFile *WzFile) error {
	if wzFile == nil {
		return fmt.Errorf("wzFile cannot be nil")
	}

	return ws.Encryption.DetectEncryption(wzFile)
}

// ToString returns a string representation of the structure
func (ws *WzStructure) ToString() string {
	return fmt.Sprintf("WzStructure with %d WzFile(s).", len(ws.WzFiles))
}

func (ws *WzStructure) LoadWzFolder(folder string, node *WzNode, useBaseWz bool) error {
	baseName := filepath.Join(folder, filepath.Base(folder))
	entryWzFileName := baseName + ".wz"
	iniFileName := baseName + ".ini"
	extraWzFileName := func(index int) string {
		return fmt.Sprintf("%s_%03d.wz", baseName, index)
	}

	var lastWzIndex *int
	if iniData, err := os.ReadFile(iniFileName); err == nil {
		lines := strings.Split(string(iniData), "\n")
		for _, line := range lines {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 && parts[0] == "LastWzIndex" {
				if index, err := strconv.Atoi(parts[1]); err == nil {
					lastWzIndex = &index
					break
				}
			}
		}
	}

	if lastWzIndex == nil {
		for i := 0; ; i++ {
			extraFile := extraWzFileName(i)
			if _, err := os.Stat(extraFile); os.IsNotExist(err) {
				break
			}
			lastWzIndex = &i
		}
	}

	if node == nil {
		node = NewWzNode(filepath.Base(entryWzFileName))
	}

	entryWzf, err := ws.LoadFile(entryWzFileName, node, useBaseWz, true)
	if err != nil {
		return fmt.Errorf("LoadFile entry failed: %v", err)
	}

	if lastWzIndex != nil {
		for i := 0; i <= *lastWzIndex; i++ {
			extraFile := extraWzFileName(i)
			tempNode := NewWzNode(filepath.Base(extraFile))
			extraWzf, err := ws.LoadFile(extraFile, tempNode, false, true)
			if err != nil {
				continue
			}
			entryWzf.MergeWzFile(extraWzf)
		}
	}

	return nil
}

func (ws *WzStructure) LoadFile(fileName string, node *WzNode, useBaseWz, loadWzAsFolder bool) (*WzFile, error) {
	wzFile, err := NewWzFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create WzFile: %v", err)
	}
	if !wzFile.Loaded {
		return nil, fmt.Errorf("the file is not a valid wz file")
	}

	ws.WzFiles = append(ws.WzFiles, wzFile)
	wzFile.TextEncoding = ws.TextEncoding

	err = ws.Encryption.DetectEncryption(wzFile)
	if err != nil {
		return nil, err
	}

	node.Value = wzFile
	wzFile.Node = node

	if _, err := wzFile.FileStream.Seek(wzFile.Header.DataStartPosition, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to data start: %v", err)
	}

	err = wzFile.GetDirTree(node)
	if err != nil {
		//wzFile.FileStream.Close()
		return nil, fmt.Errorf("failed to read directory tree: %v", err)
	}

	if pos, err := wzFile.FileStream.Seek(0, io.SeekCurrent); err == nil {
		wzFile.Header.DirEndPosition = pos
	}

	//wzFile.DetectWzType()
	//wzFile.DetectWzVersion()

	return wzFile, nil
}
