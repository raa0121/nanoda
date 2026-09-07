package nanoda

import (
	"encoding/json"
	"math"
	"sort"
)

// スタイルのID
type StyleId uint32

// キャラクターのUUID
type CharacterId string

// 話者のID
//
// Deprecated: VOICEVOX CORE 0.16で「話者」は「キャラクター」に改称された。CharacterIdを使うこと。
type SpeakerId = CharacterId

// スタイルの種類
type StyleType string

const (
	STYLE_TYPE_TALK            StyleType = "talk"            // 音声合成
	STYLE_TYPE_SINGING_TEACHER StyleType = "singing_teacher" // 歌唱音声合成用のクエリの生成
	STYLE_TYPE_FRAME_DECODE    StyleType = "frame_decode"    // 歌唱音声合成
	STYLE_TYPE_SING            StyleType = "sing"            // 歌唱音声合成用のクエリの生成と歌唱音声合成
)

// スタイル情報
type Style struct {
	Id    StyleId   `json:"id"`
	Name  string    `json:"name"`
	Type  StyleType `json:"type"`
	Order *uint32   `json:"order"` // 全体でのソート順 未設定の場合はnil
}

// typeが省略された場合はtalkとして扱う
func (s *Style) UnmarshalJSON(b []byte) (err error) {
	type style Style
	tmp := style{Type: STYLE_TYPE_TALK}
	if err = json.Unmarshal(b, &tmp); err != nil {
		return
	}
	if tmp.Type == "" {
		tmp.Type = STYLE_TYPE_TALK
	}
	*s = Style(tmp)
	return
}

// 内部用のメタ情報 VVMごとのキャラクター単位のメタ情報に対応する
type characterMeta struct {
	Name        string      `json:"name"`
	Styles      []Style     `json:"styles"`
	Version     string      `json:"version"`
	SpeakerUuid CharacterId `json:"speaker_uuid"`
	Order       *uint32     `json:"order"`
}

// キャラクター単位のメタ情報
type Character struct {
	Id      CharacterId // キャラクターのUUID
	Name    string      // キャラクター名
	Styles  []Style     // スタイルの一覧
	Version string      // キャラクターのバージョン
	Order   *uint32     // 全体でのソート順 未設定の場合はnil
}

// キャラクター単位のメタ情報
//
// Deprecated: VOICEVOX CORE 0.16で「話者」は「キャラクター」に改称された。Characterを使うこと。
type Meta struct {
	Name      string
	Styles    []Style
	SpeakerId SpeakerId
}

// 読み込んでいるすべてのVVMのメタ情報を取得する
func (v *Voicevox) getCharacterMetas() (m []characterMeta) {
	for _, vvm := range v.vvms {
		m = append(m, vvm.metas...)
	}
	return
}

/*
VVMごとのメタ情報をキャラクター単位にマージする
マージ後の並び順はVOICEVOX CORE本体の実装に合わせ、orderの昇順(未設定のものは末尾)となる
*/
func mergeCharacters(ms []characterMeta) []Character {
	characters := []Character{}
	indexes := map[CharacterId]int{}
	for _, m := range ms {
		if i, ok := indexes[m.SpeakerUuid]; ok {
			characters[i].Styles = append(characters[i].Styles, m.Styles...)
			continue
		}
		indexes[m.SpeakerUuid] = len(characters)
		characters = append(characters, Character{
			Id:      m.SpeakerUuid,
			Name:    m.Name,
			Styles:  append([]Style{}, m.Styles...),
			Version: m.Version,
			Order:   m.Order,
		})
	}
	for i := range characters {
		styles := characters[i].Styles
		sort.SliceStable(styles, func(a, b int) bool {
			return orderKey(styles[a].Order) < orderKey(styles[b].Order)
		})
	}
	sort.SliceStable(characters, func(i, j int) bool {
		return orderKey(characters[i].Order) < orderKey(characters[j].Order)
	})
	return characters
}

// orderが未設定のものが末尾に並ぶようにするためのソートキー
func orderKey(order *uint32) uint64 {
	if order == nil {
		return uint64(math.MaxUint32) + 1
	}
	return uint64(*order)
}

// キャラクター単位のメタ情報を0.15以前の形式に変換する
func toMetas(characters []Character) []Meta {
	metas := make([]Meta, 0, len(characters))
	for _, c := range characters {
		metas = append(metas, Meta{
			Name:      c.Name,
			Styles:    c.Styles,
			SpeakerId: c.Id,
		})
	}
	return metas
}

// キャラクターの一覧からスタイルの一覧を作る
func toStyles(characters []Character) []Style {
	styles := []Style{}
	for _, c := range characters {
		for _, style := range c.Styles {
			style.Name = c.Name + "(" + style.Name + ")"
			styles = append(styles, style)
		}
	}
	return styles
}
