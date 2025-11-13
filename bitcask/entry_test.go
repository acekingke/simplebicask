package bitcask

import (
	"bytes"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewEntry(t *testing.T) {
	entry := NewEntry([]byte("key"), 1, 10, 100, 1000)
	if entry.FileID != 1 {
		t.Error("file id error")
	}
	if !bytes.Equal(entry.Key, []byte("key")) {
		t.Error("key error")
	}
	if entry.TimeStamp != 1000 {
		t.Error("time stamp error")
	}
}

func Test_EntriesTest(t *testing.T) {
	entries := Entries{
		NewEntry([]byte("key2"), 1, 10, 100, 1000),
		NewEntry([]byte("key1"), 1, 10, 100, 1000),
		NewEntry([]byte("key3"), 1, 10, 100, 1000),
	}
	sort.Sort(entries)
	if !bytes.Equal(entries[0].Key, []byte("key1")) {
		t.Error("key error")
	}
}

func Test_EntryGreater(t *testing.T) {
	entry := NewEntry([]byte("key"), 1, 10, 100, 1000)
	if entry.Greater(NewEntry([]byte("key1"), 1, 10, 100, 1000)) {
		t.Error("greater error")
	}
	if entry.Greater(NewEntry([]byte("key"), 2, 10, 100, 1000)) {
		t.Error("greater error")
	}
}
func Test_Equal(t *testing.T) {
	entry := NewEntry([]byte("key"), 1, 10, 100, 1000)
	if entry.Equal(NewEntry([]byte("key1"), 1, 10, 100, 1000)) {
		t.Error("equal error")
	}
	assert.True(t, entry.Equal(NewEntry([]byte("key"), 2, 10, 100, 1000)))
}

func Test_EntryLess(t *testing.T) {
	entry := NewEntry([]byte("key"), 1, 10, 100, 1000)
	if entry.Less(NewEntry([]byte("ke"), 1, 10, 100, 1000)) {
		t.Error("less error")
	}
	assert.False(t, entry.Less(NewEntry([]byte("key"), 2, 10, 100, 1000)))

}
