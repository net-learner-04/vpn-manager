package ui

import (
	"fyne.io/fyne/v2"
)

func BuildWindow(a fyne.App, w fyne.Window) {
	ApplyTheme(a, w)
	w.SetContent(BuildNavigation())
}
