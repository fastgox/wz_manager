package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ChineseTheme 支持中文的主题
type ChineseTheme struct {
	fyne.Theme
}

// NewChineseTheme 创建支持中文的主题
func NewChineseTheme() fyne.Theme {
	return &ChineseTheme{
		Theme: theme.DefaultTheme(),
	}
}

// Font 返回字体资源
func (t *ChineseTheme) Font(style fyne.TextStyle) fyne.Resource {
	// 使用打包的中文字体资源
	return resourceSimheiTtf
}

// Size 返回尺寸
func (t *ChineseTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14 // 适中的字体大小
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// Color 返回颜色
func (t *ChineseTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

// Icon 返回图标
func (t *ChineseTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
