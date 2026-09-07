package ui

import (
	"image"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"

	"github.com/raa0121/nanoda/apps/voicevox/internal/project"
	"github.com/raa0121/nanoda/apps/voicevox/internal/voice"
)

// scriptPane は中央の台本。VOICEVOXのテキスト欄が縦に並ぶところ。
type scriptPane struct {
	guigui.DefaultWidget

	title      basicwidget.Text
	titleStyle basicwidget.TextStyle
	panel      basicwidget.Panel
	content    scriptContent

	items []guigui.LinearLayoutItem
}

func (s *scriptPane) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&s.title)
	adder.AddWidget(&s.panel)

	s.titleStyle.SetBold(true)
	s.title.SetBaseStyle(&s.titleStyle)
	s.title.SetValue("台本")
	s.title.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	s.panel.SetContent(&s.content)
	s.panel.SetAutoBorder(true)
	s.panel.SetContentConstraints(basicwidget.PanelContentConstraintsFixedWidth)
	return nil
}

func (s *scriptPane) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)
	s.items = slices.Delete(s.items, 0, len(s.items))
	s.items = append(s.items,
		guigui.LinearLayoutItem{Widget: &s.title, Size: guigui.FixedSize(u)},
		guigui.LinearLayoutItem{Widget: &s.panel, Size: guigui.FlexibleSize(1)},
	)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     s.items,
		Gap:       u / 4,
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

// scriptContent は台本の行を並べるパネルの中身。
type scriptContent struct {
	guigui.DefaultWidget

	rows guigui.WidgetSlice[*scriptRow]

	items []guigui.LinearLayoutItem
}

func (s *scriptContent) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	app := appOf(context, s)
	if app == nil {
		return nil
	}
	s.rows.SetLen(app.Project.Count())
	for _, row := range s.rows.All() {
		adder.AddWidget(row)
	}
	for i, row := range s.rows.All() {
		line := app.Project.At(i)
		row.index = i
		row.SetLine(line, voice.StyleName(app.Speakers, line.Style), i == app.Selected)
		row.OnSelected(func(context *guigui.Context) {
			app.Select(i)
		})
		row.OnTextChanged(func(context *guigui.Context, text string, committed bool) {
			app.Select(i)
			app.Project.SetText(i, text)
			if committed {
				app.RequestQuery(i)
			}
		})
		row.OnPlayPressed(func(context *guigui.Context) {
			app.Select(i)
			app.PlayLine(i)
		})
		row.OnDeletePressed(func(context *guigui.Context) {
			app.Selected = app.Project.Remove(i)
			app.SelectedMora = NoMora
		})
		row.OnAppendPressed(func(context *guigui.Context) {
			app.Selected = app.Project.Insert(i+1, line.Style)
			app.SelectedMora = NoMora
		})
	}
	return nil
}

func (s *scriptContent) layout(context *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(context)
	s.items = slices.Delete(s.items, 0, len(s.items))
	for _, row := range s.rows.All() {
		s.items = append(s.items, guigui.LinearLayoutItem{Widget: row, Size: guigui.FixedSize(3 * u / 2)})
	}
	return guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     s.items,
		Gap:       u / 4,
		Padding: guigui.Padding{
			Start:  u / 4,
			Top:    u / 4,
			End:    u / 4,
			Bottom: u / 4,
		},
	}
}

func (s *scriptContent) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	s.layout(context).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (s *scriptContent) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	return s.layout(context).Measure(context, constraints)
}

var (
	eventRowSelected      = guigui.GenerateEventKey()
	eventRowTextChanged   = guigui.GenerateEventKey()
	eventRowPlayPressed   = guigui.GenerateEventKey()
	eventRowDeletePressed = guigui.GenerateEventKey()
	eventRowAppendPressed = guigui.GenerateEventKey()
)

// scriptRow は台本の1行。キャラクター名・テキスト欄・行単位の操作ボタン。
type scriptRow struct {
	guigui.DefaultWidget

	styleButton basicwidget.Button
	textInput   basicwidget.TextInput
	playButton  basicwidget.Button
	addButton   basicwidget.Button
	delButton   basicwidget.Button

	index     int
	styleName string
	text      string
	selected  bool
	stale     bool

	items []guigui.LinearLayoutItem
}

// SetLine は親から行の中身を流し込む。
func (r *scriptRow) SetLine(line *project.Line, styleName string, selected bool) {
	r.text = line.Text
	r.styleName = styleName
	r.selected = selected
	r.stale = line.NeedsQuery()
}

func (r *scriptRow) OnSelected(f func(context *guigui.Context)) {
	guigui.SetEventHandler(r, eventRowSelected, f)
}

func (r *scriptRow) OnTextChanged(f func(context *guigui.Context, text string, committed bool)) {
	guigui.SetEventHandler(r, eventRowTextChanged, f)
}

func (r *scriptRow) OnPlayPressed(f func(context *guigui.Context)) {
	guigui.SetEventHandler(r, eventRowPlayPressed, f)
}

func (r *scriptRow) OnDeletePressed(f func(context *guigui.Context)) {
	guigui.SetEventHandler(r, eventRowDeletePressed, f)
}

func (r *scriptRow) OnAppendPressed(f func(context *guigui.Context)) {
	guigui.SetEventHandler(r, eventRowAppendPressed, f)
}

func (r *scriptRow) WriteStateKey(context *guigui.Context, w *guigui.StateKeyWriter) {
	w.WriteInt(r.index)
	w.WriteString(r.styleName)
	w.WriteBool(r.selected)
	w.WriteBool(r.stale)
}

func (r *scriptRow) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&r.styleButton)
	adder.AddWidget(&r.textInput)
	adder.AddWidget(&r.playButton)
	adder.AddWidget(&r.addButton)
	adder.AddWidget(&r.delButton)

	r.styleButton.SetText(r.styleName)
	if r.selected {
		r.styleButton.SetType(basicwidget.ButtonTypePrimary)
	} else {
		r.styleButton.SetType(basicwidget.ButtonTypeNormal)
	}
	r.styleButton.OnDown(func(context *guigui.Context) {
		guigui.DispatchEvent(r, eventRowSelected)
	})

	r.textInput.SetValue(r.text)
	r.textInput.SetPlaceholder("読み上げるテキストを入力")
	r.textInput.OnValueChanged(func(context *guigui.Context, text string, committed bool) {
		guigui.DispatchEvent(r, eventRowTextChanged, text, committed)
	})
	// Enterで次の行を足すのはVOICEVOXと同じ操作感にするため。
	r.textInput.OnHandleButtonInput(func(context *guigui.Context, widgetBounds *guigui.WidgetBounds) guigui.HandleInputResult {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
			r.textInput.CommitWithCurrentInputValue()
			guigui.DispatchEvent(r, eventRowAppendPressed)
			return guigui.HandleInputByWidget(&r.textInput)
		}
		return guigui.HandleInputResult{}
	})

	r.playButton.SetText("▶")
	r.playButton.OnDown(func(context *guigui.Context) {
		guigui.DispatchEvent(r, eventRowPlayPressed)
	})

	r.addButton.SetText("＋")
	r.addButton.OnDown(func(context *guigui.Context) {
		guigui.DispatchEvent(r, eventRowAppendPressed)
	})

	r.delButton.SetText("×")
	r.delButton.OnDown(func(context *guigui.Context) {
		guigui.DispatchEvent(r, eventRowDeletePressed)
	})
	return nil
}

// Tick はテキスト欄にカーソルが入った行を選択中に切り替える。
// フォーカスは木が完成してからでないと読めないので、Buildではなく
// ここで見ている。
func (r *scriptRow) Tick(context *guigui.Context, widgetBounds *guigui.WidgetBounds) error {
	if !r.selected && context.IsFocused(&r.textInput) {
		guigui.DispatchEvent(r, eventRowSelected)
	}
	return nil
}

func (r *scriptRow) layout(context *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(context)
	r.items = slices.Delete(r.items, 0, len(r.items))
	r.items = append(r.items,
		guigui.LinearLayoutItem{Widget: &r.styleButton, Size: guigui.FixedSize(8 * u)},
		guigui.LinearLayoutItem{Widget: &r.textInput, Size: guigui.FlexibleSize(1)},
		guigui.LinearLayoutItem{Widget: &r.playButton, Size: guigui.FixedSize(3 * u / 2)},
		guigui.LinearLayoutItem{Widget: &r.addButton, Size: guigui.FixedSize(3 * u / 2)},
		guigui.LinearLayoutItem{Widget: &r.delButton, Size: guigui.FixedSize(3 * u / 2)},
	)
	return guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items:     r.items,
		Gap:       u / 4,
	}
}

func (r *scriptRow) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	r.layout(context).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (r *scriptRow) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	u := basicwidget.UnitSize(context)
	w, ok := constraints.FixedWidth()
	if !ok {
		w = 20 * u
	}
	return image.Pt(w, 3*u/2)
}
