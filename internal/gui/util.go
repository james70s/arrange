package gui

import "github.com/jroimartin/gocui"

func Quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// FocusStatusView pus the status window ontop
func FocusStatusView(g *gocui.Gui, v *gocui.View) error {

	v.Autoscroll = true

	if _, err := g.SetCurrentView(""); err != nil {
		return err
	}

	return nil
}
