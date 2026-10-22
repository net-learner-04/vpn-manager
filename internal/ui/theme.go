package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"

	"vpn-manager/assets"
)

func rgb(r, g, b uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: 0xFF} }

var (
	colFace      = rgb(0xD4, 0xD0, 0xC8) // window / panel background
	colButton    = rgb(0xE8, 0xE5, 0xDE) // buttons, lighter than the background so they stand out
	colBorder    = rgb(0x80, 0x80, 0x80) // every border, separator and scrollbar
	colWhite     = rgb(0xFF, 0xFF, 0xFF)
	colBlack     = rgb(0x00, 0x00, 0x00)
	colNavy      = rgb(0x0A, 0x24, 0x6A) // accent (primary buttons, checks)
	colSelection = rgb(0xB4, 0xBC, 0xD6) // light so the text stays readable when selected
	colLink      = rgb(0x00, 0x00, 0xFF)
	colError     = rgb(0xCC, 0x00, 0x00)
	colSuccess   = rgb(0x00, 0x80, 0x00)
	colWarning   = rgb(0xFF, 0xA5, 0x00)

	// Translucent overlays keep the content visible underneath.
	colHover   = color.NRGBA{A: 0x14}
	colPressed = color.NRGBA{A: 0x30}
	colFocus   = color.NRGBA{R: 0x0A, G: 0x24, B: 0x6A, A: 0x30}
)

type Theme struct{}

var _ fyne.Theme = (*Theme)(nil)

func (t Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground,
		theme.ColorNameMenuBackground,
		theme.ColorNameOverlayBackground,
		theme.ColorNameHeaderBackground,
		theme.ColorNameDisabledButton:
		return colFace

	case theme.ColorNameButton:
		return colButton

	case theme.ColorNameInputBackground:
		return colWhite

	case theme.ColorNameInputBorder,
		theme.ColorNameSeparator,
		theme.ColorNameScrollBar,
		theme.ColorNameDisabled,
		theme.ColorNamePlaceHolder,
		theme.ColorNameInnerWindowBorder,
		theme.ColorNameInnerWindowBorderInactive:
		return colBorder

	case theme.ColorNameScrollBarBackground:
		return colFace

	case theme.ColorNameForeground:
		return colBlack

	case theme.ColorNameSelection:
		return colSelection

	case theme.ColorNameHover:
		return colHover

	case theme.ColorNamePressed:
		return colPressed

	case theme.ColorNameFocus:
		return colFocus

	case theme.ColorNamePrimary:
		return colNavy

	case theme.ColorNameForegroundOnPrimary,
		theme.ColorNameForegroundOnError,
		theme.ColorNameForegroundOnSuccess:
		return colWhite

	case theme.ColorNameForegroundOnWarning:
		return colBlack

	case theme.ColorNameShadow:
		return color.NRGBA{A: 0x40}

	case theme.ColorNameHyperlink:
		return colLink

	case theme.ColorNameError:
		return colError

	case theme.ColorNameSuccess:
		return colSuccess

	case theme.ColorNameWarning:
		return colWarning
	}

	return theme.DefaultTheme().Color(name, theme.VariantLight)
}

func (t Theme) Font(style fyne.TextStyle) fyne.Resource {
	// The default font has no Korean glyphs; return an embedded font here.
	return theme.DefaultTheme().Font(style)
}

func (t Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t Theme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 12
	case theme.SizeNameSubHeadingText:
		return 13
	case theme.SizeNameHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameInnerPadding:
		return 6
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameScrollBar:
		return 14
	case theme.SizeNameScrollBarSmall:
		return 6

	// Square corners everywhere for a consistent look.
	case theme.SizeNameInputRadius,
		theme.SizeNameSelectionRadius,
		theme.SizeNameScrollBarRadius,
		theme.SizeNameButtonRadius,
		theme.SizeNameCardRadius,
		theme.SizeNameDialogRadius,
		theme.SizeNameMenuRadius,
		theme.SizeNamePopupRadius,
		theme.SizeNameInnerWindowRadius,
		theme.SizeNameWindowButtonRadius:
		return 0
	}
	return theme.DefaultTheme().Size(name)
}

func ApplyTheme(a fyne.App, w fyne.Window) {
	a.Settings().SetTheme(&Theme{})

	iconRes := fyne.NewStaticResource("icon.png", assets.Icon)
	a.SetIcon(iconRes)
	w.SetIcon(iconRes)

	w.Resize(fyne.NewSize(750, 500))
	w.SetFixedSize(true)

	// System tray is only available on supported desktop platforms.
	if desk, ok := a.(desktop.App); ok {
		menu := fyne.NewMenu("VPN Manager",
			fyne.NewMenuItem("Open", func() {
				w.Show()
				w.RequestFocus()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() { a.Quit() }),
		)
		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayIcon(iconRes)

		// Hide instead of closing so the app keeps running in the tray.
		w.SetCloseIntercept(func() { w.Hide() })
	}
}
