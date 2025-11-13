package bitcask

import "bytes"

type Entry struct {
	FileID    uint32 // unit32 最大能表示5G数字， 所以uni32 已经够用
	TimeStamp uint32 //unit32 时间戳，以秒计算，可以表示136年
	ValueSize uint32
	ValuePos  uint32 // value 在数据文件中的偏移位置, 每个文件不超过1GB， 所以使用uint32
	Key       []byte
}

func NewEntry(key []byte, fileId uint32, valueSize uint32, valuePos uint32, timeStamp uint32) *Entry {
	return &Entry{
		Key:       key,
		TimeStamp: timeStamp,
		FileID:    fileId,
		ValueSize: valueSize,
		ValuePos:  valuePos,
	}
}

func NewTmpEntry(key []byte) *Entry {
	return &Entry{
		Key:       key,
		TimeStamp: 0,
		FileID:    0,
		ValueSize: 0,
		ValuePos:  0,
	}
}

// Compare compares two entries based on key, then timestamp, then file ID
// Returns:
//
//	-1 if e < other
//	 0 if e == other
//	 1 if e > other
func (e *Entry) Compare(other *Entry) int {
	if e == nil && other == nil {
		return 0
	}
	if e == nil {
		return -1
	}
	if other == nil {
		return 1
	}

	// First compare by key
	keyCompare := bytes.Compare(e.Key, other.Key)
	if keyCompare != 0 {
		return keyCompare
	}

	// If keys are equal, compare by timestamp (newer first)
	// if e.TimeStamp != other.TimeStamp {
	// 	if e.TimeStamp > other.TimeStamp {
	// 		return 1
	// 	}
	// 	return -1
	// }
	// All fields are equal
	return 0
}

// Less returns true if e should be ordered before other
func (e *Entry) Less(other *Entry) bool {
	return e.Compare(other) < 0
}

func (e *Entry) LessEq(other *Entry) bool {
	return e.Compare(other) < 0 || e.Compare(other) == 0
}

// Equal returns true if e and other have the same key, timestamp, and file ID
func (e *Entry) Equal(other *Entry) bool {
	return e.Compare(other) == 0
}

// Greater returns true if e should be ordered after other
func (e *Entry) Greater(other *Entry) bool {
	return e.Compare(other) > 0
}

func (e *Entry) GreaterEq(other *Entry) bool {
	return e.Compare(other) > 0 || e.Compare(other) == 0

}

type Entries []*Entry

func (e Entries) Len() int { return len(e) }
func (e Entries) Less(i, j int) bool {
	// First compare by key
	keyCompare := bytes.Compare(e[i].Key, e[j].Key)
	if keyCompare != 0 {
		return keyCompare < 0
	}

	// If keys are equal, compare by timestamp (newest first)
	// if e[i].TimeStamp != e[j].TimeStamp {
	// 	return e[i].TimeStamp > e[j].TimeStamp
	// }
	return false
}
func (e Entries) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
