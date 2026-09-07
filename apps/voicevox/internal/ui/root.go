package ui

import (
	"slices"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
)

// Root はウィンドウ全体。VOICEVOXのエディタと同じく
// 「ツールバー / キャラクター一覧・台本・パラメータ / ステータスバー」
// の3段構成にしている。
type Root struct {
	guigui.DefaultWidget

	app *App

	background basicwidget.Background
	toolbar    toolbar
	divider    basicwidget.Divider
	characters characterPane
	script     scriptPane
	accent     accentPane
	params     paramPane
	status     basicwidget.Text

	items        []guigui.LinearLayoutItem
	mainLayout   guigui.LinearLayout
	mainItems    []guigui.LinearLayoutItem
	centerLayout guigui.LinearLayout
	centerItems  []guigui.LinearLayoutItem
}

// NewRoot はアプリの状態を持ったルートウィジェットを作る。
func NewRoot(app *App) *Root {
	return &Root{app: app}
}

func (r *Root) Env(context *guigui.Context, key guigui.EnvKey, source *guigui.EnvSource) (any, bool) {
	if key == envKeyApp && r.app != nil {
		return r.app, true
	}
	return nil, false
}

// WriteStateKey は入力ハンドラ以外(合成結果の反映や再生の進行)で変わる
// 状態を拾って、画面の作りなおしを促す。
func (r *Root) WriteStateKey(context *guigui.Context, w *guigui.StateKeyWriter) {
	if r.app == nil {
		return
	}
	w.WriteUint64(r.app.gen)
	w.WriteInt(r.app.Selected)
	w.WriteInt(r.app.SelectedMora.Phrase)
	w.WriteInt(r.app.SelectedMora.Mora)
	w.WriteInt(r.app.busy)
	w.WriteString(r.app.Status)
	w.WriteInt(r.app.Project.Count())
	w.WriteBool(r.app.IsPlaying())
}

func (r *Root) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	if r.app == nil {
		return nil
	}
	adder.AddWidget(&r.background)
	adder.AddWidget(&r.toolbar)
	adder.AddWidget(&r.divider)
	adder.AddWidget(&r.characters)
	adder.AddWidget(&r.script)
	adder.AddWidget(&r.params)
	if r.app.Caps.Accent {
		adder.AddWidget(&r.accent)
	}
	adder.AddWidget(&r.status)

	r.status.SetValue(r.app.Status)
	r.status.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
	return nil
}

func (r *Root) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	layouter.LayoutWidget(&r.background, widgetBounds.Bounds())
	if r.app == nil {
		return
	}
	u := basicwidget.UnitSize(context)

	r.centerItems = slices.Delete(r.centerItems, 0, len(r.centerItems))
	r.centerItems = append(r.centerItems,
		guigui.LinearLayoutItem{Widget: &r.script, Size: guigui.FlexibleSize(1)},
	)
	if r.app.Caps.Accent {
		r.centerItems = append(r.centerItems,
			guigui.LinearLayoutItem{Widget: &r.accent, Size: guigui.FixedSize(7 * u)},
		)
	}
	r.centerLayout = guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     r.centerItems,
		Gap:       u / 2,
	}

	r.mainItems = slices.Delete(r.mainItems, 0, len(r.mainItems))
	r.mainItems = append(r.mainItems,
		guigui.LinearLayoutItem{Widget: &r.characters, Size: guigui.FixedSize(10 * u)},
		guigui.LinearLayoutItem{Size: guigui.FlexibleSize(1), Layout: &r.centerLayout},
		guigui.LinearLayoutItem{Widget: &r.params, Size: guigui.FixedSize(13 * u)},
	)
	r.mainLayout = guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items:     r.mainItems,
		Gap:       u / 2,
	}

	r.items = slices.Delete(r.items, 0, len(r.items))
	r.items = append(r.items,
		guigui.LinearLayoutItem{Widget: &r.toolbar, Size: guigui.FixedSize(3 * u / 2)},
		guigui.LinearLayoutItem{Widget: &r.divider, Size: guigui.FixedSize(u / 4)},
		guigui.LinearLayoutItem{Size: guigui.FlexibleSize(1), Layout: &r.mainLayout},
		guigui.LinearLayoutItem{Widget: &r.status, Size: guigui.FixedSize(u)},
	)
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     r.items,
		Gap:       u / 2,
		Padding: guigui.Padding{
			Start:  u / 2,
			Top:    u / 2,
			End:    u / 2,
			Bottom: u / 2,
		},
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}

// Tick はワーカーからの結果を取り込み、再生待ちを進める。
// どちらもハンドラの外で状態を変えるので、変化があったら
// 明示的に作りなおしを頼む。
func (r *Root) Tick(context *guigui.Context, widgetBounds *guigui.WidgetBounds) error {
	if r.app == nil {
		return nil
	}
	changed := r.app.Drain()
	if r.app.TickPlayback() {
		changed = true
	}
	if changed {
		guigui.RequestRebuild()
	}
	return nil
}
