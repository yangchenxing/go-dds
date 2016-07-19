package file

import "errors"

var (
	ErrMissBlock = errors.New("miss this block")
)

type File interface {
	FileSize() int
	BlockSize() int
	HasBlock(id uint) bool
	ReadBlock(id uint) ([]byte, error)
	WriteBlock(id uint, content []byte) error
	Close() error
}

type FileInfo struct {
	CheckSum  string `json:"checksum"`
	Timestamp int64  `json:"timestamp"`
	Size      int    `json:"size"`
	BlockSize int    `json:"block_size"`
}
