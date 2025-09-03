package ui

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/luoxk/wzlib"
)

// DataExporter 数据导出器
type DataExporter struct {
	content          *fyne.Container
	wzStructure      *wzlib.WzStructure
	exportTypeSelect *widget.Select
	pathEntry        *widget.Entry
	exportButton     *widget.Button
	progressBar      *widget.ProgressBar
	statusLabel      *widget.Label
	filterEntry      *widget.Entry
}

// NewDataExporter 创建新的数据导出器
func NewDataExporter() *DataExporter {
	de := &DataExporter{
		statusLabel: widget.NewLabel("No WZ file loaded"),
	}

	de.createContent()
	return de
}

// createContent 创建数据导出器内容
func (de *DataExporter) createContent() {
	// 路径过滤
	de.filterEntry = widget.NewEntry()
	de.filterEntry.SetPlaceHolder("Path filter (e.g.: *.img, Skill/*, leave empty for all)")

	// 导出类型选择
	de.exportTypeSelect = widget.NewSelect(
		[]string{"JSON", "XML", "CSV", "Image Files", "Audio Files"},
		de.onExportTypeChanged,
	)
	de.exportTypeSelect.SetSelected("JSON")

	// 导出路径
	de.pathEntry = widget.NewEntry()
	de.pathEntry.SetPlaceHolder("Select export path...")

	browseButton := widget.NewButton("Browse", de.browseExportPath)
	pathContainer := container.NewBorder(nil, nil, nil, browseButton, de.pathEntry)

	// 导出按钮
	de.exportButton = widget.NewButton("Start Export", de.startExport)
	de.exportButton.Disable()

	// 进度条
	de.progressBar = widget.NewProgressBar()
	de.progressBar.Hide()

	// 组装内容
	de.content = container.NewVBox(
		widget.NewLabel("Data Exporter"),
		widget.NewSeparator(),

		widget.NewLabel("Export Type:"),
		de.exportTypeSelect,

		widget.NewLabel("Path Filter:"),
		de.filterEntry,

		widget.NewLabel("Export Path:"),
		pathContainer,

		widget.NewSeparator(),
		de.exportButton,
		de.progressBar,
		de.statusLabel,

		widget.NewSeparator(),
		widget.NewLabel("Supported Export Formats:"),
		widget.NewLabel("• JSON: Export node structure and data"),
		widget.NewLabel("• XML: Export as XML format"),
		widget.NewLabel("• CSV: Export as table format"),
		widget.NewLabel("• Image Files: Export all images as PNG"),
		widget.NewLabel("• Audio Files: Export all audio files"),
	)
}

// SetWzStructure 设置WZ结构
func (de *DataExporter) SetWzStructure(wzStructure *wzlib.WzStructure) {
	de.wzStructure = wzStructure
	if wzStructure != nil {
		de.statusLabel.SetText("WZ file loaded, ready to export")
		if de.pathEntry.Text != "" {
			de.exportButton.Enable()
		}
	} else {
		de.statusLabel.SetText("No WZ file loaded")
		de.exportButton.Disable()
	}
}

// onExportTypeChanged 导出类型变化事件
func (de *DataExporter) onExportTypeChanged(selected string) {
	// 根据选择的类型更新界面
	switch selected {
	case "Image Files":
		de.filterEntry.SetText("Canvas")
		de.filterEntry.SetPlaceHolder("Will export all Canvas type nodes")
	case "Audio Files":
		de.filterEntry.SetText("Sound")
		de.filterEntry.SetPlaceHolder("Will export all Sound type nodes")
	default:
		de.filterEntry.SetPlaceHolder("Path filter (e.g.: *.img, Skill/*, leave empty for all)")
	}
}

// browseExportPath 浏览导出路径
func (de *DataExporter) browseExportPath() {
	folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		de.pathEntry.SetText(uri.Path())
		if de.wzStructure != nil {
			de.exportButton.Enable()
		}
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	folderDialog.Show()
}

// startExport 开始导出
func (de *DataExporter) startExport() {
	if de.wzStructure == nil || de.pathEntry.Text == "" {
		return
	}

	exportType := de.exportTypeSelect.Selected
	exportPath := de.pathEntry.Text
	filter := de.filterEntry.Text

	// 显示进度条
	de.progressBar.Show()
	de.progressBar.SetValue(0)
	de.exportButton.Disable()
	de.statusLabel.SetText("正在导出...")

	// 在goroutine中执行导出
	go func() {
		err := de.performExport(exportType, exportPath, filter)

		// 更新UI (需要在主线程中执行)
		de.progressBar.Hide()
		de.exportButton.Enable()

		if err != nil {
			de.statusLabel.SetText(fmt.Sprintf("导出失败: %v", err))
			dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		} else {
			de.statusLabel.SetText("导出完成")
			dialog.ShowInformation("导出完成",
				fmt.Sprintf("数据已成功导出到: %s", exportPath),
				fyne.CurrentApp().Driver().AllWindows()[0])
		}
	}()
}

// performExport 执行导出
func (de *DataExporter) performExport(exportType, exportPath, filter string) error {
	switch exportType {
	case "JSON":
		return de.exportToJSON(exportPath, filter)
	case "XML":
		return de.exportToXML(exportPath, filter)
	case "CSV":
		return de.exportToCSV(exportPath, filter)
	case "图像文件":
		return de.exportImages(exportPath, filter)
	case "音频文件":
		return de.exportSounds(exportPath, filter)
	default:
		return fmt.Errorf("不支持的导出类型: %s", exportType)
	}
}

// exportToJSON 导出为JSON
func (de *DataExporter) exportToJSON(exportPath, filter string) error {
	data := de.collectNodeData(de.wzStructure.WzNode, filter)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON编码失败: %v", err)
	}

	fileName := filepath.Join(exportPath, "wz_data.json")
	return de.writeFile(fileName, jsonData)
}

// exportToXML 导出为XML
func (de *DataExporter) exportToXML(exportPath, filter string) error {
	// TODO: 实现XML导出
	return fmt.Errorf("XML导出功能尚未实现")
}

// exportToCSV 导出为CSV
func (de *DataExporter) exportToCSV(exportPath, filter string) error {
	// TODO: 实现CSV导出
	return fmt.Errorf("CSV导出功能尚未实现")
}

// exportImages 导出图像文件
func (de *DataExporter) exportImages(exportPath, filter string) error {
	// TODO: 实现图像导出
	return fmt.Errorf("图像导出功能尚未实现")
}

// exportSounds 导出音频文件
func (de *DataExporter) exportSounds(exportPath, filter string) error {
	// TODO: 实现音频导出
	return fmt.Errorf("音频导出功能尚未实现")
}

// collectNodeData 收集节点数据
func (de *DataExporter) collectNodeData(node *wzlib.WzNode, filter string) map[string]interface{} {
	if node == nil {
		return nil
	}

	// 简单的过滤逻辑
	if filter != "" && !de.matchFilter(node, filter) {
		return nil
	}

	data := map[string]interface{}{
		"name": node.Text,
		"type": node.Type,
		"path": node.GetFullPath(),
	}

	// 添加值信息
	if node.Value != nil {
		data["value_type"] = fmt.Sprintf("%T", node.Value)
	}

	// 递归处理子节点
	if len(node.Nodes) > 0 {
		children := make(map[string]interface{})
		for _, child := range node.Nodes {
			childData := de.collectNodeData(child, filter)
			if childData != nil {
				children[child.Text] = childData
			}
		}
		if len(children) > 0 {
			data["children"] = children
		}
	}

	return data
}

// matchFilter 匹配过滤条件
func (de *DataExporter) matchFilter(node *wzlib.WzNode, filter string) bool {
	if filter == "" {
		return true
	}

	// 简单的匹配逻辑
	path := node.GetFullPath()
	name := node.Text
	nodeType := node.Type

	// 类型匹配
	if filter == nodeType {
		return true
	}

	// 路径匹配
	if strings.Contains(path, filter) {
		return true
	}

	// 名称匹配
	if strings.Contains(name, filter) {
		return true
	}

	// 通配符匹配 (简单实现)
	if strings.HasSuffix(filter, "*") {
		prefix := strings.TrimSuffix(filter, "*")
		return strings.HasPrefix(name, prefix) || strings.HasPrefix(path, prefix)
	}

	return false
}

// writeFile 写入文件
func (de *DataExporter) writeFile(fileName string, data []byte) error {
	// TODO: 实现文件写入
	return fmt.Errorf("文件写入功能尚未实现")
}

// GetContent 获取数据导出器内容
func (de *DataExporter) GetContent() fyne.CanvasObject {
	return de.content
}
