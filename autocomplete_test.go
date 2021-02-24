package tim

import (
	"fmt"
	"testing"
)

func TestKeysWithPrefix(t *testing.T) {
	mgr := NewPrefixTermMgr(func(col, accid string) []string {
		return []string{"abc", "abcd", "bc", "bcd"}
	})
	accid := "testacc"
	fmt.Printf("%#v", mgr.KeysWithPrefix("", accid, "ab"))
	t.Error("TRUE")
}
