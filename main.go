package main

import (
	"os"
	"wz_manager/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	// 设置UTF-8环境变量以确保正确的中文显示
	os.Setenv("LANG", "zh_CN.UTF-8")
	os.Setenv("LC_ALL", "zh_CN.UTF-8")
	os.Setenv("LC_CTYPE", "zh_CN.UTF-8")

	myApp := app.New()

	// 应用自定义主题
	ui.ApplyCustomTheme(myApp)

	mainWindow := myApp.NewWindow("🎮 WZ文件管理器 - MapleStory数据浏览器")
	mainWindow.Resize(fyne.NewSize(1200, 800))
	mainWindow.CenterOnScreen()

	// 创建主界面
	mainUI := ui.NewMainWindow(mainWindow)
	mainWindow.SetContent(mainUI.GetContent())

	mainWindow.ShowAndRun()
}
