package consistenthash

import (
	"fmt"
	"strconv"
	"testing"
)

func TestHashing(t *testing.T)  {
	consisHash := New(3, func(key []byte) uint32 {
		atoi, _ := strconv.Atoi(string(key))
		return uint32(atoi)
	})

	// Given the above hash function, this will give replicas with "hashes":
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	// eg 2 虚拟节点为 02, 12, 22
	consisHash.Add("6", "4", "2")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		if consisHash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

	// add 08, 18, 28
	consisHash.Add("8")

	// 此时 27 应该映射到 8
	testCases["27"] = "8"

	for k, v := range testCases {
		if consisHash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
	fmt.Println(consisHash.Show())
	consisHash.Remove("4")
	fmt.Println("-----------------------")
	fmt.Println(consisHash.Show())
}
