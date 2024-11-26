package gui

import "github.com/jroimartin/gocui"

func Quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// FocusStatusView puts the status window on top
func FocusStatusView(g *gocui.Gui, v *gocui.View) error {

	v.Autoscroll = true

	if _, err := g.SetCurrentView(""); err != nil {
		return err
	}

	return nil
}
