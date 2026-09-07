// Package ui はVOICEVOXのエディタをguiguiで再現したウィジェット群。
package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/guigui-gui/guigui"

	"github.com/raa0121/nanoda/apps/voicevox/internal/player"
	"github.com/raa0121/nanoda/apps/voicevox/internal/project"
	"github.com/raa0121/nanoda/apps/voicevox/internal/voice"
)

// envKeyApp はウィジェット木のどこからでもAppを引くための鍵。
var envKeyApp = guigui.GenerateEnvKey()

// MoraRef は選択中のモーラを指す。どちらかが負なら未選択。
type MoraRef struct {
	Phrase int
	Mora   int
}

// NoMora は「モーラを選んでいない」状態。
var NoMora = MoraRef{Phrase: -1, Mora: -1}

func (m MoraRef) Valid() bool { return m.Phrase >= 0 && m.Mora >= 0 }

// App はUI全体で共有する状態。ウィジェットからは context.Env で引く。
//
// フィールドの読み書きはすべてメインゴルーチン(Build/Layout/入力ハンドラ/
// Tick)で行う。エンジンの呼び出しだけをワーカーゴルーチンに逃がし、
// 結果は results チャネル経由でメインゴルーチンに戻して反映する。
type App struct {
	Engine   voice.Engine
	Player   *player.Player
	Project  *project.Project
	Speakers []voice.Speaker
	Caps     voice.Capabilities
	OutDir   string

	Selected     int
	SelectedMora MoraRef
	Status       string

	busy    int
	jobs    chan func()
	results chan func(*App)
	gen     uint64

	playQueue [][]byte
	playing   bool
	playedAt  time.Time
}

// NewApp はエンジンとプロジェクトからUIの状態を組み立て、
// エンジン呼び出し用のワーカーを起動する。
func NewApp(engine voice.Engine, pl *player.Player, proj *project.Project, outDir string) *App {
	a := &App{
		Engine:       engine,
		Player:       pl,
		Project:      proj,
		Speakers:     engine.Speakers(),
		Caps:         engine.Capabilities(),
		OutDir:       outDir,
		SelectedMora: NoMora,
		jobs:         make(chan func(), 64),
		results:      make(chan func(*App), 64),
	}
	a.Status = fmt.Sprintf("%s / バージョン %s", engine.Name(), engine.Version())
	// VOICEVOX COREの呼び出しはゴルーチンセーフとは限らないので、
	// ワーカーは1本だけにして直列化する。
	go func() {
		for job := range a.jobs {
			job()
		}
	}()
	// 開いた時点で選択されている行の読みを先に解析しておく。
	a.RequestQuery(0)
	return a
}

// Close はワーカーを止めてエンジンを閉じる。
func (a *App) Close() {
	close(a.jobs)
	if a.Engine != nil {
		_ = a.Engine.Close()
	}
}

// Busy は非同期処理が動いているかどうか。
func (a *App) Busy() bool { return a.busy > 0 }

// Line は選択中の行。無ければnil。
func (a *App) Line() *project.Line { return a.Project.At(a.Selected) }

// Query は選択中の行のクエリ。無ければnil。
func (a *App) Query() *voice.Query {
	if l := a.Line(); l != nil {
		return l.Query
	}
	return nil
}

// Bump は非同期に変えた状態を画面へ反映させるための世代番号を進める。
// Root.WriteStateKey がこれを見て再構築を起こす。
func (a *App) Bump() { a.gen++ }

// Select は行を選ぶ。行が変わればモーラの選択は外す。
func (a *App) Select(index int) {
	if index < 0 || index >= a.Project.Count() || index == a.Selected {
		return
	}
	a.Selected = index
	a.SelectedMora = NoMora
	a.RequestQuery(index)
}

// Drain はワーカーからの結果をメインゴルーチンに取り込む。
func (a *App) Drain() bool {
	var changed bool
	for {
		select {
		case f := <-a.results:
			f(a)
			changed = true
		default:
			if changed {
				a.Bump()
			}
			return changed
		}
	}
}

// TickPlayback は再生待ちのWAVを順に鳴らす。
func (a *App) TickPlayback() bool {
	if a.playing {
		// 再生直後は IsPlaying が false のことがあるので少し待つ。
		if a.Player.IsPlaying() || time.Since(a.playedAt) < 200*time.Millisecond {
			return false
		}
		a.playing = false
	}
	if len(a.playQueue) == 0 {
		return false
	}
	next := a.playQueue[0]
	a.playQueue = a.playQueue[1:]
	if err := a.Player.Play(next); err != nil {
		a.Status = "再生できませんでした: " + err.Error()
		a.playQueue = nil
		return true
	}
	a.playing = true
	a.playedAt = time.Now()
	return true
}

// IsPlaying は再生中か、再生待ちが残っているかを返す。
func (a *App) IsPlaying() bool { return a.playing || len(a.playQueue) > 0 }

// Stop は再生と再生待ちを止める。
func (a *App) Stop() {
	a.playQueue = nil
	a.playing = false
	a.Player.Stop()
	a.Status = "停止しました"
}

func (a *App) post(status string, work func(voice.Engine) (func(*App), error), onErr func(*App, error)) {
	if a.Engine == nil {
		return
	}
	a.busy++
	a.Status = status
	job := func() {
		apply, err := work(a.Engine)
		a.results <- func(a *App) {
			a.busy--
			if err != nil {
				a.Status = "エラー: " + err.Error()
				if onErr != nil {
					onErr(a, err)
				}
				return
			}
			if apply != nil {
				apply(a)
			}
		}
	}
	select {
	case a.jobs <- job:
	default:
		a.busy--
		a.Status = "処理が立て込んでいます。少し待ってからもう一度お試しください"
	}
}

// RequestQuery は行のAudioQueryを作りなおす。パラメータ欄とアクセント欄は
// これができて初めて中身が出る。
func (a *App) RequestQuery(index int) {
	if !a.Caps.Query {
		return
	}
	l := a.Project.At(index)
	if l == nil || strings.TrimSpace(l.Text) == "" || !l.NeedsQuery() {
		return
	}
	text, style := l.Text, l.Style
	// 二重に投げないよう、投げた時点で古い印を外す。失敗したら戻す。
	l.Stale = false
	a.post("読みを解析しています…", func(e voice.Engine) (func(*App), error) {
		q, err := e.CreateQuery(text, style)
		if err != nil {
			return nil, err
		}
		return func(a *App) {
			// 解析中にテキストが変わっていたら結果を捨てる。
			if l.Text != text || l.Style != style {
				return
			}
			l.Query = q
			l.Stale = false
			a.Project.Touch()
			a.Status = "読みを解析しました"
		}, nil
	}, func(a *App, err error) {
		l.Stale = true
	})
}

// RequestAccentUpdate はアクセント句を編集したあとに音高と音素長を
// 生成しなおす。VOICEVOXでアクセントを動かしたときと同じ処理。
func (a *App) RequestAccentUpdate(index int) {
	if !a.Caps.Accent {
		return
	}
	l := a.Project.At(index)
	if l == nil || l.Query == nil {
		return
	}
	phrases := l.Query.Clone().AccentPhrases
	style := l.Style
	a.post("アクセントを反映しています…", func(e voice.Engine) (func(*App), error) {
		res, err := e.ReplaceAccentPhrases(phrases, style)
		if err != nil {
			return nil, err
		}
		return func(a *App) {
			if l.Query == nil {
				return
			}
			l.Query.AccentPhrases = res
			a.Project.Touch()
			a.Status = "アクセントを反映しました"
		}, nil
	}, nil)
}

// synth は1行ぶんの音声を組み立てる。ワーカーゴルーチンから呼ばれる。
func (a *App) synth(e voice.Engine, l *project.Line) (*voice.Query, []byte, error) {
	if !a.Caps.Query {
		w, err := e.SynthesizeText(l.Text, l.Style)
		return nil, w, err
	}
	q := l.Query
	if q == nil || l.Stale {
		var err error
		q, err = e.CreateQuery(l.Text, l.Style)
		if err != nil {
			return nil, nil, err
		}
	}
	w, err := e.Synthesize(q, l.Style)
	if err != nil {
		return nil, nil, err
	}
	return q, w, nil
}

// snapshot はワーカーに渡すための行の写しを作る。メインゴルーチンが
// あとから書き換えても影響が出ないようにするため。
func snapshot(l *project.Line) *project.Line {
	return &project.Line{Text: l.Text, Style: l.Style, Query: l.Query.Clone(), Stale: l.Stale}
}

// PlayLine は1行を再生する。
func (a *App) PlayLine(index int) {
	l := a.Project.At(index)
	if l == nil || strings.TrimSpace(l.Text) == "" {
		a.Status = "テキストが空です"
		return
	}
	src := snapshot(l)
	a.post("音声を合成しています…", func(e voice.Engine) (func(*App), error) {
		q, w, err := a.synth(e, src)
		if err != nil {
			return nil, err
		}
		return func(a *App) {
			if q != nil && l.Text == src.Text && l.Style == src.Style {
				l.Query = q
				l.Stale = false
			}
			a.playQueue = append(a.playQueue, w)
			a.Status = "再生します"
		}, nil
	}, nil)
}

// PlayAll は台本を頭から順に再生する。
func (a *App) PlayAll() {
	lines := a.nonEmptyLines()
	if len(lines) == 0 {
		a.Status = "再生できる行がありません"
		return
	}
	a.post("すべての行を合成しています…", func(e voice.Engine) (func(*App), error) {
		wavs := make([][]byte, 0, len(lines))
		for _, l := range lines {
			_, w, err := a.synth(e, l)
			if err != nil {
				return nil, err
			}
			wavs = append(wavs, w)
		}
		return func(a *App) {
			a.playQueue = append(a.playQueue, wavs...)
			a.Status = fmt.Sprintf("%d行を再生します", len(wavs))
		}, nil
	}, nil)
}

func (a *App) nonEmptyLines() []*project.Line {
	var lines []*project.Line
	for i := 0; i < a.Project.Count(); i++ {
		l := a.Project.At(i)
		if strings.TrimSpace(l.Text) != "" {
			lines = append(lines, snapshot(l))
		}
	}
	return lines
}

// ExportLine は選択中の行をWAVとして書き出す。
func (a *App) ExportLine(index int) {
	l := a.Project.At(index)
	if l == nil || strings.TrimSpace(l.Text) == "" {
		a.Status = "テキストが空です"
		return
	}
	src := snapshot(l)
	name := fmt.Sprintf("%03d_style%d.wav", index+1, uint32(l.Style))
	a.post("書き出しています…", func(e voice.Engine) (func(*App), error) {
		_, w, err := a.synth(e, src)
		if err != nil {
			return nil, err
		}
		path, err := writeFile(a.OutDir, name, w)
		if err != nil {
			return nil, err
		}
		return func(a *App) {
			a.Status = "書き出しました: " + path
		}, nil
	}, nil)
}

// ExportAll は台本の全行をWAVとして書き出す。
func (a *App) ExportAll() {
	lines := a.nonEmptyLines()
	if len(lines) == 0 {
		a.Status = "書き出せる行がありません"
		return
	}
	a.post("すべて書き出しています…", func(e voice.Engine) (func(*App), error) {
		var last string
		for i, l := range lines {
			_, w, err := a.synth(e, l)
			if err != nil {
				return nil, err
			}
			name := fmt.Sprintf("%03d_style%d.wav", i+1, uint32(l.Style))
			last, err = writeFile(a.OutDir, name, w)
			if err != nil {
				return nil, err
			}
		}
		n := len(lines)
		dir := filepath.Dir(last)
		return func(a *App) {
			a.Status = fmt.Sprintf("%d件を %s に書き出しました", n, dir)
		}, nil
	}, nil)
}

func writeFile(dir, name string, b []byte) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// ProjectPath は保存・読み込みに使うパス。
func (a *App) ProjectPath() string {
	if a.Project.Path != "" {
		return a.Project.Path
	}
	return filepath.Join(a.OutDir, "project"+project.Ext)
}

// SaveProject は台本を保存する。
func (a *App) SaveProject() {
	path := a.ProjectPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		a.Status = "保存できませんでした: " + err.Error()
		return
	}
	if err := a.Project.Save(path); err != nil {
		a.Status = "保存できませんでした: " + err.Error()
		return
	}
	a.Status = "保存しました: " + a.Project.Path
}

// LoadProject は台本を読み込む。
func (a *App) LoadProject() {
	path := a.ProjectPath()
	p, err := project.Load(path)
	if err != nil {
		a.Status = "読み込めませんでした: " + err.Error()
		return
	}
	a.Project = p
	a.Selected = 0
	a.SelectedMora = NoMora
	a.Status = "読み込みました: " + path
	a.RequestQuery(0)
}

// currentStyle は新しい行に付けるスタイル。選択中の行があればそれに揃える。
func (a *App) currentStyle() voice.StyleID {
	if l := a.Line(); l != nil {
		return l.Style
	}
	return voice.FirstStyleID(a.Speakers)
}

// selectedMora は選択中のモーラを指す。クエリの中身を直接指しているので、
// 返り値を書き換えるとそのままプロジェクトに反映される。
func (a *App) selectedMora() *voice.Mora {
	q := a.Query()
	if q == nil || !a.SelectedMora.Valid() {
		return nil
	}
	if a.SelectedMora.Phrase >= len(q.AccentPhrases) {
		return nil
	}
	ap := &q.AccentPhrases[a.SelectedMora.Phrase]
	if a.SelectedMora.Mora >= len(ap.Moras) {
		return nil
	}
	return &ap.Moras[a.SelectedMora.Mora]
}
