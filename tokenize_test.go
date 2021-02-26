package tim

import (
	"fmt"
	"testing"
)

func TestTokenizeLiteral(t *testing.T) {
	str := `可口的 1 _ - 123 accompaniment nnn Trụ sở: (Tầng 6), tòa nhà Kailash,
		ngõ 92 Trần Thái Tông, Di3u@gmail.com Phường Dịch Vọng Hậu, Quận Cầu Giấy, Hà Nội (84)123123211 dieu
		https://translate.google.com/?hl=vi&sl=en&tl=vi&text=concrete&op=translate
		pneumonoultramicroscopicsilicovolcanoconiosis viet nam. hello viet.nam
		va co gmail com`
	literals := tokenize(str)
	fmt.Printf("%#v", literals)
	fmt.Println("")
	fmt.Printf("%#v", tokenize("cong hoa xa"))
	fmt.Printf("%#v", tokenize("cong hoa xa hoi chu nghia viet nam"))
	fmt.Printf("%#v", tokenize("cong, hoa, xa, hoi"))
	if true {
		t.Error("TRUE")
	}
}

func TestFindEmail(t *testing.T) {
	str := "hi, email is di3u@gmail.com., dieu2@gmail.com, end."
	emails := findEmail(str)
	fmt.Printf("%#v", emails)
	t.Error("TRUE")
}

func TestFindPersonalPhoneNumber(t *testing.T) {
	str := "1 _ - 123 accompaniment hi, 123 phone number is +(84) 0974-304-123., 0974403666 end. +84 2473.021.368"
	phones := findPersonalPhoneNumber(str)
	fmt.Printf("%#v", phones)
	t.Error("TRUE")
}
