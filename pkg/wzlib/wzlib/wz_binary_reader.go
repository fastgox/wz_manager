package wzlib

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type Decrypter interface {
	Decrypt(data []byte, offset int, length int)
}

type WzBinaryReader struct {
	BaseStream io.ReadSeeker
	ReaderAt   io.ReaderAt // 可选，只有底层支持才设置
	file       *os.File
}

func (r *WzBinaryReader) File() *os.File {
	return r.file
}

func NewWzBinaryReader(stream io.ReadSeeker) *WzBinaryReader {
	var ra io.ReaderAt
	if r, ok := stream.(io.ReaderAt); ok {
		ra = r
	}

	wzb := &WzBinaryReader{
		BaseStream: stream,
		ReaderAt:   ra,
	}
	if file, ok := stream.(*os.File); ok {
		wzb.file = file
	}
	return wzb
}

func (r *WzBinaryReader) PartialStream() *PartialStream {
	if r.BaseStream != nil {
		ps, ok := r.BaseStream.(*PartialStream)
		if ok {
			return ps
		}
	}
	return nil
}

func (r *WzBinaryReader) ReadCompressedSingle() (float32, error) {
	var fl int8
	err := binary.Read(r.BaseStream, binary.LittleEndian, &fl)
	if err != nil {
		return 0, err
	}

	if fl == -128 {
		var val float32
		if err := binary.Read(r.BaseStream, binary.LittleEndian, &val); err != nil {
			return 0, err
		}
		return val, nil
	}

	return float32(fl), nil
}

func (r *WzBinaryReader) ReadDouble() (float64, error) {
	var val float64
	err := binary.Read(r.BaseStream, binary.LittleEndian, &val)
	return val, err
}

func (r *WzBinaryReader) ReadByte() (byte, error) {

	var b [1]byte
	_, err := r.BaseStream.Read(b[:])
	return b[0], err
}
func (r *WzBinaryReader) ReadSByte() (int8, error) {
	var val int8
	err := binary.Read(r.BaseStream, binary.LittleEndian, &val)
	return val, err
}

func (r *WzBinaryReader) ReadInt16() (int16, error) {

	var val int16
	err := binary.Read(r.BaseStream, binary.LittleEndian, &val)
	return val, err
}
func (r *WzBinaryReader) ReadUInt16() (uint16, error) {

	var val uint16
	err := binary.Read(r.BaseStream, binary.LittleEndian, &val)
	return val, err
}

func (r *WzBinaryReader) ReadInt32() (int32, error) {

	var val int32
	err := binary.Read(r.BaseStream, binary.LittleEndian, &val)
	return val, err
}

func (r *WzBinaryReader) ReadUInt32() (uint32, error) {

	var val uint32
	err := binary.Read(r.BaseStream, binary.LittleEndian, &val)
	return val, err
}

func (r *WzBinaryReader) ReadInt64() (int64, error) {

	var val int64
	err := binary.Read(r.BaseStream, binary.LittleEndian, &val)
	return val, err
}

func (r *WzBinaryReader) ReadCompressedInt32() (int32, error) {

	var b [1]byte
	_, err := r.BaseStream.Read(b[:])
	if err != nil {
		return 0, err
	}

	if b[0] == 0x80 {
		var val int32
		err = binary.Read(r.BaseStream, binary.LittleEndian, &val)
		return val, err
	}

	return int32(int8(b[0])), nil
}

func (r *WzBinaryReader) ReadCompressedInt64() (int64, error) {

	var b [1]byte
	_, err := r.BaseStream.Read(b[:])
	if err != nil {
		return 0, err
	}

	if b[0] == 0x80 {
		var val int64
		err = binary.Read(r.BaseStream, binary.LittleEndian, &val)
		return val, err
	}

	return int64(int8(b[0])), nil
}
func (r *WzBinaryReader) ReadString(decrypter Decrypter) (string, error) {

	/*currentPos, err := r.BaseStream.Seek(0, io.SeekCurrent)
	if err != nil {
		return "", err
	}*/

	var size int8
	var err error
	size, _ = r.ReadSByte()
	if size < 0 { // ASCII/cp1252 字符串
		var usize int
		if size == -128 {
			size32, err := r.ReadInt32()
			if err != nil {
				return "", err
			}
			usize = int(size32)
		} else {
			usize = -int(size)
		}

		buffer := make([]byte, usize)
		_, err = r.BaseStream.Read(buffer)
		if err != nil {
			return "", err
		}

		decrypter.Decrypt(buffer, 0, usize)
		mask := byte(0xAA)
		for i := range buffer {
			buffer[i] ^= mask
			mask++
		}

		return string(buffer), nil
	} else if size > 0 { // UTF-16LE 字符串
		buffer := make([]byte, int(size)*2)
		_, err = r.BaseStream.Read(buffer)
		if err != nil {
			return "", err
		}

		decrypter.Decrypt(buffer, 0, int(size)*2)
		runes := make([]rune, size)
		for i := 0; i < int(size); i++ {
			runes[i] = rune(buffer[i*2]) | rune(buffer[i*2+1])<<8
			runes[i] ^= 0xAAAA
		}

		return string(runes), nil
	}

	return "", nil
}

func (r *WzBinaryReader) SkipBytes(count int64) error {

	_, err := r.BaseStream.Seek(count, io.SeekCurrent)
	return err
}

func (r *WzBinaryReader) ReadStringAt(offset int64, decrypter Decrypter) (string, error) {

	currentPos, err := r.BaseStream.Seek(0, io.SeekCurrent)
	if err != nil {
		return "", err
	}

	_, err = r.BaseStream.Seek(offset, io.SeekStart)
	if err != nil {
		return "", err
	}

	result, err := r.ReadString(decrypter)
	if err != nil {
		return "", err
	}

	_, err = r.BaseStream.Seek(currentPos, io.SeekStart)
	return result, err
}

func (r *WzBinaryReader) ReadImageObjectTypeName(decrypter Decrypter) (string, error) {

	flag, err := r.ReadByte()
	if err != nil {
		return "", err
	}

	switch flag {
	case 0x73:
		return r.ReadString(decrypter)
	case 0x1B:
		offset, err := r.ReadInt32()
		if err != nil {
			return "", err
		}
		return r.ReadStringAt(int64(offset), decrypter)
	default:
		return "", fmt.Errorf("unexpected flag '%x' when reading string", flag)
	}
}

func (r *WzBinaryReader) ReadImageString(decrypter Decrypter) (string, error) {

	flag, err := r.ReadByte()
	if err != nil {
		return "", err
	}

	switch flag {
	case 0x00:
		return r.ReadString(decrypter)
	case 0x01:
		offset, err := r.ReadInt32()
		if err != nil {
			return "", err
		}
		return r.ReadStringAt(int64(offset), decrypter)
	case 0x04:
		err := r.SkipBytes(8)
		if err != nil {
			return "", err
		}
		return "", nil
	default:
		return "", fmt.Errorf("unexpected flag '%x' when reading string", flag)
	}
}

func (r *WzBinaryReader) Seek(offset int64, i int) (int64, error) {
	return r.BaseStream.Seek(offset, i)
}

func (r *WzBinaryReader) Read(p []byte) (n int, err error) {
	return r.BaseStream.Read(p)
}

func (r *WzBinaryReader) Pos() int64 {
	now, _ := r.BaseStream.Seek(0, io.SeekCurrent)
	return now
}
