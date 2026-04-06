package ui

import (
	"coop-voicechat/coop"
	"coop-voicechat/fonts"
	"image/color"
	"log"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type UI struct {
	*app.Window

	theme  *material.Theme
	bridge *coop.Bridge
}

type (
	C = layout.Context
	D = layout.Dimensions
)

func New() *UI {
	width := unit.Dp(360)
	height := unit.Dp(96)

	window := new(app.Window)
	window.Option(
		app.Size(width, height),
		app.MinSize(width, height),
		app.MaxSize(width, height),
		app.Title("sm64coopdx Voice Chat"),
	)

	theme := material.NewTheme().WithPalette(material.Palette{
		Bg:         color.NRGBA{12, 12, 12, 255},
		Fg:         color.NRGBA{240, 240, 240, 255},
		ContrastBg: color.NRGBA{171, 43, 101, 255},
		ContrastFg: color.NRGBA{255, 255, 255, 255},
	})
	theme.Shaper = text.NewShaper(text.WithCollection(fonts.Collection()))
	theme.Face = "Nunito"

	return &UI{
		Window: window,
		theme:  &theme,
	}
}

func (ui *UI) Run() error {
	log.Println("UI running")

	ui.bridge = coop.NewBridge()
	go ui.bridge.Run()
	defer ui.bridge.Stop()

	var ops op.Ops
	for {
		switch e := ui.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			paint.Fill(gtx.Ops, ui.theme.Bg)

			e.Frame(gtx.Ops)
		}
	}
}
