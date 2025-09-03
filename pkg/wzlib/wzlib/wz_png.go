package wzlib

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"math"
)

type WzPng struct {
	Width      int
	Height     int
	DataLength int
	Form       int
	Offset     uint32
	Image      *WzImage
}

func (p *WzPng) GetRawData() ([]byte, error) {
	stream := p.Image.OpenRead()
	endPos := int64(p.Offset) + int64(p.DataLength)

	_, err := stream.Seek(int64(p.Offset)+1, io.SeekStart) // 跳过第一个字节
	if err != nil {
		return nil, fmt.Errorf("seek failed: %v", err)
	}

	// 检查是否为 zlib 压缩数据 (0x78 0x9C)
	reader := NewWzBinaryReader(stream)
	header, err := reader.ReadUInt16()
	if err != nil {
		return nil, err
	}
	var zlibStream io.ReadCloser
	if int(header) == 0x9C78 {
		// 是标准 zlib 数据流
		dataLen := endPos - reader.Pos()
		_, _ = stream.Seek(int64(p.Offset)+1, io.SeekStart)
		limitReader := io.LimitReader(stream, dataLen)
		payload, err := io.ReadAll(limitReader)
		if err != nil {
			return nil, fmt.Errorf("create zlib reader: %v", err)
		}
		zlibStream, err = zlib.NewReader(bytes.NewBuffer(payload))
		if err != nil {
			return nil, fmt.Errorf("create zlib reader: %v", err)
		}
	} else {
		// 不是 zlib，说明是分块加密
		_, _ = stream.Seek(int64(p.Offset), io.SeekStart)

		buffer := make([]byte, p.DataLength)
		total := 0
		b := bufio.NewReader(stream)

		for pos := int64(p.Offset); pos < endPos; {
			blockSizeBytes := make([]byte, 4)
			if _, err := io.ReadFull(b, blockSizeBytes); err != nil {
				return nil, err
			}
			blockSize := int(binary.LittleEndian.Uint32(blockSizeBytes))

			if pos+4+int64(blockSize) > endPos {
				return nil, fmt.Errorf("block exceeds declared data size")
			}
			n, err := io.ReadFull(b, buffer[total:total+blockSize])
			if err != nil {
				return nil, err
			}

			// 解密块
			p.Image.WzFile.WzStructure.Encryption.Keys.Decrypt(buffer, total, blockSize)
			total += n
			pos += int64(n + 4)
		}

		dataStream := bytes.NewReader(buffer[:total])
		_, _ = dataStream.Seek(2, io.SeekStart)
		zlibStream, err = zlib.NewReader(dataStream)
		if err != nil {
			return nil, fmt.Errorf("zlib decode after decryption failed: %v", err)
		}
	}

	defer zlibStream.Close()

	var rawLen int
	switch p.Form {
	case 1, 257, 513:
		rawLen = p.Width * p.Height * 2
	case 2:
		rawLen = p.Width * p.Height * 4
	case 3:
		rawLen = int(math.Ceil(float64(p.Width)/4)) * 4 * int(math.Ceil(float64(p.Height)/4)) * 4 / 8
	case 517:
		rawLen = p.Width * p.Height / 128
	case 1026, 2050:
		rawLen = p.Width * p.Height
	default:
		// 未知格式，直接读取全部解压后数据
		return io.ReadAll(zlibStream)
	}

	output := make([]byte, rawLen)
	if _, err := io.ReadFull(zlibStream, output); err != nil {
		return nil, fmt.Errorf("read decompressed data: %v", err)
	}
	return output, nil
}

func (p *WzPng) ExtractImage() (image.Image, error) {
	raw, err := p.GetRawData()
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}

	img := image.NewNRGBA(image.Rect(0, 0, p.Width, p.Height))
	img.Stride = 4 * p.Width
	var pixel []byte

	switch p.Form {
	case 1: // ARGB4444
		pixel = GetPixelDataBGRA4444(raw, p.Width, p.Height)
	case 2: // ARGB8888 (no conversion needed)
		pixel = BGRAtoRGBA(raw)

	case 3: // 黑白缩略图（暂略，待实现 GetPixelDataForm3）
		pixel = GetPixelDataForm3(raw, p.Width, p.Height)

	case 257: // ARGB1555
		pixel = GetPixelDataARGB1555(raw, p.Width, p.Height)

	case 513: // RGB565
		pixel = ConvertRGB565ToRGBA(raw, p.Width, p.Height)

	case 517: // RGB565 缩略图
		pixel = GetPixelDataForm517(raw, p.Width, p.Height)

	case 1026: // DXT3
		pixel = GetPixelDataDXT3(raw, p.Width, p.Height)

	case 2050: // DXT5
		pixel = GetPixelDataDXT5(raw, p.Width, p.Height)

	default:
		return nil, fmt.Errorf("unsupported image form: %d", p.Form)
	}

	if len(pixel) != p.Width*p.Height*4 {
		return nil, fmt.Errorf("pixel data size mismatch: got %d, expected %d", len(pixel), p.Width*p.Height*4)
	}

	// 写入 RGBA 图像
	copy(img.Pix, pixel)
	return img, nil
}

func GetPixelDataForm517(rawData []byte, width, height int) []byte {
	pixel := make([]byte, width*height*2) // 每像素 2 字节（RGB565）

	lineIndex := 0
	for j0 := 0; j0 < height/16; j0++ {
		dstIndex := lineIndex
		for i0 := 0; i0 < width/16; i0++ {
			idx := (i0 + j0*(width/16)) * 2
			b0 := rawData[idx]
			b1 := rawData[idx+1]

			for k := 0; k < 16; k++ {
				pixel[dstIndex] = b0
				pixel[dstIndex+1] = b1
				dstIndex += 2
			}
		}

		// 复制剩余 15 行（垂直方向）
		for k := 1; k < 16; k++ {
			copy(pixel[dstIndex:dstIndex+width*2], pixel[lineIndex:lineIndex+width*2])
			dstIndex += width * 2
		}

		lineIndex += width * 32 // 16 行 × 每行 width*2 字节
	}

	return pixel
}

// ConvertRGB565ToRGBA 将 RGB565 像素转换为 RGBA 图像
func ConvertRGB565ToRGBA(rgb565 []byte, width, height int) []byte {
	out := make([]byte, width*height*4)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := y*width*2 + x*2
			val := binary.LittleEndian.Uint16(rgb565[i:])

			rgba := RGB565ToRGBA(val)

			j := (y*width + x) * 4
			out[j+0] = rgba[2] // B
			out[j+1] = rgba[1] // G
			out[j+2] = rgba[0] // R
			out[j+3] = rgba[3] // A
		}
	}

	return out
}

func GetPixelDataForm3(rawData []byte, width, height int) []byte {
	pixel := make([]byte, width*height*4)
	buffer := make([]uint32, width*height)

	blockW := int(math.Ceil(float64(width) / 4.0))
	blockH := int(math.Ceil(float64(height) / 4.0))

	for y := 0; y < blockH; y++ {
		for x := 0; x < blockW; x++ {
			index := (x + y*blockW) * 2
			index2 := x*4 + y*width*4/4 // 目标 uint32 索引

			b0 := rawData[index]
			b1 := rawData[index+1]

			p := uint32((b0&0x0F)|((b0&0x0F)<<4)) |
				uint32((b0&0xF0)|((b0&0xF0)>>4))<<8 |
				uint32((b1&0x0F)|((b1&0x0F)<<4))<<16 |
				uint32((b1&0xF0)|((b1&0xF0)>>4))<<24

			for i := 0; i < 4; i++ {
				px := x*4 + i
				if px >= width {
					break
				}
				buffer[index2+i] = p
			}
		}

		// 复制剩余 3 行
		src := y * width
		for j := 1; j < 4; j++ {
			dstY := y*4 + j
			if dstY >= height {
				break
			}
			dst := dstY * width
			copy(buffer[dst:dst+width], buffer[src:src+width])
		}
	}

	// 将 buffer 复制回 pixel
	for i := 0; i < len(buffer); i++ {
		offset := i * 4
		binary.LittleEndian.PutUint32(pixel[offset:], buffer[i])
	}
	return pixel
}

func GetPixelDataBGRA4444(rawData []byte, width, height int) []byte {
	pixelBuffer := make([]byte, width*height*4)
	for i := 0; i < width*height; i++ {
		lo := rawData[i*2]
		hi := rawData[i*2+1]
		pixelBuffer[i*4+0] = (hi&0x0F)<<4 | (hi & 0x0F) // R
		pixelBuffer[i*4+1] = (lo & 0xF0) | (lo&0xF0)>>4 // G
		pixelBuffer[i*4+2] = (lo&0x0F)<<4 | (lo & 0x0F) // B
		pixelBuffer[i*4+3] = (hi & 0xF0) | (hi&0xF0)>>4 // A
	}
	return pixelBuffer
}

func ExpandAlphaTableDXT3(alpha []byte, raw []byte, offset int) {
	for i := 0; i < 16; i += 2 {
		b := raw[offset]
		alpha[i] = b & 0x0F
		alpha[i+1] = (b & 0xF0) >> 4
		offset++
	}
	for i := 0; i < 16; i++ {
		alpha[i] |= alpha[i] << 4 // 扩展为 8bit：x → x | (x << 4)
	}
}

// ExpandAlphaTableDXT5 生成 DXT5 的 8 alpha 值
func ExpandAlphaTableDXT5(alpha []byte, a0, a1 byte) {
	alpha[0] = a0
	alpha[1] = a1

	if a0 > a1 {
		for i := 2; i < 8; i++ {
			alpha[i] = byte(((8-int(i))*int(a0) + (int(i)-1)*int(a1) + 3) / 7)
		}
	} else {
		for i := 2; i < 6; i++ {
			alpha[i] = byte(((6-int(i))*int(a0) + (int(i)-1)*int(a1) + 2) / 5)
		}
		alpha[6] = 0
		alpha[7] = 255
	}
}

// ExpandAlphaIndexTableDXT5 解码 16 个 3-bit alpha 索引
func ExpandAlphaIndexTableDXT5(alphaIndex []int, raw []byte, offset int) {
	for i := 0; i < 16; i += 8 {
		flags := int(raw[offset]) |
			(int(raw[offset+1]) << 8) |
			(int(raw[offset+2]) << 16)
		offset += 3

		for j := 0; j < 8; j++ {
			mask := 0x07 << (3 * j)
			alphaIndex[i+j] = (flags & mask) >> (3 * j)
		}
	}
}

func GetPixelDataDXT3(raw []byte, width, height int) []byte {
	pixel := make([]byte, width*height*4)
	var colorTable [4][4]byte // RGBA
	var colorIdx [16]int
	var alphaTable [16]byte

	for y := 0; y < height; y += 4 {
		for x := 0; x < width; x += 4 {
			offset := x*4 + y*width

			ExpandAlphaTableDXT3(alphaTable[:], raw, offset)

			c0 := binary.LittleEndian.Uint16(raw[offset+8:])
			c1 := binary.LittleEndian.Uint16(raw[offset+10:])
			ExpandColorTable(&colorTable, c0, c1)

			ExpandColorIndexTable(colorIdx[:], raw, offset+12)

			for j := 0; j < 4; j++ {
				for i := 0; i < 4; i++ {
					idx := j*4 + i
					SetPixel(
						pixel,
						x+i, y+j, width,
						colorTable[colorIdx[idx]],
						alphaTable[idx],
					)
				}
			}
		}
	}
	return pixel
}

func GetPixelDataDXT5(raw []byte, width, height int) []byte {
	pixel := make([]byte, width*height*4)
	var colorTable [4][4]byte
	var colorIdx [16]int
	var alphaTable [8]byte
	var alphaIdx [16]int

	for y := 0; y < height; y += 4 {
		for x := 0; x < width; x += 4 {
			offset := x*4 + y*width

			a0 := raw[offset]
			a1 := raw[offset+1]
			ExpandAlphaTableDXT5(alphaTable[:], a0, a1)
			ExpandAlphaIndexTableDXT5(alphaIdx[:], raw, offset+2)

			c0 := binary.LittleEndian.Uint16(raw[offset+8:])
			c1 := binary.LittleEndian.Uint16(raw[offset+10:])
			ExpandColorTable(&colorTable, c0, c1)

			ExpandColorIndexTable(colorIdx[:], raw, offset+12)

			for j := 0; j < 4; j++ {
				for i := 0; i < 4; i++ {
					idx := j*4 + i
					SetPixel(
						pixel,
						x+i, y+j, width,
						colorTable[colorIdx[idx]],
						alphaTable[alphaIdx[idx]],
					)
				}
			}
		}
	}
	return pixel
}

func ExpandColorIndexTable(colorIndex []int, rawData []byte, offset int) {
	for i := 0; i < 16; i += 4 {
		b := rawData[offset]
		colorIndex[i+0] = int(b & 0x03)        // 00
		colorIndex[i+1] = int((b >> 2) & 0x03) // 01
		colorIndex[i+2] = int((b >> 4) & 0x03) // 10
		colorIndex[i+3] = int((b >> 6) & 0x03) // 11
		offset++
	}
}

func SetPixel(pixelData []byte, x, y, width int, color [4]byte, alpha byte) {
	offset := (y*width + x) * 4
	pixelData[offset+0] = color[2] // B
	pixelData[offset+1] = color[1] // G
	pixelData[offset+2] = color[0] // R
	pixelData[offset+3] = alpha    // A
}

// ExpandColorTable 解码 2 个 RGB565 基础颜色，并生成 4 色调色板（RGBA）
func ExpandColorTable(out *[4][4]byte, c0, c1 uint16) {
	out[0] = RGB565ToRGBA(c0)
	out[1] = RGB565ToRGBA(c1)

	if c0 > c1 {
		for i := 0; i < 3; i++ { // RGB分量
			out[2][i] = byte((2*int(out[0][i]) + int(out[1][i]) + 1) / 3)
			out[3][i] = byte((int(out[0][i]) + 2*int(out[1][i]) + 1) / 3)
		}
	} else {
		for i := 0; i < 3; i++ {
			out[2][i] = byte((int(out[0][i]) + int(out[1][i])) / 2)
			out[3][i] = 0 // Color.Black 模拟为 R,G,B = 0
		}
	}

	// 所有颜色的 Alpha 设为 255
	for i := 0; i < 4; i++ {
		out[i][3] = 255
	}
}

// RGB565ToRGBA 解码 RGB565 编码为 RGBA8888 (带精度补偿)
func RGB565ToRGBA(val uint16) [4]byte {
	r5 := (val >> 11) & 0x1F
	g6 := (val >> 5) & 0x3F
	b5 := val & 0x1F

	r := byte((r5 << 3) | (r5 >> 2)) // 5bit → 8bit（高精度）
	g := byte((g6 << 2) | (g6 >> 4)) // 6bit → 8bit
	b := byte((b5 << 3) | (b5 >> 2)) // 5bit → 8bit

	return [4]byte{r, g, b, 255} // RGBA 固定 alpha=255
}

func ARGB1555ToRGBA(val uint16) [4]byte {
	a := byte(0xFF)
	if (val & 0x8000) == 0 {
		a = 0
	}
	r := byte((val >> 10) & 0x1F)
	g := byte((val >> 5) & 0x1F)
	b := byte(val & 0x1F)
	// 5bit转8bit
	r = (r << 3) | (r >> 2)
	g = (g << 3) | (g >> 2)
	b = (b << 3) | (b >> 2)
	return [4]byte{r, g, b, a}
}

func BGRAtoRGBA(bgra []byte) []byte {
	rgba := make([]byte, len(bgra))
	for i := 0; i+3 < len(bgra); i += 4 {
		rgba[i+0] = bgra[i+2] // R
		rgba[i+1] = bgra[i+1] // G
		rgba[i+2] = bgra[i+0] // B
		rgba[i+3] = bgra[i+3] // A
	}
	return rgba
}

func GetPixelDataARGB1555(raw []byte, width, height int) []byte {
	out := make([]byte, width*height*4)
	for i := 0; i < width*height; i++ {
		val := binary.LittleEndian.Uint16(raw[i*2:])
		rgba := ARGB1555ToRGBA(val)
		copy(out[i*4:], rgba[:])
	}
	return out
}
