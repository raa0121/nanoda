//go:build !windows

package voice

import (
	"fmt"
	"sync"

	nanoda "github.com/aethiopicuschan/nanoda"
)

// core015 は VOICEVOX CORE 0.15系 を nanoda (モジュールパスにv2が付かない方)
// 経由で扱う実装。AudioQuery・アクセント句・メタ情報まで揃っているので、
// VOICEVOXエディタとほぼ同じことができる。
type core015 struct {
	v        *nanoda.Voicevox
	s        nanoda.Synthesizer
	version  string
	speakers []Speaker

	// VOICEVOX本体と同じく、必要になったスタイルのモデルだけを読み込む。
	// 全部を先に読むと起動が目に見えて遅くなるため。
	loadedMu sync.Mutex
	loaded   map[StyleID]bool
}

func open015(cfg Config) (Engine, error) {
	v, err := nanoda.NewVoicevox(cfg.CorePath, cfg.DictPath, cfg.ModelPath)
	if err != nil {
		return nil, err
	}
	s, err := v.NewSynthesizer()
	if err != nil {
		return nil, err
	}
	c := &core015{
		v:       v,
		s:       s,
		version: v.GetVersion(),
		loaded:  map[StyleID]bool{},
	}
	for _, m := range v.GetMetas() {
		sp := Speaker{Name: m.Name, UUID: string(m.SpeakerId)}
		for _, st := range m.Styles {
			sp.Styles = append(sp.Styles, Style{ID: StyleID(st.Id), Name: st.Name})
		}
		c.speakers = append(c.speakers, sp)
	}
	return c, nil
}

func (c *core015) Name() string    { return "nanoda (VOICEVOX CORE 0.15系)" }
func (c *core015) Version() string { return c.version }

func (c *core015) Capabilities() Capabilities {
	return Capabilities{Speakers: true, Query: true, Accent: true}
}

func (c *core015) Speakers() []Speaker { return c.speakers }

func (c *core015) ensureLoaded(style StyleID) error {
	c.loadedMu.Lock()
	defer c.loadedMu.Unlock()
	if c.loaded[style] {
		return nil
	}
	if err := c.s.LoadModelsFromStyleId(nanoda.StyleId(style)); err != nil {
		return fmt.Errorf("スタイル %d のモデルを読み込めませんでした: %w", style, err)
	}
	c.loaded[style] = true
	return nil
}

func (c *core015) CreateQuery(text string, style StyleID) (*Query, error) {
	if err := c.ensureLoaded(style); err != nil {
		return nil, err
	}
	aq, err := c.s.CreateAudioQuery(text, nanoda.StyleId(style))
	if err != nil {
		return nil, err
	}
	return fromNanodaQuery(aq), nil
}

func (c *core015) ReplaceAccentPhrases(phrases []AccentPhrase, style StyleID) ([]AccentPhrase, error) {
	if err := c.ensureLoaded(style); err != nil {
		return nil, err
	}
	res, err := c.s.Replace(toNanodaAccentPhrases(phrases), nanoda.StyleId(style))
	if err != nil {
		return nil, err
	}
	return fromNanodaAccentPhrases(res), nil
}

func (c *core015) Synthesize(q *Query, style StyleID) ([]byte, error) {
	if q == nil {
		return nil, fmt.Errorf("クエリがありません")
	}
	if err := c.ensureLoaded(style); err != nil {
		return nil, err
	}
	r, err := c.s.Synthesis(toNanodaQuery(q), nanoda.StyleId(style))
	if err != nil {
		return nil, err
	}
	return readAllAndClose(r)
}

func (c *core015) SynthesizeText(text string, style StyleID) ([]byte, error) {
	if err := c.ensureLoaded(style); err != nil {
		return nil, err
	}
	r, err := c.s.Tts(text, nanoda.StyleId(style))
	if err != nil {
		return nil, err
	}
	return readAllAndClose(r)
}

func (c *core015) Close() error {
	c.s.Close()
	return nil
}

func fromNanodaQuery(aq nanoda.AudioQuery) *Query {
	return &Query{
		AccentPhrases:      fromNanodaAccentPhrases(aq.AccentPhrases),
		SpeedScale:         aq.SpeedScale,
		PitchScale:         aq.PitchScale,
		IntonationScale:    aq.IntonationScale,
		VolumeScale:        aq.VolumeScale,
		PrePhonemeLength:   aq.PrePhonemeLength,
		PostPhonemeLength:  aq.PostPhonemeLength,
		OutputSamplingRate: aq.OutputSamplingRate,
		OutputStereo:       aq.OutputStereo,
		Kana:               aq.Kana,
	}
}

func toNanodaQuery(q *Query) nanoda.AudioQuery {
	return nanoda.AudioQuery{
		AccentPhrases:      toNanodaAccentPhrases(q.AccentPhrases),
		SpeedScale:         q.SpeedScale,
		PitchScale:         q.PitchScale,
		IntonationScale:    q.IntonationScale,
		VolumeScale:        q.VolumeScale,
		PrePhonemeLength:   q.PrePhonemeLength,
		PostPhonemeLength:  q.PostPhonemeLength,
		OutputSamplingRate: q.OutputSamplingRate,
		OutputStereo:       q.OutputStereo,
		Kana:               q.Kana,
	}
}

func fromNanodaAccentPhrases(src []nanoda.AccentPhrase) []AccentPhrase {
	dst := make([]AccentPhrase, len(src))
	for i, ap := range src {
		dst[i] = AccentPhrase{
			Moras:           fromNanodaMoras(ap.Moras),
			Accent:          ap.Accent,
			PauseMora:       fromNanodaMoraPtr(ap.PauseMora),
			IsInterrogative: ap.IsInterrogative,
		}
	}
	return dst
}

func toNanodaAccentPhrases(src []AccentPhrase) []nanoda.AccentPhrase {
	dst := make([]nanoda.AccentPhrase, len(src))
	for i, ap := range src {
		dst[i] = nanoda.AccentPhrase{
			Moras:           toNanodaMoras(ap.Moras),
			Accent:          ap.Accent,
			PauseMora:       toNanodaMoraPtr(ap.PauseMora),
			IsInterrogative: ap.IsInterrogative,
		}
	}
	return dst
}

func fromNanodaMoras(src []nanoda.Mora) []Mora {
	dst := make([]Mora, len(src))
	for i, m := range src {
		dst[i] = Mora(m)
	}
	return dst
}

func toNanodaMoras(src []Mora) []nanoda.Mora {
	dst := make([]nanoda.Mora, len(src))
	for i, m := range src {
		dst[i] = nanoda.Mora(m)
	}
	return dst
}

func fromNanodaMoraPtr(m *nanoda.Mora) *Mora {
	if m == nil {
		return nil
	}
	c := Mora(*m)
	return &c
}

func toNanodaMoraPtr(m *Mora) *nanoda.Mora {
	if m == nil {
		return nil
	}
	c := nanoda.Mora(*m)
	return &c
}
