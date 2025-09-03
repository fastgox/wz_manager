package ui

import (
	"image/color"
	"io"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
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
	// 尝试加载系统中文字体
	fontPaths := []string{
		"C:/Windows/Fonts/msyh.ttc",   // 微软雅黑
		"C:/Windows/Fonts/simsun.ttc", // 宋体
		"C:/Windows/Fonts/simhei.ttf", // 黑体
		"C:/Windows/Fonts/arial.ttf",  // Arial (fallback)
	}

	for _, fontPath := range fontPaths {
		if _, err := os.Stat(fontPath); err == nil {
			if fontResource := loadFontResource(fontPath); fontResource != nil {
				return fontResource
			}
		}
	}

	// 如果都找不到，使用默认字体
	return theme.DefaultTheme().Font(style)
}

// loadFontResource 加载字体资源
func loadFontResource(fontPath string) fyne.Resource {
	// 尝试从文件系统加载字体
	if uri := storage.NewFileURI(fontPath); uri != nil {
		if reader, err := storage.Reader(uri); err == nil {
			defer reader.Close()
			if data, err := io.ReadAll(reader); err == nil && len(data) > 0 {
				return fyne.NewStaticResource(filepath.Base(fontPath), data)
			}
		}
	}
	return nil
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
