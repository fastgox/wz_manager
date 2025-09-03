package wzlib

import (
	"math"
	"strconv"
)

type WzCapabilities int

const (
	WzCapabilitiesEncverMissing WzCapabilities = 1 << iota
)

type IWzVersionDetector interface {
	GetWzVersion() int
	GetHashVersion() uint
	TryGetNextVersion() bool
}

type FixedVersion struct {
	WzVersion   int
	HashVersion uint
}

func (f *FixedVersion) GetWzVersion() int {
	return f.WzVersion
}

func (f *FixedVersion) GetHashVersion() uint {
	return f.HashVersion
}

func (f *FixedVersion) TryGetNextVersion() bool {
	return false
}

type OrdinalVersionDetector struct {
	EncryptedVersion int
	startVersion     int
	versionTest      []int
	hashVersionTest  []uint
}

// NewOrdinalVersionDetector creates a new detector with the encrypted version
func NewOrdinalVersionDetector(encryptedVersion int) *OrdinalVersionDetector {
	return &OrdinalVersionDetector{
		EncryptedVersion: encryptedVersion,
		startVersion:     -1,
		versionTest:      []int{},
		hashVersionTest:  []uint{},
	}
}

// GetWzVersion returns the last successful version
func (d *OrdinalVersionDetector) GetWzVersion() int {
	if len(d.versionTest) > 0 {
		return d.versionTest[len(d.versionTest)-1]
	}
	return 0
}

// GetHashVersion returns the last hash version
func (d *OrdinalVersionDetector) GetHashVersion() uint {
	if len(d.hashVersionTest) > 0 {
		return d.hashVersionTest[len(d.hashVersionTest)-1]
	}
	return 0
}

func CalcHashVersion(wzVersion int) int {
	sum := 0
	versionStr := strconv.Itoa(wzVersion)

	for i := 0; i < len(versionStr); i++ {
		sum <<= 5
		sum += int(versionStr[i]) + 1
	}

	return sum
}

// TryGetNextVersion tries next wz version until match encrypted version
func (d *OrdinalVersionDetector) TryGetNextVersion() bool {
	for i := d.startVersion + 1; i < math.MaxInt16; i++ {
		sum := CalcHashVersion(i)
		enc := 0xff ^
			((sum >> 24) & 0xFF) ^
			((sum >> 16) & 0xFF) ^
			((sum >> 8) & 0xFF) ^
			(sum & 0xFF)

		if enc == d.EncryptedVersion {
			d.versionTest = append(d.versionTest, i)
			d.hashVersionTest = append(d.hashVersionTest, uint(sum))
			d.startVersion = i
			return true
		}
	}
	return false
}

type WzHeader struct {
	Signature         string
	Copyright         string
	FileName          string
	HeaderSize        int32
	DataSize          int64
	FileSize          int64
	DataStartPosition int64
	DirEndPosition    int64
	VersionChecked    bool
	Capabilities      WzCapabilities
	VersionDetector   IWzVersionDetector
}

func NewWzHeader(signature, copyright, fileName string, headerSize int32, dataSize, fileSize, dataStartPosition int64) *WzHeader {
	return &WzHeader{
		Signature:         signature,
		Copyright:         copyright,
		FileName:          fileName,
		HeaderSize:        headerSize,
		DataSize:          dataSize,
		FileSize:          fileSize,
		DataStartPosition: dataStartPosition,
		VersionChecked:    false,
	}
}

// SetWzVersion sets the WZ version for the header
func (h *WzHeader) SetWzVersion(wzVersion int) {
	h.VersionDetector = &FixedVersion{
		WzVersion:   wzVersion,
		HashVersion: 0, // Placeholder hash version
	}
}

// SetOrdinalVersionDetector sets an ordinal version detector for encrypted versions
func (h *WzHeader) SetOrdinalVersionDetector(encryptedVersion int) {
	h.VersionDetector = NewOrdinalVersionDetector(encryptedVersion)
}

// HasCapabilities checks if the header has a specific capability
func (h *WzHeader) HasCapabilities(cap WzCapabilities) bool {
	return cap == (h.Capabilities & cap)
}
