package ui

import (
	"slices"

	"github.com/guigui-gui/guigui"
	"github.com/guigui-gui/guigui/basicwidget"

	"github.com/raa0121/nanoda/apps/voicevox/internal/voice"
)

// characterPane は左端のキャラクター一覧。選んだスタイルが
// 選択中の行に適用される。
type characterPane struct {
	guigui.DefaultWidget

	title      basicwidget.Text
	titleStyle basicwidget.TextStyle
	list       basicwidget.List[voice.StyleID]
	note       basicwidget.Text
	styleInput basicwidget.NumberInput

	listItems []basicwidget.ListItem[voice.StyleID]
	items     []guigui.LinearLayoutItem
}

func (c *characterPane) Build(context *guigui.Context, adder *guigui.ChildAdder) error {
	app := appOf(context, c)
	if app == nil {
		return nil
	}
	adder.AddWidget(&c.title)
	if app.Caps.Speakers {
		adder.AddWidget(&c.list)
	} else {
		adder.AddWidget(&c.note)
		adder.AddWidget(&c.styleInput)
	}

	c.titleStyle.SetBold(true)
	c.title.SetBaseStyle(&c.titleStyle)
	c.title.SetValue("キャラクター")
	c.title.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	c.listItems = slices.Delete(c.listItems, 0, len(c.listItems))
	for _, sp := range app.Speakers {
		c.listItems = append(c.listItems, basicwidget.ListItem[voice.StyleID]{
			Text:         sp.Name,
			Header:       true,
			Unselectable: true,
		})
		for _, st := range sp.Styles {
			c.listItems = append(c.listItems, basicwidget.ListItem[voice.StyleID]{
				Text:        st.Name,
				Value:       st.ID,
				IndentLevel: 1,
			})
		}
	}
	c.list.SetItems(c.listItems)
	// プログラムからの選択ではOnItemSelectedは飛ばないので、
	// 毎回の再構築で選択中の行のスタイルに合わせておける。
	c.list.SelectItemByValue(app.currentStyle())
	c.list.OnItemSelected(func(context *guigui.Context, index int) {
		item, ok := c.list.ItemByIndex(index)
		if !ok || item.Header {
			return
		}
		app.Project.SetStyle(app.Selected, item.Value)
		app.SelectedMora = NoMora
		app.RequestQuery(app.Selected)
	})

	c.note.SetValue("このエンジンではキャラクター一覧を取得できません。スタイルIDを直接指定してください。")
	c.note.SetMultiline(true)
	c.note.SetWrapMode(basicwidget.WrapModeNormal)

	c.styleInput.SetMinimumValue(0)
	c.styleInput.SetMaximumValue(1 << 20)
	c.styleInput.SetValue(int(app.currentStyle()))
	c.styleInput.OnValueChanged(func(context *guigui.Context, value int, committed bool) {
		if !committed {
			return
		}
		app.Project.SetStyle(app.Selected, voice.StyleID(value))
	})
	return nil
}

func (c *characterPane) Layout(context *guigui.Context, widgetBounds *guigui.WidgetBounds, layouter *guigui.ChildLayouter) {
	app := appOf(context, c)
	if app == nil {
		return
	}
	u := basicwidget.UnitSize(context)
	c.items = slices.Delete(c.items, 0, len(c.items))
	c.items = append(c.items, guigui.LinearLayoutItem{Widget: &c.title, Size: guigui.FixedSize(u)})
	if app.Caps.Speakers {
		c.items = append(c.items, guigui.LinearLayoutItem{Widget: &c.list, Size: guigui.FlexibleSize(1)})
	} else {
		c.items = append(c.items,
			guigui.LinearLayoutItem{Widget: &c.note},
			guigui.LinearLayoutItem{Widget: &c.styleInput, Size: guigui.FixedSize(u)},
			guigui.LinearLayoutItem{Size: guigui.FlexibleSize(1)},
		)
	}
	(guigui.LinearLayout{
		Direction: guigui.LayoutDirectionVertical,
		Items:     c.items,
		Gap:       u / 4,
	}).LayoutWidgets(context, widgetBounds.Bounds(), layouter)
}
