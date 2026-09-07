package voice

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestDemoEngineCreateQuery(t *testing.T) {
	e := openDemo()
	q, err := e.CreateQuery("こんにちは、なのだ", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(q.AccentPhrases) != 2 {
		t.Fatalf("アクセント句が %d 個です。2個を期待しています", len(q.AccentPhrases))
	}
	if q.AccentPhrases[0].PauseMora == nil {
		t.Error("「、」の後ろにポーズが入っていません")
	}
	if q.SpeedScale != 1 || q.IntonationScale != 1 || q.VolumeScale != 1 {
		t.Errorf("既定のパラメータが1ではありません: %+v", q)
	}
	for _, m := range q.AccentPhrases[0].Moras {
		if m.VowelLength <= 0 {
			t.Errorf("モーラ %q の母音長が 0 以下です", m.Text)
		}
	}
}

func TestDemoEngineCreateQueryWithoutKana(t *testing.T) {
	e := openDemo()
	if _, err := e.CreateQuery("!!!", 0); err == nil {
		t.Error("仮名が無いテキストでエラーになりませんでした")
	}
}

// アクセント核を動かすと音高の並びが変わることを確かめる。
// VOICEVOXでアクセントをずらしたときと同じ振る舞い。
func TestDemoEngineAccentChangesPitch(t *testing.T) {
	e := openDemo()
	q, err := e.CreateQuery("なのだ", 0)
	if err != nil {
		t.Fatal(err)
	}
	head := q.AccentPhrases[0].Moras
	if !(head[0].Pitch > head[1].Pitch) {
		t.Errorf("頭高型なのに1モーラ目が高くありません: %v", pitches(head))
	}

	q.AccentPhrases[0].Accent = 2
	replaced, err := e.ReplaceAccentPhrases(q.AccentPhrases, 0)
	if err != nil {
		t.Fatal(err)
	}
	mid := replaced[0].Moras
	if !(mid[1].Pitch > mid[0].Pitch && mid[1].Pitch > mid[2].Pitch) {
		t.Errorf("中高型になっていません: %v", pitches(mid))
	}
}

func pitches(moras []Mora) []float64 {
	p := make([]float64, len(moras))
	for i, m := range moras {
		p[i] = m.Pitch
	}
	return p
}

func TestDemoEngineSynthesizeWAV(t *testing.T) {
	e := openDemo()
	q, err := e.CreateQuery("ずんだもんなのだ", 3)
	if err != nil {
		t.Fatal(err)
	}
	b, err := e.Synthesize(q, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(b, []byte("RIFF")) || !bytes.Equal(b[8:12], []byte("WAVE")) {
		t.Fatalf("RIFF WAVEになっていません: %x", b[:min(16, len(b))])
	}
	channels := binary.LittleEndian.Uint16(b[22:24])
	rate := binary.LittleEndian.Uint32(b[24:28])
	bits := binary.LittleEndian.Uint16(b[34:36])
	if channels != 1 || rate != demoSampleRate || bits != 16 {
		t.Errorf("フォーマットが %dch %dHz %dbit です。1ch %dHz 16bit を期待しています", channels, rate, bits, demoSampleRate)
	}
	dataSize := binary.LittleEndian.Uint32(b[40:44])
	if int(dataSize)+44 != len(b) {
		t.Errorf("dataチャンクの長さ %d がファイル長 %d と合いません", dataSize, len(b))
	}
	// 話速を倍にすると短くなる。
	q.SpeedScale = 2
	fast, err := e.Synthesize(q, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(fast) >= len(b) {
		t.Errorf("話速2.0のほうが長くなっています (%d >= %d)", len(fast), len(b))
	}
}

func TestQueryClone(t *testing.T) {
	e := openDemo()
	q, err := e.CreateQuery("なのだ", 0)
	if err != nil {
		t.Fatal(err)
	}
	c := q.Clone()
	c.AccentPhrases[0].Moras[0].Pitch = 99
	c.SpeedScale = 9
	if q.AccentPhrases[0].Moras[0].Pitch == 99 || q.SpeedScale == 9 {
		t.Error("Cloneが元のクエリを共有しています")
	}
}

func TestStyleName(t *testing.T) {
	e := openDemo()
	sp := e.Speakers()
	if got, want := StyleName(sp, sp[0].Styles[1].ID), sp[0].Name+"("+sp[0].Styles[1].Name+")"; got != want {
		t.Errorf("StyleName = %q, want %q", got, want)
	}
	if got := StyleName(sp, 9999); got != "スタイル 9999" {
		t.Errorf("未知のスタイル = %q", got)
	}
}
