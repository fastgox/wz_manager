package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/luoxk/wzlib"
)

// MainWindow 主窗口结构
type MainWindow struct {
	window        fyne.Window
	fileManager   *FileManager
	treeViewer    *TreeViewer
	imageViewer   *ImageViewer
	soundPlayer   *SoundPlayer
	dataExporter  *DataExporter
	contentViewer *ContentViewer
	content       *fyne.Container
	statusBar     *widget.Label
}

// NewMainWindow 创建新的主窗口
func NewMainWindow(window fyne.Window) *MainWindow {
	mw := &MainWindow{
		window: window,
	}

	// 初始化各个组件
	mw.fileManager = NewFileManager()
	mw.treeViewer = NewTreeViewer()
	mw.imageViewer = NewImageViewer()
	mw.soundPlayer = NewSoundPlayer()
	mw.dataExporter = NewDataExporter()
	mw.contentViewer = NewContentViewer()

	// 创建现代风格状态栏
	mw.statusBar = widget.NewLabel("🚀 WZ文件管理器已启动 - 准备就绪")
	mw.statusBar.TextStyle = fyne.TextStyle{Italic: true}

	// 设置组件间的通信
	mw.setupConnections()

	// 创建界面布局
	mw.createLayout()

	return mw
}

// setupConnections 设置组件间的通信
func (mw *MainWindow) setupConnections() {
	// 当文件管理器加载WZ文件时，更新树视图和数据导出器
	mw.fileManager.OnWzFileLoaded = func(wzStructure interface{}) {
		mw.treeViewer.LoadWzStructure(wzStructure)
		if ws, ok := wzStructure.(*wzlib.WzStructure); ok {
			mw.dataExporter.SetWzStructure(ws)
		}
	}

	// 设置树视图的文件加载回调和文件管理器引用
	mw.treeViewer.OnWzFileLoaded = mw.fileManager.OnWzFileLoaded
	mw.treeViewer.fileManager = mw.fileManager

	// 当树视图选择节点时，更新相应的查看器
	mw.treeViewer.OnNodeSelected = func(nodeType string, nodeValue interface{}, node *wzlib.WzNode) {
		// 更新内容查看器
		mw.contentViewer.ShowNodeContent(nodeType, nodeValue, node)

		// 更新状态栏
		statusMsg := fmt.Sprintf("📍 已选择: %s [%s]", node.Text, nodeType)
		if nodeType == "Canvas" {
			statusMsg += " 🖼️"
		} else if nodeType == "Sound" {
			statusMsg += " 🎵"
		}
		mw.UpdateStatusBar(statusMsg)

		// 根据类型更新特定查看器
		switch nodeType {
		case "Canvas":
			mw.imageViewer.ShowImage(nodeValue)
		case "Sound":
			mw.soundPlayer.LoadSound(nodeValue)
		}
	}
}

// createLayout 创建界面布局
func (mw *MainWindow) createLayout() {
	// 左侧面板：只显示树视图
	leftPanel := mw.treeViewer.GetContent()

	// 右侧面板：现代风格的选项卡
	rightTabs := container.NewAppTabs(
		container.NewTabItem("📄 内容查看", container.NewPadded(mw.contentViewer.GetContent())),
		container.NewTabItem("🖼️ 图像查看", container.NewPadded(mw.imageViewer.GetContent())),
		container.NewTabItem("🎵 音频播放", container.NewPadded(mw.soundPlayer.GetContent())),
		container.NewTabItem("📤 数据导出", container.NewPadded(mw.dataExporter.GetContent())),
	)

	// 设置选项卡位置为顶部
	rightTabs.SetTabLocation(container.TabLocationTop)

	// 主分割面板，左侧面板占用更少空间
	mainSplit := container.NewHSplit(leftPanel, rightTabs)
	mainSplit.SetOffset(0.25)

	// 创建状态栏卡片
	statusBarCard := widget.NewCard("", "", container.NewPadded(mw.statusBar))

	// 使用Border布局，底部放状态栏
	mw.content = container.NewBorder(
		nil,           // 顶部
		statusBarCard, // 底部：状态栏卡片
		nil,           // 左侧
		nil,           // 右侧
		mainSplit,     // 中心：主要内容
	)
}

// UpdateStatusBar 更新状态栏
func (mw *MainWindow) UpdateStatusBar(message string) {
	if mw.statusBar != nil {
		mw.statusBar.SetText(message)
	}
}

// GetContent 获取主窗口内容
func (mw *MainWindow) GetContent() fyne.CanvasObject {
	return mw.content
}
