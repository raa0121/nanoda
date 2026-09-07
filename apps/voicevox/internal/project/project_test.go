package project

import (
	"path/filepath"
	"testing"

	"github.com/raa0121/nanoda/apps/voicevox/internal/voice"
)

func TestInsertRemoveMove(t *testing.T) {
	p := New(3)
	if p.Count() != 1 {
		t.Fatalf("新規プロジェクトの行数 = %d, want 1", p.Count())
	}
	p.SetText(0, "いち")
	p.Append(3)
	p.SetText(1, "に")
	p.Append(3)
	p.SetText(2, "さん")

	if got := texts(p); !equal(got, []string{"いち", "に", "さん"}) {
		t.Fatalf("texts = %v", got)
	}

	if got := p.Move(0, 2); got != 2 {
		t.Errorf("Move returned %d, want 2", got)
	}
	if got := texts(p); !equal(got, []string{"に", "さん", "いち"}) {
		t.Errorf("Move後 = %v", got)
	}

	if got := p.Remove(1); got != 1 {
		t.Errorf("Remove returned %d, want 1", got)
	}
	if got := texts(p); !equal(got, []string{"に", "いち"}) {
		t.Errorf("Remove後 = %v", got)
	}

	// 末尾を消したときは手前を選ぶ。
	if got := p.Remove(1); got != 0 {
		t.Errorf("末尾のRemove returned %d, want 0", got)
	}
	// 最後の1行は消さずに空にする。
	if got := p.Remove(0); got != 0 || p.Count() != 1 || p.At(0).Text != "" {
		t.Errorf("最後の1行のRemove: index=%d count=%d text=%q", got, p.Count(), p.At(0).Text)
	}
}

func TestSetTextMarksStale(t *testing.T) {
	p := New(0)
	p.SetQuery(0, voice.NewQuery())
	if p.At(0).NeedsQuery() {
		t.Error("クエリを入れた直後にNeedsQueryがtrueです")
	}
	p.SetText(0, "あ")
	if !p.At(0).NeedsQuery() {
		t.Error("テキストを変えてもNeedsQueryがfalseのままです")
	}
	p.SetQuery(0, voice.NewQuery())
	p.SetStyle(0, 7)
	if !p.At(0).NeedsQuery() {
		t.Error("スタイルを変えてもNeedsQueryがfalseのままです")
	}
}

func TestSaveLoad(t *testing.T) {
	p := New(3)
	p.SetText(0, "ずんだもんなのだ")
	q := voice.NewQuery()
	q.SpeedScale = 1.5
	q.AccentPhrases = []voice.AccentPhrase{{
		Moras:  []voice.Mora{{Text: "ず", Vowel: "u", VowelLength: 0.1, Pitch: 5.6}},
		Accent: 1,
	}}
	p.SetQuery(0, q)

	path := filepath.Join(t.TempDir(), "test")
	if err := p.Save(path); err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(p.Path) != Ext {
		t.Errorf("拡張子が補われていません: %s", p.Path)
	}
	if p.IsDirty() {
		t.Error("保存後もdirtyのままです")
	}

	loaded, err := Load(p.Path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Count() != 1 {
		t.Fatalf("行数 = %d", loaded.Count())
	}
	l := loaded.At(0)
	if l.Text != "ずんだもんなのだ" || l.Style != 3 {
		t.Errorf("行が復元できていません: %+v", l)
	}
	if l.Query == nil || l.Query.SpeedScale != 1.5 || len(l.Query.AccentPhrases) != 1 {
		t.Errorf("クエリが復元できていません: %+v", l.Query)
	}
	if l.Query.AccentPhrases[0].Moras[0].Text != "ず" {
		t.Errorf("モーラが復元できていません: %+v", l.Query.AccentPhrases[0])
	}
}

func TestAtOutOfRange(t *testing.T) {
	p := New(0)
	if p.At(-1) != nil || p.At(1) != nil {
		t.Error("範囲外でnilが返りません")
	}
}

func texts(p *Project) []string {
	s := make([]string, p.Count())
	for i := range s {
		s[i] = p.At(i).Text
	}
	return s
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
