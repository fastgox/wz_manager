package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme 自定义主题
type CustomTheme struct{}

// 确保实现了 fyne.Theme 接口
var _ fyne.Theme = (*CustomTheme)(nil)

// 颜色定义
var (
	// 主色调 - 现代蓝色系
	primaryColor      = color.NRGBA{R: 59, G: 130, B: 246, A: 255} // 现代蓝色
	primaryLightColor = color.NRGBA{R: 96, G: 165, B: 250, A: 255} // 浅蓝色
	primaryDarkColor  = color.NRGBA{R: 37, G: 99, B: 235, A: 255}  // 深蓝色

	// 辅助色 - 现代绿色系
	accentColor      = color.NRGBA{R: 34, G: 197, B: 94, A: 255}  // 现代绿色
	accentLightColor = color.NRGBA{R: 74, G: 222, B: 128, A: 255} // 浅绿色

	// 背景色 - 现代白色主题
	backgroundColor = color.NRGBA{R: 249, G: 250, B: 251, A: 255} // 浅灰白色背景
	surfaceColor    = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // 纯白色表面
	cardColor       = color.NRGBA{R: 248, G: 250, B: 252, A: 255} // 卡片背景

	// 文字色 - 白色主题高对比度
	textColor          = color.NRGBA{R: 17, G: 24, B: 39, A: 255}    // 深色文字
	textSecondaryColor = color.NRGBA{R: 100, G: 116, B: 139, A: 255} // 中灰色文字
	textDisabledColor  = color.NRGBA{R: 148, G: 163, B: 184, A: 255} // 浅灰色文字

	// 边框和分割线 - 白色主题
	borderColor  = color.NRGBA{R: 229, G: 231, B: 235, A: 255} // 浅灰色边框
	dividerColor = color.NRGBA{R: 243, G: 244, B: 246, A: 255} // 分割线

	// 状态色 - 现代配色
	successColor = color.NRGBA{R: 34, G: 197, B: 94, A: 255}  // 现代绿色
	warningColor = color.NRGBA{R: 251, G: 191, B: 36, A: 255} // 现代黄色
	errorColor   = color.NRGBA{R: 239, G: 68, B: 68, A: 255}  // 现代红色

	// 悬停和选中状态 - 白色主题
	hoverColor    = color.NRGBA{R: 243, G: 244, B: 246, A: 255} // 悬停背景
	selectedColor = color.NRGBA{R: 59, G: 130, B: 246, A: 40}   // 选中背景（半透明蓝色）
	focusColor    = primaryColor                                // 焦点颜色
)

// Color 返回指定资源的颜色
func (t *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return primaryColor
	case theme.ColorNameBackground:
		return backgroundColor
	case theme.ColorNameButton:
		return primaryColor
	case theme.ColorNameDisabledButton:
		return textDisabledColor
	case theme.ColorNameDisabled:
		return textDisabledColor
	case theme.ColorNameError:
		return errorColor
	case theme.ColorNameFocus:
		return focusColor
	case theme.ColorNameForeground:
		return textColor
	case theme.ColorNameHover:
		return hoverColor
	case theme.ColorNameInputBackground:
		return surfaceColor
	case theme.ColorNameInputBorder:
		return borderColor
	case theme.ColorNameMenuBackground:
		return surfaceColor
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 128}
	case theme.ColorNamePlaceHolder:
		return textSecondaryColor
	case theme.ColorNamePressed:
		return primaryDarkColor
	case theme.ColorNameScrollBar:
		return borderColor
	case theme.ColorNameSelection:
		return selectedColor
	case theme.ColorNameSeparator:
		return dividerColor
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 25}
	case theme.ColorNameSuccess:
		return successColor
	case theme.ColorNameWarning:
		return warningColor
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font 返回指定资源的字体
func (t *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	// 使用中文字体资源
	return resourceSimheiTtf
}

// Icon 返回指定的图标资源
func (t *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size 返回指定的尺寸
func (t *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6 // 减少内边距
	case theme.SizeNameInlineIcon:
		return 18 // 稍小的图标
	case theme.SizeNameScrollBar:
		return 10 // 更细的滚动条
	case theme.SizeNameScrollBarSmall:
		return 6
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 13 // 稍小的文字
	case theme.SizeNameHeadingText:
		return 16 // 稍小的标题
	case theme.SizeNameSubHeadingText:
		return 14
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameInputBorder:
		return 1
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// ApplyCustomTheme 应用自定义主题
func ApplyCustomTheme(app fyne.App) {
	app.Settings().SetTheme(&CustomTheme{})
}

// CreateStyledCard 创建带样式的卡片容器
func CreateStyledCard(content fyne.CanvasObject) *fyne.Container {
	// 这里可以添加卡片样式，但Fyne的容器样式有限
	// 我们通过布局和间距来模拟卡片效果
	return fyne.NewContainerWithoutLayout(content)
}

// CreateStyledButton 创建带样式的按钮
func CreateStyledButton(text string, icon fyne.Resource, callback func()) *fyne.Container {
	// 由于Fyne的按钮样式限制，我们使用容器来创建更美观的按钮
	// 这里返回标准按钮，样式由主题控制
	return fyne.NewContainerWithoutLayout()
}

// GetPrimaryColor 获取主色调
func GetPrimaryColor() color.Color {
	return primaryColor
}

// GetAccentColor 获取辅助色
func GetAccentColor() color.Color {
	return accentColor
}

// GetSuccessColor 获取成功色
func GetSuccessColor() color.Color {
	return successColor
}

// GetWarningColor 获取警告色
func GetWarningColor() color.Color {
	return warningColor
}

// GetErrorColor 获取错误色
func GetErrorColor() color.Color {
	return errorColor
}
