// Package project は台本(読み上げるテキストの並び)を表す。
package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/raa0121/nanoda/apps/voicevox/internal/voice"
)

// Ext はプロジェクトファイルの拡張子。
const Ext = ".nvproj"

// Line は台本の1行。VOICEVOXのテキスト欄1つに相当する。
type Line struct {
	Text  string        `json:"text"`
	Style voice.StyleID `json:"style_id"`
	// Query はエンジンが生成したAudioQuery。まだ生成していなければnil。
	Query *voice.Query `json:"query,omitempty"`
	// Stale はテキストやスタイルが変わってQueryが古くなったことを表す。
	Stale bool `json:"-"`
}

// NeedsQuery はこの行のクエリを作りなおす必要があるかどうかを返す。
func (l *Line) NeedsQuery() bool { return l.Query == nil || l.Stale }

// Project は台本全体。
type Project struct {
	Lines []*Line `json:"lines"`

	// Path は保存先。未保存なら空。
	Path string `json:"-"`
	// dirty は最後の保存から変更があったかどうか。
	dirty bool
}

type projectFile struct {
	Version int     `json:"version"`
	Lines   []*Line `json:"lines"`
}

const fileVersion = 1

// New は空行を1つだけ持つプロジェクトを作る。
func New(style voice.StyleID) *Project {
	return &Project{Lines: []*Line{{Style: style}}}
}

func (p *Project) Count() int { return len(p.Lines) }

// At は範囲外なら nil を返す。UIは行の追加・削除で番号がずれるので、
// 呼び出し側でいちいち範囲を確かめずに済むようにしている。
func (p *Project) At(i int) *Line {
	if i < 0 || i >= len(p.Lines) {
		return nil
	}
	return p.Lines[i]
}

func (p *Project) IsDirty() bool { return p.dirty }

// Insert は index の位置に行を挿入し、その位置を返す。
func (p *Project) Insert(index int, style voice.StyleID) int {
	if index < 0 {
		index = 0
	}
	if index > len(p.Lines) {
		index = len(p.Lines)
	}
	p.Lines = append(p.Lines, nil)
	copy(p.Lines[index+1:], p.Lines[index:])
	p.Lines[index] = &Line{Style: style}
	p.dirty = true
	return index
}

// Append は末尾に行を足し、その位置を返す。
func (p *Project) Append(style voice.StyleID) int {
	return p.Insert(len(p.Lines), style)
}

// Remove は行を削除し、削除後に選択すべき位置を返す。
// 最後の1行は消さずに空にする(VOICEVOXと同じ挙動)。
func (p *Project) Remove(index int) int {
	if index < 0 || index >= len(p.Lines) {
		return index
	}
	if len(p.Lines) == 1 {
		style := p.Lines[0].Style
		p.Lines[0] = &Line{Style: style}
		p.dirty = true
		return 0
	}
	p.Lines = append(p.Lines[:index], p.Lines[index+1:]...)
	p.dirty = true
	if index >= len(p.Lines) {
		return len(p.Lines) - 1
	}
	return index
}

// Move は行を from から to へ動かし、移動後の位置を返す。
func (p *Project) Move(from, to int) int {
	if from < 0 || from >= len(p.Lines) || to < 0 || to >= len(p.Lines) || from == to {
		return from
	}
	l := p.Lines[from]
	p.Lines = append(p.Lines[:from], p.Lines[from+1:]...)
	p.Lines = append(p.Lines, nil)
	copy(p.Lines[to+1:], p.Lines[to:])
	p.Lines[to] = l
	p.dirty = true
	return to
}

// SetText はテキストを差し替える。変わった場合はクエリを古いものとして
// 印を付ける。
func (p *Project) SetText(index int, text string) {
	l := p.At(index)
	if l == nil || l.Text == text {
		return
	}
	l.Text = text
	l.Stale = true
	p.dirty = true
}

// SetStyle はスタイルを差し替える。声が変わると音高や音素長も変わるので、
// クエリは作りなおしになる。
func (p *Project) SetStyle(index int, style voice.StyleID) {
	l := p.At(index)
	if l == nil || l.Style == style {
		return
	}
	l.Style = style
	l.Stale = true
	p.dirty = true
}

// SetQuery はエンジンが生成したクエリを取り込む。
func (p *Project) SetQuery(index int, q *voice.Query) {
	l := p.At(index)
	if l == nil {
		return
	}
	l.Query = q
	l.Stale = false
	p.dirty = true
}

// Touch は行のクエリを編集したことを記録する。
func (p *Project) Touch() { p.dirty = true }

// Save はプロジェクトをJSONで書き出す。
func (p *Project) Save(path string) error {
	if filepath.Ext(path) == "" {
		path += Ext
	}
	b, err := json.MarshalIndent(projectFile{Version: fileVersion, Lines: p.Lines}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return err
	}
	p.Path = path
	p.dirty = false
	return nil
}

// Load はプロジェクトを読み込む。
func Load(path string) (*Project, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f projectFile
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	if f.Version != fileVersion {
		return nil, fmt.Errorf("対応していないプロジェクト形式です(version %d)", f.Version)
	}
	p := &Project{Lines: f.Lines, Path: path}
	if len(p.Lines) == 0 {
		p.Lines = []*Line{{}}
	}
	return p, nil
}
