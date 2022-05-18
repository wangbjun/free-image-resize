package theme

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

type MyTheme struct{}

var _ fyne.Theme = (*MyTheme)(nil)

// Font return bundled font resource
func (*MyTheme) Font(s fyne.TextStyle) fyne.Resource {
	return resourceWqyMicroheiTtc
}
func (*MyTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DarkTheme().Color(n, v)
}

func (*MyTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	if n == theme.IconNameInfo {
		return &fyne.StaticResource{
			StaticName:    "info.svg",
			StaticContent: []byte(""),
		}
	}
	return theme.DarkTheme().Icon(n)
}

func (*MyTheme) Size(n fyne.ThemeSizeName) float32 {
	return theme.DarkTheme().Size(n)
}
