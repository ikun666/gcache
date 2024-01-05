package consistentHash

import (
	"fmt"
	"strconv"
	"testing"
)

func TestHash(t *testing.T) {
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// Given the above hash function, this will give replicas with "hashes":
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.Add("6", "4", "2")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
		"29": "2",
	}

	for k, v := range testCases {
		fmt.Println(k, hash.Get(k), v)
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
	fmt.Println("-----------------")
	// Adds 8, 18, 28
	hash.Add("8")
	// 2, 4, 6,8, 12, 14, 16,18, 22, 24, 26,28
	// 27 should now map to 8.
	testCases["27"] = "8"

	for k, v := range testCases {
		fmt.Println(k, hash.Get(k), v)
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
	fmt.Println("-----------------")
	hash.Remove("2")
	// 4, 6,8, 14, 16,18, 24, 26,28
	testCases["2"] = "4"
	testCases["11"] = "4"
	testCases["23"] = "4"
	testCases["29"] = "4"
	for k, v := range testCases {
		fmt.Println(k, hash.Get(k), v)
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
}
