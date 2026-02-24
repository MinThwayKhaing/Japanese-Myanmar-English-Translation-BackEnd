package utils

import "strings"

// romajiTable maps romaji sequences (longest first within each length group) to hiragana.
var romajiTable = map[string]string{
	// 4-char
	"tchi": "っち",
	"ttsu": "っつ",

	// 3-char
	"sha": "しゃ", "shi": "し", "shu": "しゅ", "she": "しぇ", "sho": "しょ",
	"chi": "ち", "cha": "ちゃ", "chu": "ちゅ", "che": "ちぇ", "cho": "ちょ",
	"tsu": "つ",
	"kya": "きゃ", "kyu": "きゅ", "kyo": "きょ",
	"nya": "にゃ", "nyu": "にゅ", "nyo": "にょ",
	"hya": "ひゃ", "hyu": "ひゅ", "hyo": "ひょ",
	"mya": "みゃ", "myu": "みゅ", "myo": "みょ",
	"rya": "りゃ", "ryu": "りゅ", "ryo": "りょ",
	"gya": "ぎゃ", "gyu": "ぎゅ", "gyo": "ぎょ",
	"bya": "びゃ", "byu": "びゅ", "byo": "びょ",
	"pya": "ぴゃ", "pyu": "ぴゅ", "pyo": "ぴょ",
	"dya": "ぢゃ", "dyu": "ぢゅ", "dyo": "ぢょ",
	"zya": "じゃ", "zyu": "じゅ", "zyo": "じょ",
	"jya": "じゃ", "jyu": "じゅ", "jyo": "じょ",

	// 2-char
	"ka": "か", "ki": "き", "ku": "く", "ke": "け", "ko": "こ",
	"sa": "さ", "si": "し", "su": "す", "se": "せ", "so": "そ",
	"ta": "た", "ti": "ち", "tu": "つ", "te": "て", "to": "と",
	"na": "な", "ni": "に", "nu": "ぬ", "ne": "ね", "no": "の",
	"ha": "は", "hi": "ひ", "fu": "ふ", "hu": "ふ", "he": "へ", "ho": "ほ",
	"ma": "ま", "mi": "み", "mu": "む", "me": "め", "mo": "も",
	"ya": "や", "yu": "ゆ", "yo": "よ",
	"ra": "ら", "ri": "り", "ru": "る", "re": "れ", "ro": "ろ",
	"wa": "わ", "wi": "ゐ", "we": "ゑ", "wo": "を",
	"ga": "が", "gi": "ぎ", "gu": "ぐ", "ge": "げ", "go": "ご",
	"za": "ざ", "zi": "じ", "ji": "じ", "zu": "ず", "ze": "ぜ", "zo": "ぞ",
	"da": "だ", "di": "ぢ", "du": "づ", "de": "で", "do": "ど",
	"ba": "ば", "bi": "び", "bu": "ぶ", "be": "べ", "bo": "ぼ",
	"pa": "ぱ", "pi": "ぴ", "pu": "ぷ", "pe": "ぺ", "po": "ぽ",
	"ja": "じゃ", "ju": "じゅ", "je": "じぇ", "jo": "じょ",
	"nn": "ん",

	// 1-char vowels
	"a": "あ", "i": "い", "u": "う", "e": "え", "o": "お",
}

// IsRomaji reports whether s consists entirely of ASCII letters (and spaces),
// which suggests it may be romaji input rather than Japanese/Myanmar script.
func IsRomaji(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == ' ') {
			return false
		}
	}
	return true
}

// isConsonant returns true for ASCII consonant bytes.
func isConsonant(b byte) bool {
	switch b {
	case 'a', 'e', 'i', 'o', 'u':
		return false
	default:
		return b >= 'a' && b <= 'z'
	}
}

// RomajiToHiragana converts a romaji string to hiragana.
// Non-romaji characters are passed through unchanged.
// Double consonants (e.g., "kk", "tt") produce っ before the syllable.
// "n" before a consonant or at end of word becomes ん.
func RomajiToHiragana(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	var result strings.Builder
	i := 0

	for i < len(s) {
		// Double consonant → っ (skip 'n' which has its own rule)
		if i+1 < len(s) && s[i] != 'n' && isConsonant(s[i]) && s[i] == s[i+1] {
			result.WriteString("っ")
			i++
			continue
		}

		// Special "n" handling to avoid treating "na/ni/nu/ne/no" as ん+vowel
		if s[i] == 'n' {
			// "nn" → ん
			if i+1 < len(s) && s[i+1] == 'n' {
				result.WriteString("ん")
				i += 2
				continue
			}
			// Try 3-char and 2-char patterns that start with n (nya, na, ni, …)
			matched := false
			for _, length := range []int{3, 2} {
				if i+length <= len(s) {
					if h, ok := romajiTable[s[i:i+length]]; ok {
						result.WriteString(h)
						i += length
						matched = true
						break
					}
				}
			}
			if !matched {
				// Standalone n: ん before a consonant or at end of string
				if i+1 >= len(s) || isConsonant(s[i+1]) {
					result.WriteString("ん")
				} else {
					// n before a vowel with no pattern match – keep as-is
					result.WriteByte(s[i])
				}
				i++
			}
			continue
		}

		// General case: try longest match (4 → 3 → 2 → 1)
		matched := false
		for _, length := range []int{4, 3, 2, 1} {
			if i+length <= len(s) {
				if h, ok := romajiTable[s[i:i+length]]; ok {
					result.WriteString(h)
					i += length
					matched = true
					break
				}
			}
		}
		if !matched {
			result.WriteByte(s[i])
			i++
		}
	}

	return result.String()
}

// HiraganaToKatakana converts every hiragana rune (U+3041–U+3096) to its
// katakana equivalent (U+30A1–U+30F6). Other characters are unchanged.
func HiraganaToKatakana(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 0x3041 && r <= 0x3096 {
			b.WriteRune(r + 0x60)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
