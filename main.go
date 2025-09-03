package main

import (
	"wz_manager/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	myApp := app.New()

	// 设置支持中文的主题
	myApp.Settings().SetTheme(ui.NewChineseTheme())

	mainWindow := myApp.NewWindow("WZ文件管理器")
	mainWindow.Resize(fyne.NewSize(1200, 800))

	// 创建主界面
	mainUI := ui.NewMainWindow(mainWindow)
	mainWindow.SetContent(mainUI.GetContent())

	mainWindow.ShowAndRun()
}
