// Package player は合成したWAVを鳴らす。
package player

import (
	"bytes"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

// sampleRate は再生側のサンプリングレート。VOICEVOXの出力は24000Hzだが、
// wav.DecodeWithSampleRate がここに合わせてリサンプルしてくれる。
const sampleRate = 48000

// Player は一度に1つの音声を鳴らす。新しい再生を始めると前の再生は止まる。
type Player struct {
	ctx     *audio.Context
	current *audio.Player
}

// New はオーディオコンテキストを1つ作る。プロセスにつき1回だけ呼ぶこと。
func New() *Player {
	return &Player{ctx: audio.NewContext(sampleRate)}
}

// Play はWAVのバイト列を鳴らす。
func (p *Player) Play(wavBytes []byte) error {
	p.Stop()
	s, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(wavBytes))
	if err != nil {
		return err
	}
	pl, err := p.ctx.NewPlayer(s)
	if err != nil {
		return err
	}
	p.current = pl
	pl.Play()
	return nil
}

// Stop は再生中の音声を止める。
func (p *Player) Stop() {
	if p.current == nil {
		return
	}
	p.current.Pause()
	_ = p.current.Close()
	p.current = nil
}

// IsPlaying は再生中かどうかを返す。
func (p *Player) IsPlaying() bool {
	return p.current != nil && p.current.IsPlaying()
}
