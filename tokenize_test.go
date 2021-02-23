package tim

import (
	"fmt"
	"testing"
)

func TestTokenizeLiteral(t *testing.T) {
	str := "accompaniment nnn Trụ sở: (Tầng 6), tòa nhà Kailash, ngõ 92 Trần Thái Tông, di3u@gmail.com Phường Dịch Vọng Hậu, Quận Cầu Giấy, Hà Nội (84)123123211 dieu " +
		"https://translate.google.com/?hl=vi&sl=en&tl=vi&text=concrete&op=translate pneumonoultramicroscopicsilicovolcanoconiosis viet nam. hello"
	literals := tokenize(str)
	fmt.Printf("%#v", literals)
	if true {
		t.Error("TRUE")
	}
}
