package bitcask

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	MaxFileSize = 1024 * 1024 * 1024 // 1GB
	RecordSize  = 20
)

type Bitcask struct {
	Path          string
	FileIDs       []uint32
	Files         []*File
	currentFileID uint32
	CurrentFile   *File
	memDB         *SkipListArr
}

func ScanDir(path string) ([]uint32, error) {
	var fileIDs []uint32

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a file and has the .data extension
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".data") {
			// Extract file ID from filename (e.g., 123.data -> 123)
			filename := strings.TrimSuffix(info.Name(), ".data")
			if fileID, err := strconv.ParseUint(filename, 10, 32); err == nil {
				fileIDs = append(fileIDs, uint32(fileID))
			}
		}
		return nil
	})
	if len(fileIDs) > 0 {
		sort.Slice(fileIDs, func(i, j int) bool {
			return fileIDs[i] < fileIDs[j]
		})
	}
	return fileIDs, err
}
func NewBitcask(path string) *Bitcask {
	//scan the directory, get all the file id
	path = strings.TrimSuffix(path, "/")
	// if directory is not exist, create it
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}
	path = path + "/"
	fileIDs, err := ScanDir(path)
	if err != nil {
		panic(err)
	}
	// open the file
	if len(fileIDs) == 0 {
		// create a new file
		fileIDs = append(fileIDs, 1)
	}
	memdb := NewSkipListArr()
	b := &Bitcask{
		Path:    path,
		FileIDs: fileIDs,
		memDB:   memdb,
	}
	for _, fileID := range fileIDs {
		file := NewFile(fileID, path)
		file.OpenFile()
		b.Files = append(b.Files, file)
	}

	if err != nil {
		panic(err)
	}

	for _, file := range b.Files {

		for {
			entry, err := file.ReadEntry()
			if err != nil {
				break
			}
			if e := b.memDB.Search(entry); e != nil {
				// update the entry
				if e.ValueSize == 0 {
					// delete the entry
					b.memDB.Delete(e)
				}
				e.ValuePos = entry.ValuePos
				e.TimeStamp = entry.TimeStamp
				e.ValueSize = entry.ValueSize
			} else {
				b.memDB.Insert(entry)
			}
		}
	}
	return b
}

func (b *Bitcask) Open() error {
	b.CurrentFile = b.Files[len(b.Files)-1]
	b.currentFileID = b.FileIDs[len(b.FileIDs)-1]
	return nil
}

func (b *Bitcask) Close() error {

	for _, file := range b.Files {
		if err := file.CloseFile(); err != nil {
			return err
		}
	}
	return nil
}

func (b *Bitcask) Put(key []byte, value []byte) error {
	// write the record to the file
	// check the size of the file, if it's full, create a new file
	if b.CurrentFile.CurrentPos+uint32(len(key)+len(value))+20 > MaxFileSize {
		// create a new file
		b.currentFileID++
		file := NewFile(b.currentFileID, b.Path)
		b.FileIDs = append(b.FileIDs, b.currentFileID)
		file.OpenFile()
		b.CurrentFile = file
	}
	record, err := b.CurrentFile.WriteRecord(key, value)
	if err != nil {
		return err
	}
	// insert the entry into the memDB
	entry := NewEntry(key, b.currentFileID, record.ValueSize, record.ValuePos, record.TimeStamp)
	// first search in the memDB
	if e := b.memDB.Search(entry); e != nil {
		// update the entry
		e.ValueSize = record.ValueSize
		e.ValuePos = record.ValuePos
		e.TimeStamp = record.TimeStamp
		if e.ValueSize == 0 {
			// delete the entry
			b.memDB.Delete(e)
		}
	} else {
		b.memDB.Insert(entry)
	}
	b.CurrentFile.Sync()
	return nil
}

func (b *Bitcask) Get(key []byte) (*Record, error) {
	tmp := NewTmpEntry(key)
	entry := b.memDB.Search(tmp)
	if entry != nil {
		// read the value from the file
		f := b.Files[entry.FileID-1]
		value, err := f.Read(entry.ValuePos, entry.ValueSize)
		if err != nil {
			return nil, err
		}
		rec := NewRecord(entry.TimeStamp, entry.Key, entry.ValuePos, value)
		return rec, nil
	}
	return nil, nil
}
