package bitcask

import (
	"fmt"
	"testing"
)

func TestFile(t *testing.T) {
	f := NewFile(1, "/tmp/")
	err := f.OpenFile()
	if err != nil {
		t.Error(err)
	}
	rec, err := f.WriteRecord([]byte("key2"), []byte("value2"))
	if err != nil {
		t.Error(err)
	}
	println(rec)
	f.CloseFile()
}

func TestUpdate(t *testing.T) {
	f := NewFile(2, "/tmp/")
	err := f.OpenFile()
	if err != nil {
		t.Error(err)
	}
	rec, err := f.WriteRecord([]byte("key2"), []byte("value2"))
	if err != nil {
		t.Error(err)
	}
	println(rec)
	// update record, add new record with same key
	rec, err = f.WriteRecord([]byte("key2"), []byte("value3"))
	if err != nil {
		t.Error(err)
	}
	println(rec)
	// delete record, write empty value
	rec, err = f.WriteRecord([]byte("key2"), []byte(""))
	if err != nil {
		t.Error(err)
	}
	fmt.Println(rec)
	f.CloseFile()
}

func TestRead(t *testing.T) {
	f := NewFile(1, "/tmp/")
	err := f.OpenFile()
	if err != nil {
		t.Error(err)
	}
	var buf []byte
	buf, err = f.Read(114, 6)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(buf))
}
