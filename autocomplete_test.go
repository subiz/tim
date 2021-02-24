package tim

import (
	"fmt"
	"testing"
)

func TestKeysWithPrefix(t *testing.T) {
	mgr := NewPrefixTermMgr(func(col, accid string) []string {
		return []string{"abc", "abcd", "bc", "bcd", "phuong cuoi", "khanh hoa", "khanh linh"}
	})
	accid := "testacc"
	fmt.Printf("%#v", mgr.KeysWithPrefix("", accid, "kha"))
	t.Error("TRUE")
}
