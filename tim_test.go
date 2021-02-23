package tim

import (
	"fmt"
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	testCases := []struct {
		in  string
		out string
	}{
		{"cộng hoa xã hội chủ nghĩa Việt. Nam", "cong hoa xa hoi nam chu nghia viet cong-hoa xa-hoi hoi-chu hoa-xa chu-nghia nghia-viet viet-nam"},
		{`phạm kiều
 thanh`, "pham kieu thanh pham-kieu kieu-thanh"},
	}
	for _, tc := range testCases {
		out := tokenize(tc.in)
		if len(out) != len(strings.Split(tc.out, " ")) {
			fmt.Println(out)
			t.Errorf("Len should be eq for tc %s", tc.in)
		}
		outM := map[string]bool{}
		for _, term := range out {
			outM[term] = true
		}

		for _, term := range strings.Split(tc.out, " ") {
			if !outM[term] {
				fmt.Println("OUT FOR ", tc.in, out)
				t.Errorf("MISSING term %s for tc %s", term, tc.in)
			}
		}
	}
}

func TestNGram(t *testing.T) {
	testCases := []struct {
		in  string
		out string
	}{
		{"cong", "con cong ong"},
		{"123456", "123 1234 12345 12345 234 2345 23456 345 3456 456"},
	}
	for _, tc := range testCases {
		out := n_gram(tc.in, 3)
		if len(out) != len(strings.Split(tc.out, " ")) {
			fmt.Println(out)
			t.Errorf("Len should be eq for tc %s", tc.in)
		}
		outM := map[string]bool{}
		for _, term := range out {
			outM[term] = true
		}

		for _, term := range strings.Split(tc.out, " ") {
			if !outM[term] {
				fmt.Println("OUT FOR ", tc.in, out)
				t.Errorf("MISSING term %s for tc %s", term, tc.in)
			}
		}
	}
}

func TestShards(t *testing.T)  {
	fmt.Println(makeShards(10))
}
