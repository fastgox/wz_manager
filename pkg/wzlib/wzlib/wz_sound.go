package wzlib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sync"
)

// WzSoundType 表示音频类型
type WzSoundType int

const (
	WzSoundTypeUnknown WzSoundType = iota
	WzSoundTypeMp3
	WzSoundTypePcm
	WzSoundTypeBinary
)

// WzSound 表示WZ音频对象
type WzSound struct {
	Offset     uint32
	DataLength int
	Ms         int
	MediaType  *AMMediaType
	WzImage    *WzImage
}

// AMMediaType、WAVEFORMATEX、MPEGLAYER3WAVEFORMAT 结构体需根据实际情况定义
type AMMediaType struct {
	MajorType string
	SubType   string
	PbFormat  any
}

type WAVEFORMATEX struct {
	FormatTag      uint16
	Channels       uint16
	SamplesPerSec  uint32
	AvgBytesPerSec uint32
	BlockAlign     uint16
	BitsPerSample  uint16
}

type MPEGLAYER3WAVEFORMAT struct {
	Wfx WAVEFORMATEX
}

// 声音类型判断
func (ws *WzSound) SoundType() WzSoundType {
	if ws.MediaType == nil {
		return WzSoundTypeUnknown
	}
	if ws.MediaType.MajorType == "Stream" {
		if ws.MediaType.SubType == "MPEG1Audio" {
			return WzSoundTypeMp3
		} else if ws.MediaType.SubType == "WAVE" {
			switch fmt := ws.MediaType.PbFormat.(type) {
			case *MPEGLAYER3WAVEFORMAT:
				return WzSoundTypeMp3
			case *WAVEFORMATEX:
				if fmt.FormatTag == 1 { // PCM
					if ws.Ms == 1000 && int(fmt.SamplesPerSec) == ws.DataLength {
						return WzSoundTypeBinary
					}
					return WzSoundTypePcm
				}
			}
		}
	}
	return WzSoundTypeUnknown
}

// Channels 获取声道数
func (ws *WzSound) Channels() int {
	switch fmt := ws.MediaType.PbFormat.(type) {
	case *WAVEFORMATEX:
		return int(fmt.Channels)
	case *MPEGLAYER3WAVEFORMAT:
		return int(fmt.Wfx.Channels)
	default:
		return 0
	}
}

// Frequency 获取采样率
func (ws *WzSound) Frequency() int {
	switch fmt := ws.MediaType.PbFormat.(type) {
	case *WAVEFORMATEX:
		return int(fmt.SamplesPerSec)
	case *MPEGLAYER3WAVEFORMAT:
		return int(fmt.Wfx.SamplesPerSec)
	default:
		return 0
	}
}

// ExtractSound 提取音频数据
func (ws *WzSound) ExtractSound() ([]byte, error) {
	switch ws.SoundType() {
	case WzSoundTypeMp3:
		data := make([]byte, ws.DataLength)
		if err := ws.CopyTo(data, 0); err != nil {
			return nil, err
		}
		return data, nil
	case WzSoundTypePcm:
		waveFmtEx, ok := ws.MediaType.PbFormat.(*WAVEFORMATEX)
		if !ok {
			return nil, errors.New("invalid PCM format")
		}
		data := make([]byte, ws.DataLength+44)
		buf := bytes.NewBuffer(data[:0])
		// 写WAV头
		buf.Write([]byte("RIFF"))
		binary.Write(buf, binary.LittleEndian, uint32(ws.DataLength+36))
		buf.Write([]byte("WAVE"))
		buf.Write([]byte("fmt "))
		binary.Write(buf, binary.LittleEndian, uint32(16))
		binary.Write(buf, binary.LittleEndian, waveFmtEx.FormatTag)
		binary.Write(buf, binary.LittleEndian, waveFmtEx.Channels)
		binary.Write(buf, binary.LittleEndian, waveFmtEx.SamplesPerSec)
		binary.Write(buf, binary.LittleEndian, waveFmtEx.AvgBytesPerSec)
		binary.Write(buf, binary.LittleEndian, waveFmtEx.BlockAlign)
		binary.Write(buf, binary.LittleEndian, waveFmtEx.BitsPerSample)
		buf.Write([]byte("data"))
		binary.Write(buf, binary.LittleEndian, uint32(ws.DataLength))
		// 写PCM数据
		if err := ws.CopyTo(data[44:], 0); err != nil {
			return nil, err
		}
		return data, nil
	}
	return nil, nil
}

// CopyTo 复制音频数据到buffer
func (ws *WzSound) CopyTo(buffer []byte, offset int) error {
	if len(buffer)-offset < ws.DataLength {
		return errors.New("insufficient buffer size")
	}
	// 假设WzImage有OpenRead()方法返回io.ReadSeeker
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	s := ws.WzImage.OpenRead()
	s.Seek(int64(ws.Offset), io.SeekStart)
	_, err := io.ReadFull(s, buffer[offset:offset+ws.DataLength])
	return err
}
