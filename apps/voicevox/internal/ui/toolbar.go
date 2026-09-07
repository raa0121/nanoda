package ui

import (
	"slices"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"
)

// appOf はウィジェット木を遡ってAppを取り出す。まだ木に入っていない
// ウィジェットからは引けないので、nilを返しうることに注意。
func appOf(context *guigui.Context, w guigui.Widget) *App {
	v, ok := context.Env(w, envKeyApp)
	if !ok {
		return nil
	}
	a, _ := v.(*App)
	return a
}

// toolbar は再生・行の操作・書き出しをまとめた上端のバー。
type toolbar struct {
	guigui.DefaultWidget

	play      basicwidget.Button
	playAll   basicwidget.Button
	stop      basicwidget.Button
	addLine   basicwidget.Button
	delLine   basicwidget.Button
	export    basicwidget.Button
	exportAll basicwidget.Button
	save      basicwidget.Button
	load      basicwidget.Button

	items []guigui.LinearLayoutItem
}

func (t *toolbar) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	app := appOf(context, t)
	if app == nil {
		return nil
	}
	adder.AddWidget(&t.play)
	adder.AddWidget(&t.playAll)
	adder.AddWidget(&t.stop)
	adder.AddWidget(&t.addLine)
	adder.AddWidget(&t.delLine)
	adder.AddWidget(&t.export)
	adder.AddWidget(&t.exportAll)
	adder.AddWidget(&t.save)
	adder.AddWidget(&t.load)

	t.play.SetText("再生")
	t.play.SetType(basicwidget.ButtonTypePrimary)
	t.play.OnDown(func(context *guigui.Context) {
		app.PlayLine(app.Selected)
	})

	t.playAll.SetText("連続再生")
	t.playAll.OnDown(func(context *guigui.Context) {
		app.PlayAll()
	})

	t.stop.SetText("停止")
	t.stop.OnDown(func(context *guigui.Context) {
		app.Stop()
	})
	context.SetEnabled(&t.stop, app.IsPlaying())

	t.addLine.SetText("行を追加")
	t.addLine.OnDown(func(context *guigui.Context) {
		style := app.currentStyle()
		app.Selected = app.Project.Insert(app.Selected+1, style)
		app.SelectedMora = NoMora
	})

	t.delLine.SetText("行を削除")
	t.delLine.OnDown(func(context *guigui.Context) {
		app.Selected = app.Project.Remove(app.Selected)
		app.SelectedMora = NoMora
	})

	t.export.SetText("書き出し")
	t.export.OnDown(func(context *guigui.Context) {
		app.ExportLine(app.Selected)
	})

	t.exportAll.SetText("全部書き出し")
	t.exportAll.OnDown(func(context *guigui.Context) {
		app.ExportAll()
	})

	t.save.SetText("保存")
	t.save.OnDown(func(context *guigui.Context) {
		app.SaveProject()
	})

	t.load.SetText("開く")
	t.load.OnDown(func(context *guigui.Context) {
		app.LoadProject()
	})

	// 合成中はエンジンを叩くボタンを伏せる。ワーカーは1本なので
	// 押せてもキューに積まれるだけだが、押せないほうが分かりやすい。
	busy := app.Busy()
	for _, b := range []*basicwidget.Button{&t.play, &t.playAll, &t.export, &t.exportAll} {
		context.SetEnabled(b, !busy)
	}
	return nil
}

func (t *toolbar) layout(context *guigui.Context) guigui.LinearLayout {
	u := basicwidget.UnitSize(context)
	t.items = slices.Delete(t.items, 0, len(t.items))
	for _, b := range []*basicwidget.Button{
		&t.play, &t.playAll, &t.stop, &t.addLine, &t.delLine,
		&t.export, &t.exportAll, &t.save, &t.load,
	} {
		t.items = append(t.items, guigui.LinearLayoutItem{Widget: b})
	}
	t.items = append(t.items, guigui.LinearLayoutItem{Size: guigui.FlexibleSize(1)})
	return guigui.LinearLayout{
		Direction: guigui.LayoutDirectionHorizontal,
		Items:     t.items,
		Gap:       u / 4,
	}
}

func (t *toolbar) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	t.layout(context).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}
