package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

type CustomTheme struct {
	fyne.Theme
}

func NewCustomTheme() *CustomTheme {
	return &CustomTheme{Theme: theme.DefaultTheme()}
}

func (t *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		// 设置背景色为浅绿色
		return color.NRGBA{R: 240, G: 248, B: 240, A: 255}
	case theme.ColorNamePrimary:
		// 设置主色调为绿色
		return color.NRGBA{R: 46, G: 139, B: 87, A: 255}
	case theme.ColorNameForeground:
		// 设置前景色（文字颜色）
		return color.NRGBA{R: 0, G: 100, B: 0, A: 255}
	default:
		return t.Theme.Color(name, variant)
	}
}
