package bitcask

import (
	"fmt"
	"testing"
)

func Benchmark_Put(b *testing.B) {
	bitcask := NewBitcask("/tmp/benchmark_put/")
	defer bitcask.Close()

	err := bitcask.Open()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte(fmt.Sprintf("value_%0122d", i))
		err := bitcask.Put(key, value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Get(b *testing.B) {
	bitcask := NewBitcask("/tmp/benchmark_get/")
	defer bitcask.Close()

	err := bitcask.Open()
	if err != nil {
		b.Fatal(err)
	}

	// 预先插入数据
	for i := 0; i < 10000; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte(fmt.Sprintf("value_%0122d", i))
		err := bitcask.Put(key, value)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("key_%d", i%10000))
		_, err := bitcask.Get(key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_PutAndGetProperty(b *testing.B) {
	bitcask := NewBitcask("/tmp/benchmark_put_get/")
	defer bitcask.Close()

	err := bitcask.Open()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte(fmt.Sprintf("value_%0122d", i))

		err := bitcask.Put(key, value)
		if err != nil {
			b.Fatal(err)
		}

		_, err = bitcask.Get(key)
		if err != nil {
			b.Fatal(err)
		}
	}
}
