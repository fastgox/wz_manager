package ui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/luoxk/wzlib"
)

// FileManager 文件管理器
type FileManager struct {
	content         *fyne.Container
	fileList        *widget.List
	loadedFiles     []string
	wzStructures    map[string]*wzlib.WzStructure
	mergedStructure *wzlib.WzStructure
	OnWzFileLoaded  func(wzStructure interface{})
	statusLabel     *widget.Label
}

// NewFileManager 创建新的文件管理器
func NewFileManager() *FileManager {
	fm := &FileManager{
		loadedFiles:  make([]string, 0),
		wzStructures: make(map[string]*wzlib.WzStructure),
		statusLabel:  widget.NewLabel("正在加载默认数据..."),
	}

	fm.createContent()

	// 自动加载 ./data 目录下的 WZ 文件
	go fm.loadDefaultDataFiles()

	return fm
}

// createContent 创建文件管理器内容
func (fm *FileManager) createContent() {
	// 文件列表
	fm.fileList = widget.NewList(
		func() int {
			return len(fm.loadedFiles)
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("文件名")
			// 设置文件列表项的最小高度，更紧凑
			label.Resize(fyne.NewSize(150, 22))
			return label
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id < len(fm.loadedFiles) {
				fileName := filepath.Base(fm.loadedFiles[id])
				label.SetText(fileName)
			}
		},
	)

	// 文件列表选择事件
	fm.fileList.OnSelected = func(id widget.ListItemID) {
		if id < len(fm.loadedFiles) {
			filePath := fm.loadedFiles[id]
			if _, exists := fm.wzStructures[filePath]; exists {
				// 总是传递合并后的结构，而不是单个文件的结构
				if fm.mergedStructure != nil && fm.OnWzFileLoaded != nil {
					fm.OnWzFileLoaded(fm.mergedStructure)
				}
				fm.statusLabel.SetText(fmt.Sprintf("Selected: %s", filepath.Base(filePath)))
			}
		}
	}

	// 创建紧凑的按钮（只显示图标，减少文字）
	loadButton := widget.NewButtonWithIcon("加载", theme.FolderOpenIcon(), fm.loadWzFile)
	loadButton.Importance = widget.HighImportance

	removeButton := widget.NewButtonWithIcon("移除", theme.DeleteIcon(), fm.removeSelectedFile)
	removeButton.Importance = widget.MediumImportance

	clearButton := widget.NewButtonWithIcon("清空", theme.ContentClearIcon(), fm.clearFileList)
	clearButton.Importance = widget.LowImportance

	// 使用垂直布局让按钮更紧凑
	buttonContainer := container.NewVBox(
		container.NewHBox(loadButton, removeButton),
		clearButton,
	)

	// 创建标题标签
	titleLabel := widget.NewLabel("📂 WZ文件管理")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	// 组装内容 - 使用 Border 布局让文件列表占用剩余全部高度
	topContent := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		buttonContainer,
		widget.NewSeparator(),
		fm.statusLabel,
	)

	fm.content = container.NewBorder(
		topContent,                       // 顶部：标题、按钮、状态标签
		nil,                              // 底部：无
		nil,                              // 左侧：无
		nil,                              // 右侧：无
		container.NewScroll(fm.fileList), // 中心：文件列表占用剩余全部空间
	)
}

// loadWzFile 加载WZ文件
func (fm *FileManager) loadWzFile() {
	// 创建文件选择对话框
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		filePath := reader.URI().Path()

		log.Printf("Attempting to load WZ file: %s", filePath)

		// 检查文件是否已经加载
		for _, loadedFile := range fm.loadedFiles {
			if loadedFile == filePath {
				fm.statusLabel.SetText("File already loaded")
				return
			}
		}

		// 加载WZ文件
		wzStructure := &wzlib.WzStructure{}
		log.Printf("Created WzStructure, calling LoadWzFile...")
		loadErr := wzStructure.LoadWzFile(filePath)
		if loadErr != nil {
			log.Printf("LoadWzFile failed: %v", loadErr)
			dialog.ShowError(fmt.Errorf("Failed to load WZ file: %v", loadErr), fyne.CurrentApp().Driver().AllWindows()[0])
			fm.statusLabel.SetText("Load failed")
			return
		}

		log.Printf("LoadWzFile succeeded")

		// 检查是否成功加载
		if wzStructure.WzNode == nil {
			log.Printf("WzNode is nil")
			dialog.ShowError(fmt.Errorf("WZ file loaded but no data found"), fyne.CurrentApp().Driver().AllWindows()[0])
			fm.statusLabel.SetText("No data found")
			return
		}

		log.Printf("WzNode loaded successfully, has %d child nodes", len(wzStructure.WzNode.Nodes))

		// 添加到列表
		fm.loadedFiles = append(fm.loadedFiles, filePath)
		fm.wzStructures[filePath] = wzStructure
		fm.fileList.Refresh()

		// 自动选择新加载的文件
		fm.fileList.Select(len(fm.loadedFiles) - 1)

		fm.statusLabel.SetText(fmt.Sprintf("Successfully loaded: %s", filepath.Base(filePath)))

	}, fyne.CurrentApp().Driver().AllWindows()[0])

	// 设置文件过滤器
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".wz"}))
	fileDialog.Show()
}

// removeSelectedFile 移除选中的文件
func (fm *FileManager) removeSelectedFile() {
	// 获取当前选中的项目
	selectedID := -1
	if len(fm.loadedFiles) > 0 {
		// 简单实现：移除最后选中的文件
		selectedID = len(fm.loadedFiles) - 1
	}

	if selectedID < 0 || selectedID >= len(fm.loadedFiles) {
		return
	}

	// 移除选中的文件
	filePath := fm.loadedFiles[selectedID]
	delete(fm.wzStructures, filePath)

	// 重建文件列表
	newFiles := make([]string, 0)
	for i, file := range fm.loadedFiles {
		if i != selectedID {
			newFiles = append(newFiles, file)
		}
	}

	fm.loadedFiles = newFiles
	fm.fileList.UnselectAll()
	fm.fileList.Refresh()
	fm.statusLabel.SetText("Removed selected file")
}

// clearFileList 清空文件列表
func (fm *FileManager) clearFileList() {
	fm.loadedFiles = make([]string, 0)
	fm.wzStructures = make(map[string]*wzlib.WzStructure)
	fm.fileList.UnselectAll()
	fm.fileList.Refresh()
	fm.statusLabel.SetText("File list cleared")
}

// loadDefaultDataFiles 自动加载 ./data 目录下的 WZ 文件
func (fm *FileManager) loadDefaultDataFiles() {
	dataDir := "./data"

	// 检查 data 目录是否存在
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fm.statusLabel.SetText("data 目录不存在")
		return
	}

	// 读取 data 目录中的文件
	files, err := os.ReadDir(dataDir)
	if err != nil {
		log.Printf("读取 data 目录失败: %v", err)
		fm.statusLabel.SetText("读取 data 目录失败")
		return
	}

	loadedCount := 0
	totalWzFiles := 0

	// 统计 .wz 文件数量
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".wz") {
			totalWzFiles++
		}
	}

	if totalWzFiles == 0 {
		fm.statusLabel.SetText("data 目录中没有找到 .wz 文件")
		return
	}

	// 加载每个 .wz 文件
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".wz") {
			filePath := filepath.Join(dataDir, file.Name())

			log.Printf("正在加载 WZ 文件: %s", filePath)
			fm.statusLabel.SetText(fmt.Sprintf("正在加载: %s (%d/%d)", file.Name(), loadedCount+1, totalWzFiles))

			// 检查文件是否已经加载
			alreadyLoaded := false
			for _, loadedFile := range fm.loadedFiles {
				if loadedFile == filePath {
					alreadyLoaded = true
					break
				}
			}

			if alreadyLoaded {
				continue
			}

			// 加载 WZ 文件
			wzStructure := &wzlib.WzStructure{}
			loadErr := wzStructure.LoadWzFile(filePath)
			if loadErr != nil {
				log.Printf("加载 WZ 文件失败 %s: %v", filePath, loadErr)
				continue
			}

			// 检查是否成功加载
			if wzStructure.WzNode == nil {
				log.Printf("WZ 文件加载成功但没有数据: %s", filePath)
				continue
			}

			log.Printf("成功加载 WZ 文件: %s，包含 %d 个子节点", filePath, len(wzStructure.WzNode.Nodes))

			// 添加到列表
			fm.loadedFiles = append(fm.loadedFiles, filePath)
			fm.wzStructures[filePath] = wzStructure
			loadedCount++

			// 刷新界面
			fm.fileList.Refresh()
		}
	}

	// 更新状态
	if loadedCount > 0 {
		fm.statusLabel.SetText(fmt.Sprintf("成功加载 %d 个 WZ 文件", loadedCount))

		// 创建合并的WZ结构
		fm.createMergedStructure()

		// 自动选择第一个文件
		if len(fm.loadedFiles) > 0 {
			fm.fileList.Select(0)
		}
	} else {
		fm.statusLabel.SetText("没有成功加载任何 WZ 文件")
	}
}

// createMergedStructure 创建合并的WZ结构
func (fm *FileManager) createMergedStructure() {
	if len(fm.wzStructures) == 0 {
		return
	}

	// 创建根节点
	rootNode := &wzlib.WzNode{
		Text:  "所有WZ文件",
		Type:  "Directory",
		Nodes: make([]*wzlib.WzNode, 0),
	}

	// 将所有WZ文件的根节点添加到合并结构中
	for filePath, wzStructure := range fm.wzStructures {
		if wzStructure.WzNode != nil {
			// 创建文件节点
			fileName := filepath.Base(filePath)
			fileNode := &wzlib.WzNode{
				Text:       fileName,
				Type:       "WzFile",
				Nodes:      wzStructure.WzNode.Nodes,
				ParentNode: rootNode,
			}

			// 更新子节点的父节点引用
			for _, child := range fileNode.Nodes {
				child.ParentNode = fileNode
			}

			rootNode.Nodes = append(rootNode.Nodes, fileNode)
		}
	}

	// 创建合并的WZ结构
	fm.mergedStructure = &wzlib.WzStructure{
		WzNode: rootNode,
	}

	// 通知界面更新
	if fm.OnWzFileLoaded != nil {
		fm.OnWzFileLoaded(fm.mergedStructure)
	}
}

// GetContent 获取文件管理器内容
func (fm *FileManager) GetContent() fyne.CanvasObject {
	return fm.content
}
