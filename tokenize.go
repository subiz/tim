package tim

import (
	"regexp"
	"strings"
)

type MatchLiteral struct {
	Str  string
	Psrc []int
}

func tokenizeLiteral(str string) []string {
	literals := make([]*MatchLiteral, 0)
	biliterals := make([]*MatchLiteral, 0)
	concreteliterals := make([]*MatchLiteral, 0)
	var iter rune
	token := make([]rune, 0)
	var prevtoken []rune
	// TODO psrc
	str += " "
	for _, r := range str {
		if iter == ' ' && r == ' ' {
			continue
		}
		if _, has := Str_split_map[r]; has {
			if len(token) > 0 && isLiteral(&token) {
				literals = append(literals, &MatchLiteral{Str: string(token)})
			}
			if len(prevtoken) > 0 && len(token) > 0 && isLiteral(&prevtoken) {
				biliterals = append(biliterals, &MatchLiteral{Str: string(prevtoken) + " " + string(token)})
			}
			if len(token) > 0 && (Regexp_email.MatchString(string(token)) || Regexp_phone.MatchString(string(token))) {
				concreteliterals = append(concreteliterals, &MatchLiteral{Str: string(token)})
			}
			prevtoken = token
			token = make([]rune, 0)
			continue
		}
		iter = r
		if nr, has := Norm_map[r]; has {
			token = append(token, nr)
		}
	}
	strarr := make([]string, len(literals)+len(biliterals)+len(concreteliterals))
	strindex := 0
	for _, literal := range literals {
		strarr[strindex] = literal.Str
		strindex++
	}
	for _, literal := range biliterals {
		strarr[strindex] = literal.Str
		strindex++
	}
	for _, literal := range concreteliterals {
		strarr[strindex] = literal.Str
		strindex++
	}
	return strarr
}

func isLiteral(token *[]rune) bool {
	if len(*token) > 45 {
		return false
	}
	// TODO first consonants
	// TOOD rhyme: accompaniment, main sound, end sound
	found := false
	for _, r := range *token {
		if _, has := Vietnam_vowel_unaccented_map[r]; has {
			found = true
			break
		}
	}
	if !found {
		return false
	}

	return true
}

var Vietnam_letter = map[rune]rune{
	'ạ': 'a', 'ả': 'a', 'ã': 'a', 'à': 'a', 'á': 'a', 'â': 'a', 'ậ': 'a', 'ầ': 'a', 'ấ': 'a',
	'ẩ': 'a', 'ẫ': 'a', 'ă': 'a', 'ắ': 'a', 'ằ': 'a', 'ặ': 'a', 'ẳ': 'a', 'ẵ': 'a',
	'ó': 'o', 'ò': 'o', 'ọ': 'o', 'õ': 'o', 'ỏ': 'o', 'ô': 'o', 'ộ': 'o', 'ổ': 'o', 'ỗ': 'o',
	'ồ': 'o', 'ố': 'o', 'ơ': 'o', 'ờ': 'o', 'ớ': 'o', 'ợ': 'o', 'ở': 'o', 'ỡ': 'o',
	'é': 'e', 'è': 'e', 'ẻ': 'e', 'ẹ': 'e', 'ẽ': 'e', 'ê': 'e', 'ế': 'e', 'ề': 'e', 'ệ': 'e', 'ể': 'e', 'ễ': 'e',
	'ú': 'u', 'ù': 'u', 'ụ': 'u', 'ủ': 'u', 'ũ': 'u', 'ư': 'u', 'ự': 'u', 'ữ': 'u', 'ử': 'u', 'ừ': 'u', 'ứ': 'u',
	'í': 'i', 'ì': 'i', 'ị': 'i', 'ỉ': 'i', 'ĩ': 'i',
	'ý': 'y', 'ỳ': 'y', 'ỷ': 'y', 'ỵ': 'y', 'ỹ': 'y',
	'đ': 'd',
}

const Vietnam_word_max_len = 7
const Vietnam_vowel = "i, e, ê, ư, u, o, ô, ơ, a, ă, â"

var Vietnam_vowel_unaccented_map map[rune]struct{}

var Str_split = ` ,;:/\&=`
var Str_letter = "abcdefghijklmnopqrstuvwxyz"
var Str_digit = "0123456789"
var Str_special = "@-_."

var Norm_map map[rune]rune
var Str_split_map map[rune]struct{}

const RegexEmail = `([a-zA-Z0-9._-]+@[a-zA-Z0-9._-]+\.[a-zA-Z0-9_-]+)`
const RegexPhone = `([0-9._-]+)`

var Regexp_email = regexp.MustCompile(RegexEmail)
var Regexp_phone = regexp.MustCompile(RegexPhone)

func init() {
	Norm_map = make(map[rune]rune)
	for _, r := range Str_letter {
		Norm_map[r] = r
	}
	upperstr := strings.ToUpper(Str_letter)
	runeArr := []rune(Str_letter)
	for i, r := range upperstr {
		Norm_map[r] = runeArr[i]
	}
	for _, r := range Str_digit {
		Norm_map[r] = r
	}
	Str_split_map = make(map[rune]struct{})
	for _, r := range Str_split {
		// Norm_map[r] = r
		Str_split_map[r] = struct{}{}
	}
	for _, r := range Str_special {
		Norm_map[r] = r
	}

	accentedLetters := make([]rune, len(Vietnam_letter))
	unaccentedLetters := make([]rune, len(Vietnam_letter))
	lindex := 0
	for al, unal := range Vietnam_letter {
		accentedLetters[lindex] = al
		unaccentedLetters[lindex] = unal
		lindex++
	}
	upperaccentedLetters := make([]rune, len(accentedLetters))
	upperalindex := 0
	for _, r := range strings.ToUpper(string(accentedLetters)) {
		upperaccentedLetters[upperalindex] = r
		upperalindex++
	}

	for vi, r := range Vietnam_letter {
		Norm_map[vi] = r
	}

	Vietnam_vowel_unaccented_map = make(map[rune]struct{})
	for _, r := range Vietnam_vowel {
		if nr, has := Norm_map[r]; has {
			Vietnam_vowel_unaccented_map[nr] = struct{}{}
		}
	}
}
