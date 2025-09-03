package wzlib

type WzDirectory struct {
	Name                 string  // 目录名称
	WzFile               *WzFile // 所属 WzFile 对象
	Size                 int     // 目录大小
	Checksum             int     // 校验值
	Offset               int
	HashedOffset         uint32 // 哈希偏移量
	HashedOffsetPosition uint32 // 哈希偏移位置
}

// NewWzDirectory 创建一个新的 WzDirectory 实例
func NewWzDirectory(name string, size int, checksum int, hashedOffset uint32, hashedOffsetPosition uint32, wzFile *WzFile) *WzDirectory {
	return &WzDirectory{
		Name:                 name,
		WzFile:               wzFile,
		Size:                 size,
		Checksum:             checksum,
		HashedOffset:         hashedOffset,
		HashedOffsetPosition: hashedOffsetPosition,
	}
}
