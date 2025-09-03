package ui

import (
	"fmt"
	"log"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/luoxk/wzlib"
)

// FileManager 文件管理器
type FileManager struct {
	content        *fyne.Container
	fileList       *widget.List
	loadedFiles    []string
	wzStructures   map[string]*wzlib.WzStructure
	OnWzFileLoaded func(wzStructure interface{})
	statusLabel    *widget.Label
}

// NewFileManager 创建新的文件管理器
func NewFileManager() *FileManager {
	fm := &FileManager{
		loadedFiles:  make([]string, 0),
		wzStructures: make(map[string]*wzlib.WzStructure),
		statusLabel:  widget.NewLabel("未加载文件"),
	}

	fm.createContent()
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
			return widget.NewLabel("文件名")
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
			if wzStructure, exists := fm.wzStructures[filePath]; exists {
				if fm.OnWzFileLoaded != nil {
					fm.OnWzFileLoaded(wzStructure)
				}
				fm.statusLabel.SetText(fmt.Sprintf("Selected: %s", filepath.Base(filePath)))
			}
		}
	}

	// 按钮
	loadButton := widget.NewButton("加载WZ文件", fm.loadWzFile)
	removeButton := widget.NewButton("移除文件", fm.removeSelectedFile)
	clearButton := widget.NewButton("清空列表", fm.clearFileList)

	buttonContainer := container.NewHBox(loadButton, removeButton, clearButton)

	// 组装内容
	fm.content = container.NewVBox(
		widget.NewLabel("WZ文件管理"),
		widget.NewSeparator(),
		buttonContainer,
		widget.NewSeparator(),
		fm.statusLabel,
		fm.fileList,
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

// GetContent 获取文件管理器内容
func (fm *FileManager) GetContent() fyne.CanvasObject {
	return fm.content
}
