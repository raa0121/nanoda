package ui

import (
	"image"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"

	"github.com/raa0121/nanoda/apps/voicevox/internal/voice"
)

// pitchRange は音高(対数F0)を棒の高さに写すときの範囲。
const (
	pitchMin = 4.8
	pitchMax = 6.6
)

// accentPane は下部のアクセント欄。VOICEVOXの「イントネーション」欄に
// あたる部分で、モーラごとの音高を棒グラフで見せ、アクセント核の位置を
// 動かせるようにしている。
type accentPane struct {
	guigui.DefaultWidget

	title      basicwidget.Text
	titleStyle basicwidget.TextStyle
	panel      basicwidget.Panel
	content    accentContent

	items []guigui.LinearLayoutItem
}

func (a *accentPane) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&a.title)
	adder.AddWidget(&a.panel)

	a.titleStyle.SetBold(true)
	a.title.SetBaseStyle(&a.titleStyle)
	a.title.SetValue("アクセント・イントネーション")
	a.title.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	a.panel.SetContent(&a.content)
	a.panel.SetAutoBorder(true)
	a.panel.SetContentConstraints(basicwidget.PanelContentConstraintsFixedHeight)
	return nil
}

func (a *accentPane) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)
	a.items = slices.Delete(a.items, 0, len(a.items))
	a.items = append(a.items,
		guigui.LinearLayoutItem{Widget: &a.title, Size: guigui.FixedSize(u)},
		guigui.LinearLayoutItem{Widget: &a.panel, Size: guigui.FlexibleSize(1)},
	)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     a.items,
		Gap:       u / 4,
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

// accentContent はアクセント句を横に並べる。
type accentContent struct {
	guigui.DefaultWidget

	phrases guigui.WidgetSlice[*accentPhraseWidget]
	note    basicwidget.Text

	count int
	items []guigui.LinearLayoutItem
}

func (a *accentContent) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	app := appOf(context, a)
	if app == nil {
		return nil
	}
	q := app.Query()
	if q == nil || len(q.AccentPhrases) == 0 {
		a.count = 0
		adder.AddWidget(&a.note)
		if app.Busy() {
			a.note.SetValue("解析中です…")
		} else {
			a.note.SetValue("テキストを入力して確定すると、モーラごとの音高とアクセント核をここで調整できます。")
		}
		a.note.SetMultiline(true)
		a.note.SetWrapMode(basicwidget.WrapModeNormal)
		return nil
	}

	a.count = len(q.AccentPhrases)
	a.phrases.SetLen(a.count)
	for _, w := range a.phrases.All() {
		adder.AddWidget(w)
	}
	for i, w := range a.phrases.All() {
		phrase := &q.AccentPhrases[i]
		w.SetPhrase(i, phrase, app.SelectedMora)
		w.OnMoraSelected(func(context *guigui.Context, moraIndex int) {
			app.SelectedMora = MoraRef{Phrase: i, Mora: moraIndex}
		})
		w.OnAccentChanged(func(context *guigui.Context, accent int) {
			phrase.Accent = accent
			app.Project.Touch()
			app.RequestAccentUpdate(app.Selected)
		})
		w.OnInterrogativeChanged(func(context *guigui.Context, value bool) {
			phrase.IsInterrogative = value
			app.Project.Touch()
			app.RequestAccentUpdate(app.Selected)
		})
	}
	return nil
}

func (a *accentContent) layout(context *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(context)
	a.items = slices.Delete(a.items, 0, len(a.items))
	if a.count == 0 {
		a.items = append(a.items, guigui.LinearLayoutItem{Widget: &a.note, Size: guigui.FlexibleSize(1)})
	} else {
		for _, w := range a.phrases.All() {
			a.items = append(a.items, guigui.LinearLayoutItem{Widget: w})
		}
		a.items = append(a.items, guigui.LinearLayoutItem{Size: guigui.FlexibleSize(1)})
	}
	return guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items:     a.items,
		Gap:       u / 2,
		Padding: guigui.Padding{
			Start:  u / 4,
			Top:    u / 4,
			End:    u / 4,
			Bottom: u / 4,
		},
	}
}

func (a *accentContent) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	a.layout(context).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (a *accentContent) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	return a.layout(context).Measure(context, constraints)
}

var (
	eventMoraSelected         = guigui.GenerateEventKey()
	eventAccentChanged        = guigui.GenerateEventKey()
	eventInterrogativeChanged = guigui.GenerateEventKey()
	eventMoraClicked          = guigui.GenerateEventKey()
)

// accentPhraseWidget はアクセント句1つ分。モーラの棒グラフと、
// アクセント核の位置・疑問形の切り替えを持つ。
type accentPhraseWidget struct {
	guigui.DefaultWidget

	moras              guigui.WidgetSlice[*moraWidget]
	accentLabel        basicwidget.Text
	accentInput        basicwidget.NumberInput
	interrogative      basicwidget.Checkbox
	interrogativeLabel basicwidget.Text
	pauseLabel         basicwidget.Text

	index           int
	moraCount       int
	accent          int
	hasPause        bool
	isInterrogative bool

	items       []guigui.LinearLayoutItem
	moraLayout  guigui.LinearLayout
	moraItems   []guigui.LinearLayoutItem
	ctrlLayout  guigui.LinearLayout
	ctrlItems   []guigui.LinearLayoutItem
	phraseMoras []moraView
}

// moraView は表示に必要な最小限のモーラ情報。
type moraView struct {
	Text     string
	Pitch    float64
	Selected bool
}

func (a *accentPhraseWidget) OnMoraSelected(f func(context *guigui.Context, moraIndex int)) {
	guigui.SetEventHandler(a, eventMoraSelected, f)
}

func (a *accentPhraseWidget) OnAccentChanged(f func(context *guigui.Context, accent int)) {
	guigui.SetEventHandler(a, eventAccentChanged, f)
}

func (a *accentPhraseWidget) OnInterrogativeChanged(f func(context *guigui.Context, value bool)) {
	guigui.SetEventHandler(a, eventInterrogativeChanged, f)
}

// SetPhrase は親から表示内容を流し込む。
func (a *accentPhraseWidget) SetPhrase(index int, phrase *voice.AccentPhrase, selected MoraRef) {
	a.index = index
	a.accent = phrase.Accent
	a.hasPause = phrase.PauseMora != nil
	a.isInterrogative = phrase.IsInterrogative
	a.moraCount = len(phrase.Moras)
	a.phraseMoras = slices.Delete(a.phraseMoras, 0, len(a.phraseMoras))
	for i, m := range phrase.Moras {
		a.phraseMoras = append(a.phraseMoras, moraView{
			Text:     m.Text,
			Pitch:    m.Pitch,
			Selected: selected.Phrase == index && selected.Mora == i,
		})
	}
}

func (a *accentPhraseWidget) WriteStateKey(context *guigui.Context, w *guigui.StateKeyWriter) {
	w.WriteInt(a.index)
	w.WriteInt(a.accent)
	w.WriteBool(a.hasPause)
	w.WriteBool(a.isInterrogative)
	for _, m := range a.phraseMoras {
		w.WriteString(m.Text)
		w.WriteFloat64(m.Pitch)
		w.WriteBool(m.Selected)
	}
}

func (a *accentPhraseWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	a.moras.SetLen(len(a.phraseMoras))
	for _, w := range a.moras.All() {
		adder.AddWidget(w)
	}
	adder.AddWidget(&a.accentLabel)
	adder.AddWidget(&a.accentInput)
	adder.AddWidget(&a.interrogative)
	adder.AddWidget(&a.interrogativeLabel)
	if a.hasPause {
		adder.AddWidget(&a.pauseLabel)
	}

	for i, w := range a.moras.All() {
		m := a.phraseMoras[i]
		w.SetMora(m.Text, m.Pitch, m.Selected)
		w.OnClicked(func(context *guigui.Context) {
			guigui.DispatchEvent(a, eventMoraSelected, i)
		})
	}

	a.accentLabel.SetValue("核")
	a.accentLabel.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	a.accentInput.SetMinimumValue(0)
	a.accentInput.SetMaximumValue(max(a.moraCount, 0))
	a.accentInput.SetValue(a.accent)
	a.accentInput.OnValueChanged(func(context *guigui.Context, value int, committed bool) {
		if !committed || value == a.accent {
			return
		}
		guigui.DispatchEvent(a, eventAccentChanged, value)
	})

	a.interrogative.SetValue(a.isInterrogative)
	a.interrogative.OnValueChanged(func(context *guigui.Context, value bool) {
		guigui.DispatchEvent(a, eventInterrogativeChanged, value)
	})
	a.interrogativeLabel.SetValue("？")
	a.interrogativeLabel.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	a.pauseLabel.SetValue("、")
	a.pauseLabel.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
	return nil
}

func (a *accentPhraseWidget) layout(context *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(context)

	a.moraItems = slices.Delete(a.moraItems, 0, len(a.moraItems))
	for _, w := range a.moras.All() {
		a.moraItems = append(a.moraItems, guigui.LinearLayoutItem{Widget: w, Size: guigui.FixedSize(5 * u / 4)})
	}
	if a.hasPause {
		a.moraItems = append(a.moraItems, guigui.LinearLayoutItem{Widget: &a.pauseLabel, Size: guigui.FixedSize(3 * u / 4)})
	}
	a.moraLayout = guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items:     a.moraItems,
		Gap:       u / 8,
	}

	a.ctrlItems = slices.Delete(a.ctrlItems, 0, len(a.ctrlItems))
	a.ctrlItems = append(a.ctrlItems,
		guigui.LinearLayoutItem{Widget: &a.accentLabel, Size: guigui.FixedSize(3 * u / 4)},
		guigui.LinearLayoutItem{Widget: &a.accentInput, Size: guigui.FixedSize(5 * u / 2)},
		guigui.LinearLayoutItem{Widget: &a.interrogative, Size: guigui.FixedSize(u)},
		guigui.LinearLayoutItem{Widget: &a.interrogativeLabel, Size: guigui.FixedSize(3 * u / 4)},
		guigui.LinearLayoutItem{Size: guigui.FlexibleSize(1)},
	)
	a.ctrlLayout = guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items:     a.ctrlItems,
		Gap:       u / 4,
	}

	a.items = slices.Delete(a.items, 0, len(a.items))
	a.items = append(a.items,
		guigui.LinearLayoutItem{Size: guigui.FlexibleSize(1), Layout: &a.moraLayout},
		guigui.LinearLayoutItem{Size: guigui.FixedSize(u), Layout: &a.ctrlLayout},
	)
	return guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     a.items,
		Gap:       u / 4,
	}
}

func (a *accentPhraseWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	a.layout(context).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (a *accentPhraseWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	u := basicwidget.UnitSize(context)
	w := len(a.phraseMoras)*(5*u/4+u/8) + 4*u
	if a.hasPause {
		w += 3 * u / 4
	}
	h, ok := constraints.FixedHeight()
	if !ok {
		h = 4 * u
	}
	return image.Pt(w, h)
}

// moraWidget はモーラ1つ。上に音高の棒、下に仮名を出す。
// 棒を押すとそのモーラが選択され、右のパラメータ欄で音高と長さを
// 細かく直せるようになる。
type moraWidget struct {
	guigui.DefaultWidget

	label basicwidget.Text

	text     string
	pitch    float64
	selected bool
}

func (m *moraWidget) SetMora(text string, pitch float64, selected bool) {
	m.text = text
	m.pitch = pitch
	m.selected = selected
}

func (m *moraWidget) OnClicked(f func(context *guigui.Context)) {
	guigui.SetEventHandler(m, eventMoraClicked, f)
}

// WriteStateKey は自前のDrawが読む値をすべて書き出す。これが無いと
// 値だけが変わったときに再描画されない。
func (m *moraWidget) WriteStateKey(context *guigui.Context, w *guigui.StateKeyWriter) {
	w.WriteString(m.text)
	w.WriteFloat64(m.pitch)
	w.WriteBool(m.selected)
}

func (m *moraWidget) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&m.label)
	m.label.SetValue(m.text)
	m.label.SetHorizontalAlign(basicwidget.HorizontalAlignCenter)
	m.label.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
	// ラベルがクリックを食べてしまうと棒を押せないので通過させる。
	context.SetPassthrough(&m.label, true)
	return nil
}

func (m *moraWidget) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)
	b := widgetBounds.Bounds()
	labelBounds := b
	labelBounds.Min.Y = maxInt(b.Max.Y-u, b.Min.Y)
	layouter.LayoutWidget(&m.label, labelBounds)
}

func (m *moraWidget) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	u := basicwidget.UnitSize(context)
	h, ok := constraints.FixedHeight()
	if !ok {
		h = 3 * u
	}
	return image.Pt(5*u/4, h)
}

func (m *moraWidget) HandlePointingInput(context *guigui.Context, widgetBounds *guigui.WidgetBounds) guigui.HandleInputResult {
	if widgetBounds.IsHitAtCursor() && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		guigui.DispatchEvent(m, eventMoraClicked)
		return guigui.HandleInputByWidget(m)
	}
	return guigui.HandleInputResult{}
}

func (m *moraWidget) CursorShape(context *guigui.Context, widgetBounds *guigui.WidgetBounds) (ebiten.CursorShapeType, bool) {
	if widgetBounds.IsHitAtCursor() {
		return ebiten.CursorShapePointer, true
	}
	return 0, false
}

func (m *moraWidget) Draw(context *guigui.Context, widgetBounds *guigui.WidgetBounds, dst *ebiten.Image) {
	u := basicwidget.UnitSize(context)
	b := widgetBounds.Bounds()
	graph := b
	graph.Max.Y = maxInt(b.Max.Y-u, b.Min.Y)
	if graph.Dy() <= 0 {
		return
	}

	tint := basicwidget.InfoTintColor()
	if m.selected {
		tint = basicwidget.AccentTintColor()
	}
	clr := basicwidget.TextColorFromTint(context, tint)

	// 無声化したモーラや句末のポーズは音高を持たないので、
	// 棒ではなく細い線だけを出す。
	ratio := (m.pitch - pitchMin) / (pitchMax - pitchMin)
	if m.pitch <= 0 {
		ratio = 0
	}
	ratio = clampFloat(ratio, 0, 1)

	barWidth := float32(graph.Dx()) * 0.6
	x := float32(graph.Min.X) + (float32(graph.Dx())-barWidth)/2
	height := float32(graph.Dy()) * float32(ratio)
	minHeight := float32(2 * context.Scale())
	if height < minHeight {
		height = minHeight
	}
	y := float32(graph.Max.Y) - height
	vector.DrawFilledRect(dst, x, y, barWidth, height, clr, false)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
