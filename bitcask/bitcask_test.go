package bitcask

import (
	"fmt"
	"testing"
)

func Test_Bitcask(t *testing.T) {
	b := NewBitcask("/tmp/")
	b.Open()
	b.Put([]byte("key"), []byte("value"))
	b.Put([]byte("key"), []byte("value2"))
	b.Close()
}

func Test_ScanDir(t *testing.T) {
	b := NewBitcask("/tmp/")
	b.Open()
	key := []byte("key")
	v, _ := b.Get(key)
	fmt.Println(string(v.Key))
	fmt.Println(v.ValuePos)
	fmt.Println(string(v.Value))
	//b.Put([]byte("key"), []byte("value2"))
	b.Close()
}
