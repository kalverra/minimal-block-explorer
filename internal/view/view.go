package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type View interface {
	Show() tview.Primitive
	Controls() ControlMapping
	End()
}

type ControlMapping struct {
	NormalControls  NormalKeyControls
	SpecialControls SpecialKeyControls
}

type NormalKeyControls map[rune]Control
type SpecialKeyControls map[tcell.Key]Control

type Control struct {
	Key         string
	Description string
	Fn          func(app *App) error
}
