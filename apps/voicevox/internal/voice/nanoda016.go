package voice

import (
	nanoda2 "github.com/aethiopicuschan/nanoda/v2"
)

// core016 は VOICEVOX CORE 0.16系 を nanoda/v2 経由で扱う実装。
//
// nanoda/v2 は「テキストからWAVを作る最小の経路」までしか移植されて
// おらず、AudioQuery・アクセント句・メタ情報取得は未対応
// (https://github.com/aethiopicuschan/nanoda の「対応状況」を参照)。
// そのため Capabilities で機能が無いことを申告し、UIは話速などの
// パラメータ欄を伏せてテキスト読み上げだけを提供する。
// nanoda/v2 がAudioQueryに対応したら、core015 と同じ実装をここに
// 足すだけでエディタとして完全になる。
type core016 struct {
	v       *nanoda2.Voicevox
	s       *nanoda2.Synthesizer
	version string
}

func open016(cfg Config) (Engine, error) {
	v, err := nanoda2.New(cfg.CorePath, cfg.DictPath, cfg.ModelPath)
	if err != nil {
		return nil, err
	}
	s, err := v.NewSynthesizer()
	if err != nil {
		return nil, err
	}
	// 0.16系はスタイルIDから引ける形でモデルを列挙できないため、
	// 起動時にまとめて読み込む。
	if err := s.LoadAllModels(); err != nil {
		s.Close()
		return nil, err
	}
	return &core016{v: v, s: s, version: v.Version()}, nil
}

func (c *core016) Name() string    { return "nanoda/v2 (VOICEVOX CORE 0.16系)" }
func (c *core016) Version() string { return c.version }

func (c *core016) Capabilities() Capabilities {
	return Capabilities{Speakers: false, Query: false, Accent: false}
}

func (c *core016) Speakers() []Speaker { return nil }

func (c *core016) CreateQuery(text string, style StyleID) (*Query, error) {
	return nil, ErrUnsupported
}

func (c *core016) ReplaceAccentPhrases(phrases []AccentPhrase, style StyleID) ([]AccentPhrase, error) {
	return nil, ErrUnsupported
}

func (c *core016) Synthesize(q *Query, style StyleID) ([]byte, error) {
	return nil, ErrUnsupported
}

func (c *core016) SynthesizeText(text string, style StyleID) ([]byte, error) {
	r, err := c.s.Tts(text, nanoda2.StyleId(style))
	if err != nil {
		return nil, err
	}
	return readAllAndClose(r)
}

func (c *core016) Close() error {
	c.s.Close()
	return nil
}
