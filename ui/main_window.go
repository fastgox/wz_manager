package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/luoxk/wzlib"
)

// MainWindow 主窗口结构
type MainWindow struct {
	window       fyne.Window
	fileManager  *FileManager
	treeViewer   *TreeViewer
	imageViewer  *ImageViewer
	soundPlayer  *SoundPlayer
	dataExporter *DataExporter
	content      *container.Split
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

	// 当树视图选择节点时，更新相应的查看器
	mw.treeViewer.OnNodeSelected = func(nodeType string, nodeValue interface{}) {
		switch nodeType {
		case "Canvas":
			mw.imageViewer.ShowImage(nodeValue)
		case "Sound":
			mw.soundPlayer.LoadSound(nodeValue)
		default:
			// 显示节点信息
		}
	}
}

// createLayout 创建界面布局
func (mw *MainWindow) createLayout() {
	// 左侧面板：文件管理和树视图
	leftPanel := container.NewVSplit(
		mw.fileManager.GetContent(),
		mw.treeViewer.GetContent(),
	)
	leftPanel.SetOffset(0.3)

	// 右侧面板：图像查看器、音频播放器等
	rightTabs := container.NewAppTabs(
		container.NewTabItem("图像查看", mw.imageViewer.GetContent()),
		container.NewTabItem("音频播放", mw.soundPlayer.GetContent()),
		container.NewTabItem("数据导出", mw.dataExporter.GetContent()),
	)

	// 主分割面板
	mw.content = container.NewHSplit(leftPanel, rightTabs)
	mw.content.SetOffset(0.3)
}

// GetContent 获取主窗口内容
func (mw *MainWindow) GetContent() fyne.CanvasObject {
	return mw.content
}
