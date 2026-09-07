package ui

import (
	"fmt"
	"image"
	"math"
	"slices"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
)

// sliderScale はSliderが整数しか扱えないため、実際の値を整数に写すときの倍率。
const sliderScale = 100

var eventParamChanged = guigui.GenerateEventKey()

// paramSlider は「ラベル / スライダー / 現在値」の1行。
// VOICEVOXの右側のパラメータ欄と同じ見た目にするための部品。
type paramSlider struct {
	guigui.DefaultWidget

	label  basicwidget.Text
	slider basicwidget.Slider
	value  basicwidget.Text

	name           string
	min, max, step float64
	format         string
	current        float64
	items          []guigui.LinearLayoutItem
}

// Configure は表示名と範囲を決める。範囲はVOICEVOXのエディタに合わせている。
func (p *paramSlider) Configure(name string, minimum, maximum, step float64, format string) {
	p.name = name
	p.min = minimum
	p.max = maximum
	p.step = step
	p.format = format
}

func (p *paramSlider) SetValue(v float64) { p.current = v }

func (p *paramSlider) OnChanged(f func(context *guigui.Context, value float64)) {
	guigui.SetEventHandler(p, eventParamChanged, f)
}

func (p *paramSlider) WriteStateKey(context *guigui.Context, w *guigui.StateKeyWriter) {
	w.WriteFloat64(p.current)
	w.WriteFloat64(p.min)
	w.WriteFloat64(p.max)
	w.WriteString(p.name)
}

func (p *paramSlider) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&p.label)
	adder.AddWidget(&p.slider)
	adder.AddWidget(&p.value)

	p.label.SetValue(p.name)
	p.label.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	p.slider.SetMinimumValue(toSliderValue(p.min))
	p.slider.SetMaximumValue(toSliderValue(p.max))
	p.slider.SetStep(toSliderValue(p.step))
	p.slider.SetValue(toSliderValue(p.current))
	p.slider.OnValueChanged(func(context *guigui.Context, value int) {
		p.current = float64(value) / sliderScale
		guigui.DispatchEvent(p, eventParamChanged, p.current)
	})

	p.value.SetValue(fmt.Sprintf(p.format, p.current))
	p.value.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
	p.value.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
	return nil
}

func toSliderValue(v float64) int { return int(math.Round(v * sliderScale)) }

func (p *paramSlider) layout(context *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(context)
	p.items = slices.Delete(p.items, 0, len(p.items))
	p.items = append(p.items,
		guigui.LinearLayoutItem{Widget: &p.label, Size: guigui.FixedSize(4 * u)},
		guigui.LinearLayoutItem{Widget: &p.slider, Size: guigui.FlexibleSize(1)},
		guigui.LinearLayoutItem{Widget: &p.value, Size: guigui.FixedSize(2 * u)},
	)
	return guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items:     p.items,
		Gap:       u / 4,
	}
}

func (p *paramSlider) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	p.layout(context).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (p *paramSlider) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	u := basicwidget.UnitSize(context)
	w, ok := constraints.FixedWidth()
	if !ok {
		w = 12 * u
	}
	return image.Pt(w, u)
}

// paramPane は右側のパラメータ欄。
type paramPane struct {
	guigui.DefaultWidget

	title      basicwidget.Text
	titleStyle basicwidget.TextStyle
	panel      basicwidget.Panel
	content    paramContent

	items []guigui.LinearLayoutItem
}

func (p *paramPane) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	adder.AddWidget(&p.title)
	adder.AddWidget(&p.panel)

	p.titleStyle.SetBold(true)
	p.title.SetBaseStyle(&p.titleStyle)
	p.title.SetValue("パラメータ")
	p.title.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	p.panel.SetContent(&p.content)
	p.panel.SetAutoBorder(true)
	p.panel.SetContentConstraints(basicwidget.PanelContentConstraintsFixedWidth)
	return nil
}

func (p *paramPane) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	u := basicwidget.UnitSize(context)
	p.items = slices.Delete(p.items, 0, len(p.items))
	p.items = append(p.items,
		guigui.LinearLayoutItem{Widget: &p.title, Size: guigui.FixedSize(u)},
		guigui.LinearLayoutItem{Widget: &p.panel, Size: guigui.FlexibleSize(1)},
	)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     p.items,
		Gap:       u / 4,
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

// paramIndex はスライダーの並び順。AudioQueryのどの項目に対応するかを表す。
const (
	paramSpeed = iota
	paramPitch
	paramIntonation
	paramVolume
	paramPre
	paramPost
	paramCount
)

const (
	moraParamPitch = iota
	moraParamLength
	moraParamCount
)

// paramContent はパラメータ欄の中身。AudioQueryが無いときは
// 代わりに理由を出す。
type paramContent struct {
	guigui.DefaultWidget

	sliders     [paramCount]paramSlider
	note        basicwidget.Text
	divider     basicwidget.Divider
	moraTitle   basicwidget.Text
	moraSliders [moraParamCount]paramSlider

	showParams bool
	showMora   bool
	items      []guigui.LinearLayoutItem
}

func (p *paramContent) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	app := appOf(context, p)
	if app == nil {
		return nil
	}
	q := app.Query()
	p.showParams = app.Caps.Query && q != nil
	mora := app.selectedMora()
	p.showMora = p.showParams && mora != nil

	if p.showParams {
		for i := range p.sliders {
			adder.AddWidget(&p.sliders[i])
		}
	} else {
		adder.AddWidget(&p.note)
	}
	if p.showMora {
		adder.AddWidget(&p.divider)
		adder.AddWidget(&p.moraTitle)
		for i := range p.moraSliders {
			adder.AddWidget(&p.moraSliders[i])
		}
	}

	switch {
	case !app.Caps.Query:
		p.note.SetValue("このエンジンではAudioQueryを扱えないため、パラメータの調整はできません。VOICEVOX CORE 0.15系を指定すると調整できます。")
	case app.Busy():
		p.note.SetValue("解析中です…")
	default:
		p.note.SetValue("テキストを入力して確定すると、話速や抑揚を調整できます。")
	}
	p.note.SetMultiline(true)
	p.note.SetWrapMode(basicwidget.WrapModeNormal)

	p.sliders[paramSpeed].Configure("話速", 0.5, 2, 0.01, "%.2f")
	p.sliders[paramPitch].Configure("音高", -0.15, 0.15, 0.01, "%.2f")
	p.sliders[paramIntonation].Configure("抑揚", 0, 2, 0.01, "%.2f")
	p.sliders[paramVolume].Configure("音量", 0, 2, 0.01, "%.2f")
	p.sliders[paramPre].Configure("開始無音", 0, 1.5, 0.01, "%.2f")
	p.sliders[paramPost].Configure("終了無音", 0, 1.5, 0.01, "%.2f")

	if q != nil {
		p.sliders[paramSpeed].SetValue(q.SpeedScale)
		p.sliders[paramPitch].SetValue(q.PitchScale)
		p.sliders[paramIntonation].SetValue(q.IntonationScale)
		p.sliders[paramVolume].SetValue(q.VolumeScale)
		p.sliders[paramPre].SetValue(q.PrePhonemeLength)
		p.sliders[paramPost].SetValue(q.PostPhonemeLength)
	}

	targets := [paramCount]*float64{}
	if q != nil {
		targets = [paramCount]*float64{
			&q.SpeedScale, &q.PitchScale, &q.IntonationScale,
			&q.VolumeScale, &q.PrePhonemeLength, &q.PostPhonemeLength,
		}
	}
	for i := range p.sliders {
		target := targets[i]
		p.sliders[i].OnChanged(func(context *guigui.Context, value float64) {
			if target == nil {
				return
			}
			*target = value
			app.Project.Touch()
		})
	}

	p.moraSliders[moraParamPitch].Configure("音高", 3, 6.5, 0.01, "%.2f")
	p.moraSliders[moraParamLength].Configure("長さ", 0, 0.6, 0.01, "%.2f")
	if mora != nil {
		p.moraTitle.SetValue("選択中のモーラ: " + mora.Text)
		p.moraSliders[moraParamPitch].SetValue(mora.Pitch)
		p.moraSliders[moraParamLength].SetValue(mora.VowelLength)
	}
	p.moraTitle.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
	p.moraSliders[moraParamPitch].OnChanged(func(context *guigui.Context, value float64) {
		if mora == nil {
			return
		}
		mora.Pitch = value
		app.Project.Touch()
	})
	p.moraSliders[moraParamLength].OnChanged(func(context *guigui.Context, value float64) {
		if mora == nil {
			return
		}
		mora.VowelLength = value
		app.Project.Touch()
	})
	return nil
}

func (p *paramContent) layout(context *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(context)
	p.items = slices.Delete(p.items, 0, len(p.items))
	if p.showParams {
		for i := range p.sliders {
			p.items = append(p.items, guigui.LinearLayoutItem{Widget: &p.sliders[i], Size: guigui.FixedSize(u)})
		}
	} else {
		p.items = append(p.items, guigui.LinearLayoutItem{Widget: &p.note})
	}
	if p.showMora {
		p.items = append(p.items,
			guigui.LinearLayoutItem{Widget: &p.divider, Size: guigui.FixedSize(u / 2)},
			guigui.LinearLayoutItem{Widget: &p.moraTitle, Size: guigui.FixedSize(u)},
		)
		for i := range p.moraSliders {
			p.items = append(p.items, guigui.LinearLayoutItem{Widget: &p.moraSliders[i], Size: guigui.FixedSize(u)})
		}
	}
	return guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     p.items,
		Gap:       u / 4,
		Padding: guigui.Padding{
			Start:  u / 4,
			Top:    u / 4,
			End:    u / 4,
			Bottom: u / 4,
		},
	}
}

func (p *paramContent) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	p.layout(context).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

func (p *paramContent) Measure(context *guigui.Context, constraints guigui.Constraints) image.Point {
	return p.layout(context).Measure(context, constraints)
}
