package voice

import "testing"

func TestSplitMoras(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"ずんだもん", []string{"ず", "ん", "だ", "も", "ん"}},
		{"きょうは", []string{"きょ", "う", "は"}},
		{"ぎゅっと", []string{"ぎゅ", "っ", "と"}},
		{"ファイル", []string{"ファ", "イ", "ル"}},
		{"コーヒー", []string{"コ", "ー", "ヒ", "ー"}},
		{"ーあ", []string{"あ"}}, // 先頭の長音は読み飛ばす
	}
	for _, tt := range tests {
		moras := splitMoras(tt.in)
		if len(moras) != len(tt.want) {
			t.Errorf("splitMoras(%q) = %d moras, want %d (%v)", tt.in, len(moras), len(tt.want), moraTexts(moras))
			continue
		}
		for i, w := range tt.want {
			if moras[i].Text != w {
				t.Errorf("splitMoras(%q)[%d].Text = %q, want %q", tt.in, i, moras[i].Text, w)
			}
		}
	}
}

func moraTexts(moras []Mora) []string {
	s := make([]string, len(moras))
	for i, m := range moras {
		s[i] = m.Text
	}
	return s
}

func TestSplitMorasConsonantAndVowel(t *testing.T) {
	moras := splitMoras("きょうん")
	if len(moras) != 3 {
		t.Fatalf("got %d moras (%v), want 3", len(moras), moraTexts(moras))
	}
	if moras[0].Consonant == nil || *moras[0].Consonant != "ky" {
		t.Errorf("きょ の子音が ky ではありません: %v", moras[0].Consonant)
	}
	if moras[0].Vowel != "o" {
		t.Errorf("きょ の母音 = %q, want %q", moras[0].Vowel, "o")
	}
	if moras[2].Vowel != "N" {
		t.Errorf("ん の母音 = %q, want %q", moras[2].Vowel, "N")
	}
}

func TestSplitPhrases(t *testing.T) {
	phrases, pauses := splitPhrases("こんにちは、なのだ。げんきです")
	want := []string{"こんにちは", "なのだ", "げんきです"}
	if len(phrases) != len(want) {
		t.Fatalf("got %v, want %v", phrases, want)
	}
	for i, w := range want {
		if phrases[i] != w {
			t.Errorf("phrases[%d] = %q, want %q", i, phrases[i], w)
		}
	}
	if !pauses[0] || !pauses[1] || pauses[2] {
		t.Errorf("pauses = %v, want [true true false]", pauses)
	}
}
