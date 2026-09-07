package voice

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
)

// demoEngine は VOICEVOX CORE が無くてもUIを動かせるようにするための
// オフライン実装。テキストを仮名として読み、簡単な正弦波合成で音を作る。
//
// あくまでUIの動作確認用で、VOICEVOXの音声を再現するものではない。
// 実際の音声が必要な場合は -core などでVOICEVOX COREを指定すること。
type demoEngine struct {
	speakers []Speaker
}

const demoSampleRate = 24000

func openDemo() Engine {
	names := []string{"デモボイス・アオ", "デモボイス・ミド", "デモボイス・アカ"}
	styles := []string{"ノーマル", "あまあま", "ツンツン"}
	e := &demoEngine{}
	var id StyleID
	for i, n := range names {
		sp := Speaker{Name: n, UUID: fmt.Sprintf("demo-%02d", i)}
		for _, s := range styles {
			sp.Styles = append(sp.Styles, Style{ID: id, Name: s})
			id++
		}
		e.speakers = append(e.speakers, sp)
	}
	return e
}

func (e *demoEngine) Name() string    { return "デモエンジン (VOICEVOX CORE なし)" }
func (e *demoEngine) Version() string { return "demo" }

func (e *demoEngine) Capabilities() Capabilities {
	return Capabilities{Speakers: true, Query: true, Accent: true}
}

func (e *demoEngine) Speakers() []Speaker { return e.speakers }

func (e *demoEngine) CreateQuery(text string, style StyleID) (*Query, error) {
	q := NewQuery()
	q.OutputSamplingRate = demoSampleRate
	texts, pauses := splitPhrases(text)
	for i, t := range texts {
		moras := splitMoras(t)
		if len(moras) == 0 {
			continue
		}
		ap := AccentPhrase{Moras: moras, Accent: defaultAccent(moras)}
		if pauses[i] {
			ap.PauseMora = &Mora{Text: "、", Vowel: "pau", VowelLength: 0.3}
		}
		q.AccentPhrases = append(q.AccentPhrases, ap)
	}
	if len(q.AccentPhrases) == 0 {
		return nil, fmt.Errorf("読み上げられる仮名がありません(デモエンジンは仮名のみ解析します)")
	}
	q.AccentPhrases = e.applyProsody(q.AccentPhrases, style)
	q.Kana = kanaOf(q.AccentPhrases)
	return q, nil
}

func (e *demoEngine) ReplaceAccentPhrases(phrases []AccentPhrase, style StyleID) ([]AccentPhrase, error) {
	return e.applyProsody(phrases, style), nil
}

// applyProsody はアクセント位置から音高と音素長を割り当てなおす。
// VOICEVOX COREの replace_mora_data に相当する処理の簡易版。
func (e *demoEngine) applyProsody(phrases []AccentPhrase, style StyleID) []AccentPhrase {
	base := 5.35 + 0.06*float64(style%9)
	out := make([]AccentPhrase, len(phrases))
	for i, ap := range phrases {
		c := ap.clone()
		for j := range c.Moras {
			m := &c.Moras[j]
			m.VowelLength = vowelLength(m.Vowel)
			if m.Consonant != nil {
				l := consonantLength(*m.Consonant)
				m.ConsonantLength = &l
			}
			m.Pitch = moraPitch(base, j, len(c.Moras), c.Accent, m.Vowel)
		}
		if c.IsInterrogative && len(c.Moras) > 0 {
			c.Moras[len(c.Moras)-1].Pitch += 0.3
		}
		out[i] = c
	}
	return out
}

// defaultAccent はアクセント核の初期位置を返す。日本語の解析は行わない
// ので、VOICEVOX同様あとからUIで直せることを前提に1を既定とする。
func defaultAccent(moras []Mora) int {
	if len(moras) == 0 {
		return 0
	}
	return 1
}

func moraPitch(base float64, index, count, accent int, vowel string) float64 {
	if vowel == "cl" || vowel == "pau" {
		return 0
	}
	high := base + 0.22
	low := base - 0.04
	var p float64
	switch {
	case accent == 1:
		if index == 0 {
			p = high
		} else {
			p = low
		}
	case accent <= 0: // 平板型
		if index == 0 {
			p = low
		} else {
			p = high
		}
	default:
		switch {
		case index == 0:
			p = low
		case index < accent:
			p = high
		default:
			p = low
		}
	}
	// 語尾に向けてわずかに下げると自然に聞こえる。
	if count > 1 {
		p -= 0.05 * float64(index) / float64(count-1)
	}
	return p
}

func vowelLength(vowel string) float64 {
	switch vowel {
	case "cl":
		return 0.07
	case "N":
		return 0.09
	case "pau":
		return 0.3
	case "a", "o":
		return 0.12
	default:
		return 0.105
	}
}

func consonantLength(consonant string) float64 {
	switch consonant {
	case "s", "sh", "h", "f", "z", "j", "ts", "ch":
		return 0.075
	case "n", "m", "ny", "my", "N":
		return 0.055
	default:
		return 0.045
	}
}

func kanaOf(phrases []AccentPhrase) string {
	var sb strings.Builder
	for i, ap := range phrases {
		if i > 0 {
			sb.WriteString("/")
		}
		for _, m := range ap.Moras {
			sb.WriteString(m.Text)
		}
		if ap.Accent > 0 {
			fmt.Fprintf(&sb, "'%d", ap.Accent)
		}
		if ap.PauseMora != nil {
			sb.WriteString("、")
		}
	}
	return sb.String()
}

func (e *demoEngine) SynthesizeText(text string, style StyleID) ([]byte, error) {
	q, err := e.CreateQuery(text, style)
	if err != nil {
		return nil, err
	}
	return e.Synthesize(q, style)
}

// vowelHarmonics は母音ごとの倍音の強さ。フォルマントの真似事で、
// 母音の区別がなんとなく付く程度のもの。
var vowelHarmonics = map[string][]float64{
	"a": {1, 0.70, 0.50, 0.35, 0.20, 0.12, 0.08, 0.05},
	"i": {1, 0.30, 0.15, 0.45, 0.35, 0.12, 0.06, 0.03},
	"u": {1, 0.50, 0.20, 0.12, 0.08, 0.05, 0.03, 0.02},
	"e": {1, 0.60, 0.40, 0.30, 0.25, 0.10, 0.05, 0.03},
	"o": {1, 0.75, 0.35, 0.18, 0.10, 0.06, 0.04, 0.02},
	"N": {1, 0.40, 0.15, 0.05, 0.03, 0.02, 0.01, 0.01},
}

func (e *demoEngine) Synthesize(q *Query, style StyleID) ([]byte, error) {
	if q == nil {
		return nil, fmt.Errorf("クエリがありません")
	}
	rate := q.OutputSamplingRate
	if rate <= 0 {
		rate = demoSampleRate
	}
	speed := q.SpeedScale
	if speed <= 0 {
		speed = 1
	}

	mean := meanPitch(q)
	var out []float64
	silence := func(sec float64) {
		n := int(sec / speed * float64(rate))
		for i := 0; i < n; i++ {
			out = append(out, 0)
		}
	}
	noise := rand.New(rand.NewPCG(0x766f6963, 0x65766f78))

	silence(q.PrePhonemeLength)
	phase := 0.0
	for _, ap := range q.AccentPhrases {
		for _, m := range ap.Moras {
			if m.Consonant != nil && m.ConsonantLength != nil {
				n := int(*m.ConsonantLength / speed * float64(rate))
				amp := 0.06 * q.VolumeScale
				for i := 0; i < n; i++ {
					out = append(out, (noise.Float64()*2-1)*amp*consonantEnvelope(i, n))
				}
			}
			if m.Pitch <= 0 {
				silence(m.VowelLength)
				continue
			}
			pitch := mean + (m.Pitch-mean)*q.IntonationScale + q.PitchScale
			f0 := math.Exp(pitch)
			harmonics, ok := vowelHarmonics[m.Vowel]
			if !ok {
				harmonics = vowelHarmonics["a"]
			}
			n := int(m.VowelLength / speed * float64(rate))
			for i := 0; i < n; i++ {
				var v float64
				for k, h := range harmonics {
					v += h * math.Sin(phase*float64(k+1))
				}
				v /= 4
				out = append(out, v*0.5*q.VolumeScale*vowelEnvelope(i, n))
				phase += 2 * math.Pi * f0 / float64(rate)
			}
			phase = math.Mod(phase, 2*math.Pi)
		}
		if ap.PauseMora != nil {
			silence(ap.PauseMora.VowelLength)
		}
	}
	silence(q.PostPhonemeLength)

	return encodeWAV(out, rate), nil
}

func meanPitch(q *Query) float64 {
	var sum float64
	var n int
	for _, ap := range q.AccentPhrases {
		for _, m := range ap.Moras {
			if m.Pitch > 0 {
				sum += m.Pitch
				n++
			}
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func vowelEnvelope(i, n int) float64 {
	if n <= 1 {
		return 0
	}
	attack := min(n/6+1, n)
	release := min(n/3+1, n)
	switch {
	case i < attack:
		return float64(i) / float64(attack)
	case i >= n-release:
		return float64(n-i) / float64(release)
	default:
		return 1
	}
}

func consonantEnvelope(i, n int) float64 {
	if n <= 1 {
		return 0
	}
	return 1 - float64(i)/float64(n)
}

func (e *demoEngine) Close() error { return nil }
