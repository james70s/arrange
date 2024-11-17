package gui

import (
	"fmt"

	"github.com/james70s/arrange/internal/ver"
	"github.com/james70s/arrange/pkg/color"
	"github.com/james70s/arrange/pkg/config"
	"github.com/jroimartin/gocui"
)

func BannerView(g *gocui.Gui) error {
	maxX, _ := g.Size()

	if view, err := g.SetView("banner", 0, 0, maxX, 11); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		// view.Autoscroll = true
		// view.Wrap = true
		view.Frame = false

		// view.FgColor = gocui.ColorWhite
		// // view.BgColor = gocui.ColorBlack
		// // view.BgColor = gocui.ColorWhite
		// view.BgColor = gocui.ColorDefault
		// gocui.Attribute(0)

		fmt.Fprintln(view, color.String(
			config.C.Color.Green,
			ver.Banner(),
		))
	}
	return nil
}
