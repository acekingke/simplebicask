package bitcask

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

type RecordHeader struct {
	Crc       uint32 // unit32 最大能表示5G数字， 所以uni32 已经够用
	TimeStamp uint32 //unit32 时间戳，以秒计算，可以表示136年
	KeySize   uint32
	ValueSize uint32
	ValuePos  uint32 // value 在数据文件中的偏移位置, 每个文件不超过1GB， 所以使用uint32
}

type Record struct {
	Crc       uint32 // unit32 最大能表示5G数字， 所以uni32 已经够用
	TimeStamp uint32 //unit32 时间戳，以秒计算，可以表示136年
	KeySize   uint32
	ValueSize uint32
	ValuePos  uint32 // value 在数据文件中的偏移位置, 每个文件不超过1GB， 所以使用uint32
	Key       []byte
	Value     []byte
}

func NewRecord(timeStamp uint32, key []byte, valuePos uint32, value []byte) *Record {
	h := crc32.NewIEEE()
	h.Write(key)
	h.Write(value)
	return &Record{
		Crc:       h.Sum32(),
		TimeStamp: timeStamp,
		KeySize:   uint32(len(key)),
		ValueSize: uint32(len(value)),
		ValuePos:  valuePos,
		Key:       key,
		Value:     value,
	}
}

func (r *Record) Encode() []byte {
	// bigendian
	data := make([]byte, 4+4+4+4+4+len(r.Key)+len(r.Value))
	binary.BigEndian.PutUint32(data[0:4], r.Crc)
	binary.BigEndian.PutUint32(data[4:8], r.TimeStamp)
	binary.BigEndian.PutUint32(data[8:12], r.KeySize)
	binary.BigEndian.PutUint32(data[12:16], r.ValueSize)
	binary.BigEndian.PutUint32(data[16:20], r.ValuePos)
	copy(data[20:20+len(r.Key)], r.Key)
	copy(data[20+len(r.Key):], r.Value)
	return data
}
func Decode(data []byte) (*Record, error) {
	crc := binary.BigEndian.Uint32(data[0:4])
	timeStamp := binary.BigEndian.Uint32(data[4:8])
	keySize := binary.BigEndian.Uint32(data[8:12])
	valueSize := binary.BigEndian.Uint32(data[12:16])
	valuePos := binary.BigEndian.Uint32(data[16:20])

	key := make([]byte, keySize)
	value := make([]byte, valueSize)

	copy(key, data[20:20+keySize])
	copy(value, data[20+keySize:20+keySize+valueSize])
	record := NewRecord(timeStamp, key, valuePos, value)
	if crc != record.Crc {

		return nil, fmt.Errorf("crc32 check failed")
	}
	record.ValuePos = valuePos // Set ValuePos separately since it's not in constructor
	return record, nil

}
func DecodeHeader(data []byte) *RecordHeader {
	crc := binary.BigEndian.Uint32(data[0:4])
	timeStamp := binary.BigEndian.Uint32(data[4:8])
	keySize := binary.BigEndian.Uint32(data[8:12])
	valueSize := binary.BigEndian.Uint32(data[12:16])
	valuePos := binary.BigEndian.Uint32(data[16:20])
	return &RecordHeader{
		Crc:       crc,
		TimeStamp: timeStamp,
		KeySize:   keySize,
		ValueSize: valueSize,
		ValuePos:  valuePos,
	}
}

// bufio write records
