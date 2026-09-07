package voice

import "strings"

// kanaBase は仮名1文字と、それが表す子音・母音(VOICEVOXの音素名)の対応。
// デモエンジンがテキストをモーラに切るために使う簡易テーブルで、
// OpenJTalkの解析を置き換えるものではない。
var kanaBase = map[rune][2]string{
	'あ': {"", "a"}, 'い': {"", "i"}, 'う': {"", "u"}, 'え': {"", "e"}, 'お': {"", "o"},
	'か': {"k", "a"}, 'き': {"k", "i"}, 'く': {"k", "u"}, 'け': {"k", "e"}, 'こ': {"k", "o"},
	'が': {"g", "a"}, 'ぎ': {"g", "i"}, 'ぐ': {"g", "u"}, 'げ': {"g", "e"}, 'ご': {"g", "o"},
	'さ': {"s", "a"}, 'し': {"sh", "i"}, 'す': {"s", "u"}, 'せ': {"s", "e"}, 'そ': {"s", "o"},
	'ざ': {"z", "a"}, 'じ': {"j", "i"}, 'ず': {"z", "u"}, 'ぜ': {"z", "e"}, 'ぞ': {"z", "o"},
	'た': {"t", "a"}, 'ち': {"ch", "i"}, 'つ': {"ts", "u"}, 'て': {"t", "e"}, 'と': {"t", "o"},
	'だ': {"d", "a"}, 'ぢ': {"j", "i"}, 'づ': {"z", "u"}, 'で': {"d", "e"}, 'ど': {"d", "o"},
	'な': {"n", "a"}, 'に': {"n", "i"}, 'ぬ': {"n", "u"}, 'ね': {"n", "e"}, 'の': {"n", "o"},
	'は': {"h", "a"}, 'ひ': {"h", "i"}, 'ふ': {"f", "u"}, 'へ': {"h", "e"}, 'ほ': {"h", "o"},
	'ば': {"b", "a"}, 'び': {"b", "i"}, 'ぶ': {"b", "u"}, 'べ': {"b", "e"}, 'ぼ': {"b", "o"},
	'ぱ': {"p", "a"}, 'ぴ': {"p", "i"}, 'ぷ': {"p", "u"}, 'ぺ': {"p", "e"}, 'ぽ': {"p", "o"},
	'ま': {"m", "a"}, 'み': {"m", "i"}, 'む': {"m", "u"}, 'め': {"m", "e"}, 'も': {"m", "o"},
	'や': {"y", "a"}, 'ゆ': {"y", "u"}, 'よ': {"y", "o"},
	'ら': {"r", "a"}, 'り': {"r", "i"}, 'る': {"r", "u"}, 'れ': {"r", "e"}, 'ろ': {"r", "o"},
	'わ': {"w", "a"}, 'ゐ': {"w", "i"}, 'ゑ': {"w", "e"}, 'を': {"", "o"},
	'ん': {"", "N"}, 'っ': {"", "cl"}, 'ー': {"", "long"},
	'ゔ': {"v", "u"},
}

// smallVowel は小書き仮名が表す母音。
var smallVowel = map[rune]string{
	'ぁ': "a", 'ぃ': "i", 'ぅ': "u", 'ぇ': "e", 'ぉ': "o",
	'ゃ': "a", 'ゅ': "u", 'ょ': "o",
}

// palatalized は拗音になったときの子音。sh/ch/j は変化しない。
var palatalized = map[string]string{
	"k": "ky", "g": "gy", "n": "ny", "h": "hy", "b": "by", "p": "py",
	"m": "my", "r": "ry", "s": "sh", "z": "j", "t": "ch", "d": "j",
	"sh": "sh", "ch": "ch", "j": "j",
}

// phraseBreaks はアクセント句を切る文字。「、」「。」はポーズにもなる。
const phraseBreaks = "、。，．,.!?！？「」『』（）()・…‥　 \t\n"

func isPause(r rune) bool { return r == '、' || r == '。' || r == ',' || r == '.' }

func katakanaToHiragana(r rune) rune {
	if r >= 'ァ' && r <= 'ヶ' {
		return r - 0x60
	}
	return r
}

// splitMoras はテキストを仮名の並びとみなしてモーラに切る。
// 解析できない文字は読み飛ばす。
func splitMoras(text string) []Mora {
	rs := []rune(text)
	var moras []Mora
	for i := 0; i < len(rs); i++ {
		r := katakanaToHiragana(rs[i])
		base, ok := kanaBase[r]
		if !ok {
			continue
		}
		consonant, vowel := base[0], base[1]
		display := string(rs[i])

		switch vowel {
		case "long":
			// 長音は直前の母音を伸ばす。先頭に来た場合は読み飛ばす。
			if len(moras) == 0 {
				continue
			}
			prev := moras[len(moras)-1]
			moras = append(moras, Mora{Text: display, Vowel: prev.Vowel})
			continue
		case "cl", "N":
			moras = append(moras, Mora{Text: display, Vowel: vowel})
			continue
		}

		// 次が小書き仮名なら拗音・外来音として1モーラにまとめる。
		if i+1 < len(rs) {
			if sv, ok := smallVowel[katakanaToHiragana(rs[i+1])]; ok {
				if p, ok := palatalized[consonant]; ok && isYouon(rs[i+1]) {
					consonant = p
				}
				vowel = sv
				display += string(rs[i+1])
				i++
			}
		}
		m := Mora{Text: display, Vowel: vowel}
		if consonant != "" {
			c := consonant
			m.Consonant = &c
		}
		moras = append(moras, m)
	}
	return moras
}

func isYouon(r rune) bool {
	r = katakanaToHiragana(r)
	return r == 'ゃ' || r == 'ゅ' || r == 'ょ'
}

// splitPhrases はテキストを句読点などで区切り、アクセント句ごとの
// 文字列と、その後ろにポーズを入れるかどうかを返す。
func splitPhrases(text string) (phrases []string, pauses []bool) {
	var cur strings.Builder
	flush := func(pause bool) {
		if cur.Len() > 0 {
			phrases = append(phrases, cur.String())
			pauses = append(pauses, pause)
			cur.Reset()
		} else if pause && len(pauses) > 0 {
			pauses[len(pauses)-1] = true
		}
	}
	for _, r := range text {
		if strings.ContainsRune(phraseBreaks, r) {
			flush(isPause(r))
			continue
		}
		cur.WriteRune(r)
	}
	flush(false)
	return
}
