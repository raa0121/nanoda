// Package voice はVOICEVOX COREを叩くエンジンの抽象を提供する。
//
// VOICEVOX COREのバージョンによってnanodaが公開しているAPIが異なる
// (0.15系はAudioQueryまで扱えるが、0.16系はまだTtsのみ)ため、
// UIから見た機能差は Capabilities として表現し、実装は
// nanoda015 / nanoda016 / demo の3種類を用意している。
package voice

import "errors"

// ErrUnsupported はエンジンがその操作に対応していないことを表す。
var ErrUnsupported = errors.New("この機能には対応していません")

// StyleID はスタイルのID。VOICEVOX COREのVoicevoxStyleIdと同じ。
type StyleID uint32

// Style はキャラクターの一つのスタイル。
type Style struct {
	ID   StyleID
	Name string
}

// Speaker はキャラクター単位のメタ情報。
type Speaker struct {
	Name   string
	UUID   string
	Styles []Style
}

// Mora はモーラ(子音＋母音)。VOICEVOX COREのJSON表現と同じ構造を持つ。
type Mora struct {
	Text            string   `json:"text"`
	Consonant       *string  `json:"consonant"`
	ConsonantLength *float64 `json:"consonant_length"`
	Vowel           string   `json:"vowel"`
	VowelLength     float64  `json:"vowel_length"`
	Pitch           float64  `json:"pitch"`
}

// AccentPhrase はアクセント句。
type AccentPhrase struct {
	Moras           []Mora `json:"moras"`
	Accent          int    `json:"accent"`
	PauseMora       *Mora  `json:"pause_mora"`
	IsInterrogative bool   `json:"is_interrogative"`
}

// Query は音声合成用のクエリ(VOICEVOXのAudioQueryに相当)。
type Query struct {
	AccentPhrases      []AccentPhrase `json:"accent_phrases"`
	SpeedScale         float64        `json:"speed_scale"`
	PitchScale         float64        `json:"pitch_scale"`
	IntonationScale    float64        `json:"intonation_scale"`
	VolumeScale        float64        `json:"volume_scale"`
	PrePhonemeLength   float64        `json:"pre_phoneme_length"`
	PostPhonemeLength  float64        `json:"post_phoneme_length"`
	OutputSamplingRate int            `json:"output_sampling_rate"`
	OutputStereo       bool           `json:"output_stereo"`
	Kana               string         `json:"kana"`
}

// NewQuery は各パラメータをVOICEVOXの既定値にしたクエリを返す。
func NewQuery() *Query {
	return &Query{
		SpeedScale:         1,
		PitchScale:         0,
		IntonationScale:    1,
		VolumeScale:        1,
		PrePhonemeLength:   0.1,
		PostPhonemeLength:  0.1,
		OutputSamplingRate: 24000,
	}
}

// Clone はクエリの深いコピーを返す。UI側で編集してもエンジンが返した
// 値を壊さないようにするために使う。
func (q *Query) Clone() *Query {
	if q == nil {
		return nil
	}
	c := *q
	c.AccentPhrases = make([]AccentPhrase, len(q.AccentPhrases))
	for i, ap := range q.AccentPhrases {
		c.AccentPhrases[i] = ap.clone()
	}
	return &c
}

func (a AccentPhrase) clone() AccentPhrase {
	c := a
	c.Moras = append([]Mora(nil), a.Moras...)
	if a.PauseMora != nil {
		m := *a.PauseMora
		c.PauseMora = &m
	}
	return c
}

// Text はクエリが読み上げる内容をモーラから組み立てて返す。
func (q *Query) Text() string {
	if q == nil {
		return ""
	}
	var s string
	for _, ap := range q.AccentPhrases {
		for _, m := range ap.Moras {
			s += m.Text
		}
		if ap.PauseMora != nil {
			s += "、"
		}
	}
	return s
}

// Duration はクエリが表す音声のおおよその長さを秒で返す。
func (q *Query) Duration() float64 {
	if q == nil {
		return 0
	}
	d := q.PrePhonemeLength + q.PostPhonemeLength
	for _, ap := range q.AccentPhrases {
		for _, m := range ap.Moras {
			d += m.length()
		}
		if ap.PauseMora != nil {
			d += ap.PauseMora.length()
		}
	}
	if q.SpeedScale > 0 {
		d /= q.SpeedScale
	}
	return d
}

func (m Mora) length() float64 {
	d := m.VowelLength
	if m.ConsonantLength != nil {
		d += *m.ConsonantLength
	}
	return d
}

// Capabilities はエンジンが提供できる機能を表す。
// VOICEVOX COREのバージョンとnanodaの対応状況で変わる。
type Capabilities struct {
	// Speakers はキャラクター一覧を取得できるかどうか。
	Speakers bool
	// Query はAudioQueryを作って各種パラメータを編集できるかどうか。
	Query bool
	// Accent はアクセント句を編集し、音高・音素長を再生成できるかどうか。
	Accent bool
}

// Engine は音声合成エンジン。実装はどれもゴルーチンセーフではないので、
// 呼び出し側で直列化すること(app.Worker が担当している)。
type Engine interface {
	// Name はエンジンの表示名。
	Name() string
	// Version は読み込んだVOICEVOX COREのバージョン。
	Version() string
	// Capabilities は利用できる機能。
	Capabilities() Capabilities
	// Speakers はキャラクター一覧。Capabilities().Speakers が false なら空。
	Speakers() []Speaker
	// CreateQuery はテキストからAudioQueryを生成する。
	CreateQuery(text string, style StyleID) (*Query, error)
	// ReplaceAccentPhrases はアクセント句の音高と音素長を生成しなおす。
	ReplaceAccentPhrases(phrases []AccentPhrase, style StyleID) ([]AccentPhrase, error)
	// Synthesize はAudioQueryからWAVを合成する。
	Synthesize(q *Query, style StyleID) ([]byte, error)
	// SynthesizeText はテキストから直接WAVを合成する(AudioQueryを経由しない)。
	SynthesizeText(text string, style StyleID) ([]byte, error)
	// Close はエンジンが確保したネイティブ資源を解放する。
	Close() error
}

// FirstStyleID は最初のスタイルのIDを返す。キャラクターが一人もいない
// 場合は0を返す。
func FirstStyleID(speakers []Speaker) StyleID {
	for _, sp := range speakers {
		if len(sp.Styles) > 0 {
			return sp.Styles[0].ID
		}
	}
	return 0
}

// StyleName は "ずんだもん(あまあま)" 形式の表示名を返す。
// 該当するスタイルが無い場合は "スタイル 3" のようにIDを返す。
func StyleName(speakers []Speaker, id StyleID) string {
	for _, sp := range speakers {
		for _, st := range sp.Styles {
			if st.ID == id {
				return sp.Name + "(" + st.Name + ")"
			}
		}
	}
	return "スタイル " + itoa(uint32(id))
}

func itoa(v uint32) string {
	if v == 0 {
		return "0"
	}
	var buf [10]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
