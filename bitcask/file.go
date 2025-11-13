package bitcask

import (
	"fmt"
	"hash/crc32"
	"os"
	"time"
)

type File struct {
	FileID     uint32
	Path       string
	CurrentPos uint32
	FileSize   uint32
	Fd         *os.File
	//FileLock   *FileLock
}

func NewFile(fileID uint32, Path string) *File {
	return &File{
		FileID:     fileID,
		Path:       Path,
		CurrentPos: 0,
	}
}

func (f *File) OpenFile() error {
	fd, err := os.OpenFile(fmt.Sprintf("%s%d.data", f.Path, f.FileID), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	f.Fd = fd
	size, err := f.Size()
	if err != nil {
		return err
	}
	f.CurrentPos = uint32(0)
	f.FileSize = uint32(size)
	return nil
}

func (f *File) CloseFile() error {
	err := f.Fd.Close()
	f.Fd = nil
	return err
}

func (f *File) Write(data []byte) (int, error) {
	return f.Fd.Write(data)
}
func (f *File) Read(offset uint32, size uint32) ([]byte, error) {
	buf := make([]byte, size)
	_, err := f.Fd.ReadAt(buf, int64(offset))
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	return f.Fd.Seek(offset, whence)
}
func (f *File) ReadAt(buf []byte, offset int64) (int, error) {
	return f.Fd.ReadAt(buf, offset)
}
func (f *File) ReadEntry() (*Entry, error) {
	oldPos := f.CurrentPos
	buf, err := f.Read(f.CurrentPos, 20)
	if err != nil {
		return nil, err
	}
	f.CurrentPos += 20
	header := DecodeHeader(buf)
	key, err := f.Read(f.CurrentPos, header.KeySize)
	if err != nil {
		return nil, err
	}
	// Entry set
	entry := NewEntry(key, f.FileID, header.ValueSize, header.ValuePos, header.TimeStamp)
	f.CurrentPos += header.KeySize
	value, err := f.Read(f.CurrentPos, header.ValueSize)
	h := crc32.NewIEEE()
	h.Write(key)
	h.Write(value)
	if header.Crc != h.Sum32() {
		// Truncate file from oldPos to filePos
		f.Truncate(int64(oldPos))
		return nil, fmt.Errorf("checksum error")
	}
	if err != nil {
		return nil, err
	}
	// checksum

	f.CurrentPos += header.ValueSize
	return entry, nil
}
func (f *File) Sync() error {
	return f.Fd.Sync()
}
func (f *File) Stat() (os.FileInfo, error) {
	return f.Fd.Stat()
}
func (f *File) Truncate(size int64) error {
	return f.Fd.Truncate(size)
}
func (f *File) Size() (int64, error) {
	stat, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}
func (f *File) Delete() error {
	return os.Remove(fmt.Sprintf("%s%d.data", f.Path, f.FileID))
}
func (f *File) Rename(newPath string) error {
	return os.Rename(fmt.Sprintf("%s%d.data", f.Path, f.FileID), newPath)
}
func (f *File) WriteRecord(key, value []byte) (*Record, error) {
	rec := NewRecord(uint32(time.Now().Unix()), key, f.CurrentPos+20+uint32(len(key)), value)
	nums, err := f.Write(rec.Encode())
	if err != nil {
		return nil, err
	}
	f.CurrentPos += uint32(nums)
	return rec, nil
}
