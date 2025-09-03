package wzlib

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
)

type WzCryptoKeyType int

const (
	Unknown WzCryptoKeyType = iota
	BMS
	KMS
	GMS
)

var (
	BmsCryptoKey = NewWzCryptoKey([]byte{0, 0, 0, 0})
	KmsCryptoKey = NewWzCryptoKey([]byte{0xb9, 0x7d, 0x63, 0xe9})
	GmsCryptoKey = NewWzCryptoKey([]byte{0x4D, 0x23, 0xc7, 0x2b})
)

type WzCryptoKey struct {
	iv        []byte
	keys      []byte
	isEmptyIV bool
}

// NewWzCryptoKey creates a new WzCryptoKey instance
func NewWzCryptoKey(iv []byte) *WzCryptoKey {
	key := &WzCryptoKey{
		iv: iv,
	}
	if len(iv) == 0 || int32(iv[0])|int32(iv[1])|int32(iv[2])|int32(iv[3]) == 0 {
		key.isEmptyIV = true
	}
	return key
}

func (k *WzCryptoKey) GetKey(index int) (byte, error) {
	if k.isEmptyIV {
		return 0, nil
	}
	if k.keys == nil || len(k.keys) <= index {
		if err := k.EnsureKeySize(index + 1); err != nil {
			return 0, err
		}
	}
	return k.keys[index], nil
}

// EnsureKeySize ensures the keys array is large enough
func (k *WzCryptoKey) EnsureKeySize(size int) error {
	if k.isEmptyIV {
		return nil
	}
	if len(k.keys) >= size {
		return nil
	}

	size = ((size + 63) / 64) * 64 // Round up to the nearest multiple of 64
	startIndex := 0

	if k.keys == nil {
		k.keys = make([]byte, size)
	} else {
		startIndex = len(k.keys)
		newKeys := make([]byte, size)
		copy(newKeys, k.keys)
		k.keys = newKeys
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return err
	}

	blockSize := block.BlockSize()
	aesEncryptor := ECBEncrypter(block)

	for i := startIndex; i < size; i += blockSize {
		blockData := make([]byte, blockSize)
		if i == 0 {
			for j := 0; j < blockSize; j++ {
				blockData[j] = k.iv[j%len(k.iv)]
			}
		} else {
			copy(blockData, k.keys[i-blockSize:i])
		}
		aesEncryptor.CryptBlocks(k.keys[i:i+blockSize], blockData)
	}

	return nil
}

// ECBEncrypter wraps an AES cipher in ECB mode
func ECBEncrypter(block cipher.Block) cipher.BlockMode {
	return &ecbEncrypter{block: block}
}

type ecbEncrypter struct {
	block cipher.Block
}

func (x *ecbEncrypter) BlockSize() int {
	return x.block.BlockSize()
}

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.BlockSize() != 0 {
		panic("input not full blocks")
	}
	if len(dst) < len(src) {
		panic("output smaller than input")
	}

	for len(src) > 0 {
		x.block.Encrypt(dst[:x.BlockSize()], src[:x.BlockSize()])
		src = src[x.BlockSize():]
		dst = dst[x.BlockSize():]
	}
}

// Decrypt decrypts a buffer in-place
func (k *WzCryptoKey) Decrypt(buffer []byte, startIndex, length int) {
	if k.isEmptyIV {
		return
	}

	if err := k.EnsureKeySize(length); err != nil {
		log.Println(err)
		return
	}

	for i := 0; i < length; i++ {
		buffer[startIndex+i] ^= k.keys[i]
	}
}

// AES key used for encryption
var aesKey = []byte{
	0x13, 0x00, 0x00, 0x00,
	0x08, 0x00, 0x00, 0x00,
	0x06, 0x00, 0x00, 0x00,
	0xB4, 0x00, 0x00, 0x00,
	0x1B, 0x00, 0x00, 0x00,
	0x0F, 0x00, 0x00, 0x00,
	0x33, 0x00, 0x00, 0x00,
	0x52, 0x00, 0x00, 0x00,
}

type WzCrypto struct {
	Keys    *WzCryptoKey
	ListWZ  bool
	EncType WzCryptoKeyType
}

// NewWzCrypto 创建一个新的 WzCrypto 实例
func NewWzCrypto() *WzCrypto {
	return &WzCrypto{
		ListWZ:  false,
		EncType: Unknown,
	}
}

// Reset 重置加密状态
func (wc *WzCrypto) Reset() {
	wc.ListWZ = false
	wc.EncType = Unknown
}

// DetectEncryption 检测文件的加密类型
func (wc *WzCrypto) DetectEncryption(wzFile *WzFile) error {

	// 保存文件流的当前位置
	oldOff, err := wzFile.FileStream.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("failed to get current file position: %v", err)
	}

	// 定位到加密数据的开始位置
	_, err = wzFile.FileStream.Seek(wzFile.Header.DataStartPosition, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to data start position: %v", err)
	}

	// 创建一个二进制读取器
	reader := NewWzBinaryReader(wzFile.FileStream)

	// 读取加密数据长度
	dataLen, err := reader.ReadCompressedInt32()
	if err != nil || dataLen <= 0 {
		// 如果数据长度无效，返回错误
		return fmt.Errorf("invalid or missing compressed data length: %v", err)
	}

	// 跳过一个字节
	_, err = wzFile.FileStream.Seek(1, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("failed to skip byte in file stream: %v", err)
	}
	llen, _ := reader.ReadSByte()
	llen = -llen
	// 读取加密数据
	bytes := make([]byte, int32(llen))
	_, err = io.ReadFull(wzFile.FileStream, bytes)
	if err != nil {
		return fmt.Errorf("failed to read encrypted data bytes: %v", err)
	}
	for i := int8(0); i < llen; i++ {
		bytes[i] ^= 0xAA + byte(i)
	}

	// 尝试每种加密键
	for _, keySet := range []struct {
		encType WzCryptoKeyType
		keys    *WzCryptoKey
	}{
		{BMS, BmsCryptoKey},
		{KMS, KmsCryptoKey},
		{GMS, GmsCryptoKey},
	} {
		// 解密并转换为字符串
		decrypted := make([]byte, 0)
		for i := 0; i < len(bytes); i++ {
			a, err := keySet.keys.GetKey(i)
			if err != nil {
				log.Println(err)
				a = 0
			}
			decrypted = append(decrypted, bytes[i]^a)
		}

		// 判断是否为合法的节点名称
		if wc.IsLegalNodeName(string(decrypted)) {
			wc.EncType = keySet.encType
			wc.Keys = keySet.keys
			break
		}
	}

	// 恢复文件流的原始位置
	_, err = wzFile.FileStream.Seek(oldOff, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to restore file position: %v", err)
	}

	return nil
}

func (wc *WzCrypto) IsLegalNodeName(nodeName string) bool {
	// MSEA 225 has a node named "Base,Character,Effect,..."; wzlib must handle it
	if strings.HasSuffix(nodeName, ".img") || strings.HasSuffix(nodeName, ".lua") {
		return true
	}
	matched, err := regexp.MatchString(`^[A-Za-z0-9_,]+$`, nodeName)
	if err != nil {
		return false
	}
	return matched
}

// Decrypt 解密数据
func (wc *WzCrypto) Decrypt(data []byte) []byte {
	var key *WzCryptoKey
	switch wc.EncType {
	case BMS:
		key = BmsCryptoKey
	case KMS:
		key = KmsCryptoKey
	case GMS:
		key = GmsCryptoKey
	default:
		return data // 如果未知加密类型，返回原始数据
	}

	// 示例解密逻辑：简单异或
	for i := 0; i < len(data); i++ {
		data[i] ^= key.iv[i%len(key.iv)]
	}
	return data
}
