package bitcask

import (
	"fmt"
	"testing"
)

func Test_base(t *testing.T) {
	skipArr := NewSkipListArr()
	skipArr.Insert(NewTmpEntry([]byte("10")))
	skipArr.Insert(NewTmpEntry([]byte("11")))
	skipArr.Insert(NewTmpEntry([]byte("12")))
	skipArr.Insert(NewTmpEntry([]byte("13")))
	key := skipArr.Search(NewTmpEntry([]byte("10")))
	fmt.Println("Search 10:", key)
	fmt.Println(skipArr)
}
