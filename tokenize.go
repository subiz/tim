package tim

import (
	"regexp"
	"strings"
)

type MatchLiteral struct {
	Str  string
	Psrc []int
}

func tokenize(str string) []string {
	tokens := make([]string, 0)
	tokens = append(tokens, tokenizeLiteral(str)...)
	tokens = append(tokens, Email_regexp.FindAllString(str, -1)...)
	tokens = append(tokens, findPersonalPhoneNumber(str)...)
	return tokens
}

const Email_regex = `([a-zA-Z0-9._-]+@[a-zA-Z0-9._-]+\.[a-zA-Z0-9_-]+)`
const Email_letter = "abcdefghijklmnopqrstuvwxyz"
const Email_digit = "0123456789"
const Email_format = "@-_."
const Email_min_len = 5

var Email_regexp = regexp.MustCompile(Email_regex)
var Email_norm_map map[rune]rune

func findEmail(str string) []string {
	if len(str) < Email_min_len {
		return nil
	}
	str += " "
	emails := make([]string, 0)
	rarr := []rune(str)
	from, to := 0, 0
	for i, r := range rarr {
		if _, has := Email_norm_map[r]; has {
			to++
			continue
		}

		token := string(rarr[from:to])
		if len(token) >= Email_min_len && Email_regexp.MatchString(token) {
			normtoken := make([]rune, len(token))
			for j, tr := range token {
				normtoken[j] = Email_norm_map[tr]
			}
			// TODO trim format rune
			emails = append(emails, string(normtoken))
		}

		from, to = i+1, i+1
	}
	return emails
}

// +84 2473.021.368
const PersonalPhoneNumber_digit = "0123456789"
const PersonalPhoneNumber_format = "+() .-"
const PersonalPhoneNumber_min_len = 7

var PersonalPhoneNumber_regexp = regexp.MustCompile(RegexPhone)
var PersonalPhoneNumber_norm_map map[rune]rune

func findPersonalPhoneNumber(str string) []string {
	if len(str) < PersonalPhoneNumber_min_len {
		return nil
	}
	str += ","
	phoneNumbers := make([]string, 0)
	rarr := []rune(str)
	from, to := 0, 0
	for i, r := range rarr {
		if _, has := PersonalPhoneNumber_norm_map[r]; has {
			to++
			continue
		}

		token := string(rarr[from:to])
		if len(token) >= PersonalPhoneNumber_min_len {
			normtoken := make([]rune, 0)
			for _, tr := range token {
				if '0' <= tr && tr <= '9' {
					normtoken = append(normtoken, tr)
				}
			}
			if len(normtoken) >= PersonalPhoneNumber_min_len {
				phoneNumbers = append(phoneNumbers, string(normtoken))
			}
		}

		from, to = i+1, i+1
	}
	return phoneNumbers
}

// TODO psrc
func tokenizeLiteral(str string) []string {
	literals := make([]*MatchLiteral, 0)
	biliterals := make([]*MatchLiteral, 0)

	str += " "
	rarr := []rune(str)
	from, to := 0, 0
	var prevliteral string
	for i, r := range rarr {
		if _, has := Norm_map[r]; has {
			to++
			continue
		}

		normtoken := make([]rune, to-from)
		for j, tr := range rarr[from:to] {
			normtoken[j] = Norm_map[tr]
		}
		literal := string(normtoken)
		// TODO trim format rune
		if len(literal) > 0 && !STOP_WORDS[literal] && isLiteral(literal) {
			literals = append(literals, &MatchLiteral{Str: literal})
		}
		if len(prevliteral) > 0 && len(literal) > 0 && !STOP_WORDS[literal] && isLiteral(prevliteral) {
			biliterals = append(biliterals, &MatchLiteral{Str: string(prevliteral) + " " + literal})
		}

		prevliteral = literal
		from, to = i+1, i+1
	}

	strarr := make([]string, len(literals)+len(biliterals))
	strindex := 0
	for _, literal := range literals {
		strarr[strindex] = literal.Str
		strindex++
	}
	for _, literal := range biliterals {
		strarr[strindex] = literal.Str
		strindex++
	}
	return strarr
}

func isLiteral(token string) bool {
	if len(token) < 3 || len(token) > 45 {
		return false
	}
	// TODO first consonants
	// TOOD rhyme: accompaniment, main sound, end sound
	found := false
	for _, r := range token {
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

const Token_min_len = 3
const Token_max_len = 45
const Vietnam_word_max_len = 7
const Vietnam_vowel = "i, e, ê, ư, u, o, ô, ơ, a, ă, â"

var Vietnam_vowel_unaccented_map map[rune]struct{}

var Str_letter = "abcdefghijklmnopqrstuvwxyz"
var Str_digit = "0123456789"
var Str_special = "-_"

var Norm_map map[rune]rune

const RegexPhone = `([0-9._-]{3,})`

var Regexp_phone = regexp.MustCompile(RegexPhone)

func initEmailKit() {
	Email_norm_map = make(map[rune]rune)
	for _, r := range Email_digit {
		Email_norm_map[r] = r
	}
	for _, r := range Email_letter {
		Email_norm_map[r] = r
	}
	upperstr := strings.ToUpper(Email_letter)
	runeArr := []rune(Email_letter)
	for i, r := range upperstr {
		Email_norm_map[r] = runeArr[i]
	}
	for _, r := range Email_format {
		Email_norm_map[r] = r
	}
}

func initPhoneKit() {
	PersonalPhoneNumber_norm_map = make(map[rune]rune)
	for _, r := range PersonalPhoneNumber_digit {
		PersonalPhoneNumber_norm_map[r] = r
	}
	for _, r := range PersonalPhoneNumber_format {
		PersonalPhoneNumber_norm_map[r] = r
	}
}

func init() {
	initEmailKit()
	initPhoneKit()

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
	for _, r := range Str_special {
		Norm_map[r] = r
	}

	for vi, r := range Vietnam_letter {
		Norm_map[vi] = r
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
	for i, upperal := range upperaccentedLetters {
		Norm_map[upperal] = unaccentedLetters[i]
	}

	Vietnam_vowel_unaccented_map = make(map[rune]struct{})
	for _, r := range Vietnam_vowel {
		if nr, has := Norm_map[r]; has {
			Vietnam_vowel_unaccented_map[nr] = struct{}{}
		}
	}
}
