package nanoda

import (
	"encoding/json"
	"testing"
)

func order(n uint32) *uint32 {
	return &n
}

// VVMをまたいだキャラクターのマージとorder順の並びを確認する
func TestMergeCharacters(t *testing.T) {
	ms := []characterMeta{
		{
			Name:        "B",
			Styles:      []Style{{Id: 3, Name: "B_1", Type: STYLE_TYPE_TALK, Order: order(0)}},
			Version:     "0.0.0",
			SpeakerUuid: "f34ab151-c0f5-4e0a-9ad2-51ce30dba24d",
			Order:       order(1),
		},
		{
			Name:        "A",
			Styles:      []Style{{Id: 2, Name: "A_2", Type: STYLE_TYPE_TALK, Order: order(1)}},
			Version:     "0.0.0",
			SpeakerUuid: "a2e19d3d-1b2b-4b3d-9c2c-0b6d4f5c1d0e",
			Order:       order(0),
		},
		{
			Name:        "A",
			Styles:      []Style{{Id: 1, Name: "A_1", Type: STYLE_TYPE_TALK, Order: order(0)}},
			Version:     "0.0.0",
			SpeakerUuid: "a2e19d3d-1b2b-4b3d-9c2c-0b6d4f5c1d0e",
			Order:       order(0),
		},
		{
			Name:        "C",
			Styles:      []Style{{Id: 4, Name: "C_1", Type: STYLE_TYPE_TALK}},
			Version:     "0.0.0",
			SpeakerUuid: "c07a0e2c-1e26-4d4d-9a6f-6c8f6c6b8f9a",
		},
	}

	characters := mergeCharacters(ms)

	if len(characters) != 3 {
		t.Fatalf("expected 3 characters, got %d", len(characters))
	}
	// orderの昇順、orderが未設定のものは末尾
	names := []string{characters[0].Name, characters[1].Name, characters[2].Name}
	for i, expected := range []string{"A", "B", "C"} {
		if names[i] != expected {
			t.Errorf("characters[%d].Name: expected %q, got %q", i, expected, names[i])
		}
	}
	// 同一のキャラクターのスタイルはマージされ、orderの昇順に並ぶ
	if len(characters[0].Styles) != 2 {
		t.Fatalf("expected 2 styles, got %d", len(characters[0].Styles))
	}
	if characters[0].Styles[0].Id != 1 || characters[0].Styles[1].Id != 2 {
		t.Errorf("styles are not sorted by order: %v", characters[0].Styles)
	}
	if characters[0].Id != "a2e19d3d-1b2b-4b3d-9c2c-0b6d4f5c1d0e" {
		t.Errorf("unexpected character id: %q", characters[0].Id)
	}
}

// スタイルの一覧が「キャラクター名(スタイル名)」の形式になることを確認する
func TestToStyles(t *testing.T) {
	characters := []Character{
		{
			Name:   "ずんだもん",
			Styles: []Style{{Id: 3, Name: "ノーマル"}, {Id: 1, Name: "あまあま"}},
		},
	}
	styles := toStyles(characters)
	if len(styles) != 2 {
		t.Fatalf("expected 2 styles, got %d", len(styles))
	}
	if styles[0].Name != "ずんだもん(ノーマル)" {
		t.Errorf("unexpected style name: %q", styles[0].Name)
	}
	if styles[1].Id != 1 {
		t.Errorf("unexpected style id: %d", styles[1].Id)
	}
}

// typeが省略されているスタイルがtalkとして扱われることを確認する
func TestStyleUnmarshalDefaultsToTalk(t *testing.T) {
	var s Style
	if err := json.Unmarshal([]byte(`{"id":0,"name":"ノーマル","order":0}`), &s); err != nil {
		t.Fatal(err)
	}
	if s.Type != STYLE_TYPE_TALK {
		t.Errorf("expected %q, got %q", STYLE_TYPE_TALK, s.Type)
	}
	if s.Order == nil || *s.Order != 0 {
		t.Errorf("unexpected order: %v", s.Order)
	}

	var withoutOrder Style
	if err := json.Unmarshal([]byte(`{"id":0,"name":"ノーマル","type":"sing"}`), &withoutOrder); err != nil {
		t.Fatal(err)
	}
	if withoutOrder.Type != STYLE_TYPE_SING {
		t.Errorf("expected %q, got %q", STYLE_TYPE_SING, withoutOrder.Type)
	}
	if withoutOrder.Order != nil {
		t.Errorf("expected nil order, got %v", *withoutOrder.Order)
	}
}
