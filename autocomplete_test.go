package tim

import (
	"fmt"
	"testing"
)

func TestKeysWithPrefix(t *testing.T) {
	mgr := NewPrefixTermMgr(func(col, accid string) []string {
		strarr := []string{"abc", "abcd", "bc", "bcd", "phuong cuoi", "khanh hoa", "khanh linh"}
		for _, r := range "abcdefghijklmnopqrstuvwxyz" {
			strarr = append(strarr, "dieu"+string(r))
		}
		return strarr
	})
	accid := "testacc"
	fmt.Printf("%#v", mgr.KeysWithPrefix("", accid, "dieu"))
	t.Error("TRUE")
}
