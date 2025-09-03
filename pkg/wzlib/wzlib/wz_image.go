package wzlib

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"math/bits"
)

type WzImage struct {
	Name              string
	Size              int
	Checksum          int
	HashedOffset      uint32
	HashedOffsetPos   uint32
	Offset            int64
	WzFile            *WzFile
	Node              *WzNode
	Extracted         bool
	ChecksumChecked   bool
	EncryptionChecked bool
	EncryptionType    WzCryptoKeyType
	Stream            io.ReadSeeker
	Type              string
}

// NewWzImage creates a new WzImage instance
func NewWzImage(name string, size, checksum int, hashOffset, hashPos uint32, wzFile *WzFile) *WzImage {
	wz := &WzImage{
		Name:            name,
		Size:            size,
		Checksum:        checksum,
		HashedOffset:    hashOffset,
		HashedOffsetPos: hashPos,
		WzFile:          wzFile,
		Node:            NewWzNode(name),
	}
	wz.Offset = int64(wz.WzFile.CalcOffset(hashPos, hashOffset))

	stream, err := NewPartialStream(wzFile.FileStream.File(), wz.Offset, int64(wz.Size))
	if err != nil {
		// fallback: 将数据读入内存，使用 bytes.Reader（确保线程安全）
		buf := make([]byte, wz.Size)
		_, _ = wzFile.FileStream.BaseStream.Seek(wz.Offset, io.SeekStart)
		_, _ = io.ReadFull(wzFile.FileStream.BaseStream, buf)
		stream = bytes.NewReader(buf)
	}

	wz.Stream = stream
	return wz
}

func NewImageStream(r *WzBinaryReader, offset int64, size int64) (io.ReadSeeker, error) {
	if r.ReaderAt != nil {
		return io.NewSectionReader(r.ReaderAt, offset, size), nil
	}
	return nil, fmt.Errorf("underlying stream does not support ReaderAt")
}

func (wf *WzFile) CalcOffset(filePos uint32, hashedOffset uint32) uint32 {
	offset := (filePos - 0x3C) ^ 0xFFFFFFFF
	offset *= uint32(wf.Header.VersionDetector.GetHashVersion())
	offset -= 0x581C3F6D

	distance := offset & 0x1F
	offset = bits.RotateLeft32(offset, int(distance))

	offset ^= hashedOffset
	offset += 0x78
	return offset
}

// CalcChecksum calculates the checksum of the image
func (img *WzImage) CalcChecksum() (int, error) {
	stream := img.Stream
	buf := make([]byte, 4096)
	checksum := 0
	size := img.Size

	for size > 0 {
		toRead := len(buf)
		if size < toRead {
			toRead = size
		}

		read, err := stream.Read(buf[:toRead])
		if err != nil {
			return 0, err
		}

		for _, b := range buf[:read] {
			checksum += int(b)
		}

		size -= read
	}

	return checksum, nil
}

// TryExtract extracts the image data
func (img *WzImage) TryExtract() error {
	if img.Extracted {
		return nil
	}

	stream := img.Stream
	if !img.ChecksumChecked {
		calculatedChecksum, err := img.CalcChecksum()
		if err != nil {
			return err
		}
		if calculatedChecksum != img.Checksum {
			return errors.New("checksum mismatch")
		}
		img.ChecksumChecked = true
	}
	stream.Seek(0, io.SeekStart)

	reader := NewWzBinaryReader(stream)
	err := img.ExtractImg(reader, img.Node)
	if err != nil {
		return err
	}
	img.Extracted = true
	return nil
}

func (img *WzImage) ExtractImg(reader *WzBinaryReader, parent *WzNode) error {
	tag, err := reader.ReadImageObjectTypeName(img.WzFile.WzStructure.Encryption.Keys)
	if err != nil {
		return err
	}
	//var pp = reader.PartialStream()

	switch tag {
	case "Property":
		reader.SkipBytes(2)
		entries, err := reader.ReadCompressedInt32()
		if err != nil {
			return err
		}
		for i := 0; i < int(entries); i++ {
			err = img.ExtractValue(reader, parent)
			if err != nil {
				return err
			}
		}
		parent.Type = "Property"
	case "Shape2D#Vector2D":
		x, _ := reader.ReadCompressedInt32()
		y, _ := reader.ReadCompressedInt32()
		parent.Value = image.Pt(int(x), int(y))
		parent.Type = "Shape2D#Vector2D"
	case "Canvas":
		reader.SkipBytes(1)
		first, _ := reader.ReadByte()
		if first == 0x01 {
			reader.SkipBytes(2)
			entries, err := reader.ReadCompressedInt32()
			if err != nil {
				return err
			}
			for i := 0; i < int(entries); i++ {
				err = img.ExtractValue(reader, parent)
				if err != nil {
					return err
				}
			}
		}
		w, _ := reader.ReadCompressedInt32()
		h, _ := reader.ReadCompressedInt32()
		form, _ := reader.ReadCompressedInt32()
		b, _ := reader.ReadByte()
		form += int32(b)
		reader.SkipBytes(4)
		dataLen, _ := reader.ReadInt32()
		//ps := pp
		pos := reader.Pos()
		/*if ps != nil {
			pos += ps.Offset
		}*/
		wz_png := &WzPng{
			Width:      int(w),
			Height:     int(h),
			DataLength: int(dataLen),
			Form:       int(form),
			Offset:     uint32(pos),
			Image:      img,
		}
		parent.Value = wz_png
		parent.Type = "Canvas"
		reader.SkipBytes(int64(dataLen))
	case "Shape2D#Convex2D":
		entries, err := reader.ReadCompressedInt32()
		if err != nil {
			return err
		}
		var points = make([]image.Point, entries)
		var virtualNode = NewWzNode("")

		for i := 0; i < int(entries); i++ {
			img.ExtractImg(reader, virtualNode)
			if point, ok := virtualNode.Value.(image.Point); ok {
				points[i] = point
			}
		}
		parent.Value = points
		parent.Type = "Shape2D#Convex2D"

	case "UOL":
		// 跳过1字节（通常为0x00）
		reader.SkipBytes(1)
		// 读取UOL字符串
		uolStr, err := reader.ReadImageString(img.WzFile.WzStructure.Encryption.Keys)
		if err != nil {
			return err
		}
		parent.Value = NewWzUol(uolStr)
		parent.Type = "UOL"
	default:
		return fmt.Errorf("unsupported tag type: %s", tag)
	}
	return nil
}

// ExtractValue extracts a single value from the reader
func (img *WzImage) ExtractValue(reader *WzBinaryReader, parent *WzNode) error {
	key, err := reader.ReadImageString(img.WzFile.WzStructure.Encryption.Keys)
	if err != nil {
		return err
	}

	child := parent.AddChild(NewWzNode(key))
	// 打印parent和child的路径，便于调试树结构
	flag, err := reader.ReadByte()
	if err != nil {
		return err
	}

	switch flag {
	case 0x00:
		child.Value = nil
	case 0x02, 0x0B:
		val, err := reader.ReadInt16()
		if err != nil {
			return err
		}
		child.Value = val
	case 0x03, 0x13:
		val, err := reader.ReadCompressedInt32()
		if err != nil {
			return err
		}
		child.Value = val
	case 0x14:
		val, err := reader.ReadCompressedInt64()
		if err != nil {
			return err
		}
		child.Value = val
	case 0x04:
		val, err := reader.ReadCompressedSingle()
		if err != nil {
			return err
		}
		child.Value = val
	case 0x05:
		val, err := reader.ReadDouble()
		if err != nil {
			return err
		}
		child.Value = val
	case 0x08:
		val, err := reader.ReadImageString(img.WzFile.WzStructure.Encryption.Keys)
		if err != nil {
			return err
		}
		child.Value = val
	case 0x09:
		// 1. 读取结构块长度
		objDataLen, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("读取结构体长度失败: %v", err)
		}

		// 2. 计算结构体的结束位置
		curPos := reader.Pos()
		endPos := curPos + int64(objDataLen)

		// 3. 提取结构体内容
		if err := img.ExtractImg(reader, child); err != nil {
			return fmt.Errorf("提取结构体内容失败: %v", err)
		}

		// 4. 检查是否完全读完结构体
		newPos := reader.Pos()
		if newPos != endPos {
			return fmt.Errorf("结构体未完整读取 (%d != %d)", newPos, endPos)
		}
	default:
		return fmt.Errorf("unknown flag: %x", flag)
	}

	return nil
}

// OpenRead 返回一个指向当前 WzImage 数据的 io.ReadSeeker
func (img *WzImage) OpenRead() io.ReadSeeker {
	// 如果 Stream 已经存在且指向正确位置，直接返回
	if img.Stream != nil {
		img.Stream.Seek(0, io.SeekStart)
		return img.Stream
	}
	// 否则重新创建一个新的 PartialStream 或 bytes.Reader
	stream, err := NewPartialStream(img.WzFile.FileStream.File(), img.Offset, int64(img.Size))
	if err != nil {
		// fallback: 读入内存
		buf := make([]byte, img.Size)
		_, _ = img.WzFile.FileStream.BaseStream.Seek(img.Offset, io.SeekStart)
		_, _ = io.ReadFull(img.WzFile.FileStream.BaseStream, buf)
		return bytes.NewReader(buf)
	}
	img.Stream = stream
	return stream
}
