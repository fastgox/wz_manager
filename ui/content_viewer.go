package ui

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/luoxk/wzlib"
)

// ContentViewer 通用内容查看器
type ContentViewer struct {
	content     *fyne.Container
	tabs        *container.AppTabs
	infoTab     *container.TabItem
	dataTab     *container.TabItem
	hexTab      *container.TabItem
	currentNode *wzlib.WzNode
}

// NewContentViewer 创建新的内容查看器
func NewContentViewer() *ContentViewer {
	cv := &ContentViewer{}
	cv.createContent()
	return cv
}

// createContent 创建内容
func (cv *ContentViewer) createContent() {
	// 创建选项卡
	cv.tabs = container.NewAppTabs()

	// 信息选项卡
	cv.infoTab = container.NewTabItem("ℹ️ 节点信息", widget.NewLabel("请选择一个节点"))
	cv.tabs.Append(cv.infoTab)

	// 数据选项卡
	cv.dataTab = container.NewTabItem("📊 数据内容", widget.NewLabel("无数据"))
	cv.tabs.Append(cv.dataTab)

	// 十六进制选项卡
	cv.hexTab = container.NewTabItem("🔢 十六进制", widget.NewLabel("无数据"))
	cv.tabs.Append(cv.hexTab)

	cv.content = container.NewBorder(nil, nil, nil, nil, cv.tabs)
}

// ShowNodeContent 显示节点内容
func (cv *ContentViewer) ShowNodeContent(nodeType string, nodeValue interface{}, node *wzlib.WzNode) {
	cv.currentNode = node

	// 更新信息选项卡
	cv.updateInfoTab(nodeType, nodeValue, node)

	// 更新数据选项卡
	cv.updateDataTab(nodeType, nodeValue, node)

	// 更新十六进制选项卡
	cv.updateHexTab(nodeType, nodeValue, node)
}

// updateInfoTab 更新信息选项卡
func (cv *ContentViewer) updateInfoTab(nodeType string, nodeValue interface{}, node *wzlib.WzNode) {
	var infoText strings.Builder

	// 基本信息
	infoText.WriteString(fmt.Sprintf("节点名称: %s\n", node.Text))
	infoText.WriteString(fmt.Sprintf("节点类型: %s\n", nodeType))
	infoText.WriteString(fmt.Sprintf("完整路径: %s\n", node.GetFullPath()))
	infoText.WriteString(fmt.Sprintf("子节点数量: %d\n", len(node.Nodes)))

	// 父节点信息
	if node.ParentNode != nil {
		infoText.WriteString(fmt.Sprintf("父节点: %s\n", node.ParentNode.Text))
	} else {
		infoText.WriteString("父节点: 无\n")
	}

	// 值类型信息
	if nodeValue != nil {
		valueType := reflect.TypeOf(nodeValue)
		infoText.WriteString(fmt.Sprintf("值类型: %s\n", valueType.String()))

		// 根据类型显示特定信息
		switch v := nodeValue.(type) {
		case *wzlib.WzImage:
			infoText.WriteString(fmt.Sprintf("图像是否已提取: %t\n", v.Extracted))
		case *wzlib.WzSound:
			infoText.WriteString("音频文件\n")
		case string:
			infoText.WriteString(fmt.Sprintf("字符串长度: %d\n", len(v)))
		case int, int32, int64:
			infoText.WriteString(fmt.Sprintf("数值: %v\n", v))
		case float32, float64:
			infoText.WriteString(fmt.Sprintf("浮点数: %v\n", v))
		}
	} else {
		infoText.WriteString("值类型: 无\n")
	}

	// 创建滚动的文本显示
	infoLabel := widget.NewLabel(infoText.String())
	infoLabel.Wrapping = fyne.TextWrapWord
	scrollInfo := container.NewScroll(infoLabel)

	cv.infoTab.Content = scrollInfo
	cv.tabs.Refresh()
}

// updateDataTab 更新数据选项卡
func (cv *ContentViewer) updateDataTab(nodeType string, nodeValue interface{}, node *wzlib.WzNode) {
	var content fyne.CanvasObject

	switch nodeType {
	case "Canvas":
		// 图像数据
		content = cv.createImageContent(nodeValue)
	case "Sound":
		// 音频数据
		content = cv.createSoundContent(nodeValue)
	case "String":
		// 字符串数据
		content = cv.createStringContent(nodeValue)
	case "Int", "Short", "Long":
		// 数值数据
		content = cv.createNumberContent(nodeValue)
	case "Float", "Double":
		// 浮点数据
		content = cv.createFloatContent(nodeValue)
	case "Vector":
		// 向量数据
		content = cv.createVectorContent(nodeValue)
	default:
		// 通用数据显示
		content = cv.createGenericContent(nodeValue, node)
	}

	cv.dataTab.Content = content
	cv.tabs.Refresh()
}

// createImageContent 创建图像内容
func (cv *ContentViewer) createImageContent(nodeValue interface{}) fyne.CanvasObject {
	if img, ok := nodeValue.(*wzlib.WzImage); ok {
		var infoText strings.Builder
		infoText.WriteString("图像信息:\n")
		infoText.WriteString(fmt.Sprintf("是否已提取: %t\n", img.Extracted))

		// 尝试提取图像
		if !img.Extracted {
			img.TryExtract()
		}

		// 显示图像属性
		if img.Extracted {
			infoText.WriteString("图像已提取，可在图像查看选项卡中查看\n")
		} else {
			infoText.WriteString("图像提取失败\n")
		}

		label := widget.NewLabel(infoText.String())
		return container.NewScroll(label)
	}

	return widget.NewLabel("无效的图像数据")
}

// createSoundContent 创建音频内容
func (cv *ContentViewer) createSoundContent(nodeValue interface{}) fyne.CanvasObject {
	if sound, ok := nodeValue.(*wzlib.WzSound); ok {
		var infoText strings.Builder
		infoText.WriteString("音频信息:\n")
		infoText.WriteString(fmt.Sprintf("音频对象: %v\n", sound))
		infoText.WriteString("可在音频播放选项卡中播放\n")

		label := widget.NewLabel(infoText.String())
		return container.NewScroll(label)
	}

	return widget.NewLabel("无效的音频数据")
}

// createStringContent 创建字符串内容
func (cv *ContentViewer) createStringContent(nodeValue interface{}) fyne.CanvasObject {
	if str, ok := nodeValue.(string); ok {
		var infoText strings.Builder
		infoText.WriteString(fmt.Sprintf("字符串内容:\n%s\n\n", str))
		infoText.WriteString(fmt.Sprintf("长度: %d 字符\n", len(str)))
		infoText.WriteString(fmt.Sprintf("字节数: %d\n", len([]byte(str))))

		label := widget.NewLabel(infoText.String())
		label.Wrapping = fyne.TextWrapWord
		return container.NewScroll(label)
	}

	return widget.NewLabel("无效的字符串数据")
}

// createNumberContent 创建数值内容
func (cv *ContentViewer) createNumberContent(nodeValue interface{}) fyne.CanvasObject {
	var infoText strings.Builder
	infoText.WriteString("数值信息:\n")

	switch v := nodeValue.(type) {
	case int:
		infoText.WriteString(fmt.Sprintf("整数值: %d\n", v))
		infoText.WriteString(fmt.Sprintf("十六进制: 0x%X\n", v))
		infoText.WriteString(fmt.Sprintf("二进制: %b\n", v))
	case int32:
		infoText.WriteString(fmt.Sprintf("32位整数: %d\n", v))
		infoText.WriteString(fmt.Sprintf("十六进制: 0x%X\n", v))
	case int64:
		infoText.WriteString(fmt.Sprintf("64位整数: %d\n", v))
		infoText.WriteString(fmt.Sprintf("十六进制: 0x%X\n", v))
	default:
		infoText.WriteString(fmt.Sprintf("数值: %v\n", v))
	}

	label := widget.NewLabel(infoText.String())
	return container.NewScroll(label)
}

// createFloatContent 创建浮点数内容
func (cv *ContentViewer) createFloatContent(nodeValue interface{}) fyne.CanvasObject {
	var infoText strings.Builder
	infoText.WriteString("浮点数信息:\n")

	switch v := nodeValue.(type) {
	case float32:
		infoText.WriteString(fmt.Sprintf("32位浮点数: %f\n", v))
		infoText.WriteString(fmt.Sprintf("科学计数法: %e\n", v))
	case float64:
		infoText.WriteString(fmt.Sprintf("64位浮点数: %f\n", v))
		infoText.WriteString(fmt.Sprintf("科学计数法: %e\n", v))
	default:
		infoText.WriteString(fmt.Sprintf("浮点数: %v\n", v))
	}

	label := widget.NewLabel(infoText.String())
	return container.NewScroll(label)
}

// createVectorContent 创建向量内容
func (cv *ContentViewer) createVectorContent(nodeValue interface{}) fyne.CanvasObject {
	var infoText strings.Builder
	infoText.WriteString("向量信息:\n")
	infoText.WriteString(fmt.Sprintf("向量数据: %v\n", nodeValue))

	label := widget.NewLabel(infoText.String())
	return container.NewScroll(label)
}

// createGenericContent 创建通用内容
func (cv *ContentViewer) createGenericContent(nodeValue interface{}, node *wzlib.WzNode) fyne.CanvasObject {
	var infoText strings.Builder

	if nodeValue != nil {
		infoText.WriteString("数据内容:\n")
		infoText.WriteString(fmt.Sprintf("%v\n\n", nodeValue))

		// 使用反射显示详细信息
		value := reflect.ValueOf(nodeValue)
		infoText.WriteString(fmt.Sprintf("类型: %s\n", value.Type().String()))
		infoText.WriteString(fmt.Sprintf("种类: %s\n", value.Kind().String()))
	} else {
		infoText.WriteString("此节点没有数据内容\n")
	}

	// 如果有子节点，显示子节点列表
	if len(node.Nodes) > 0 {
		infoText.WriteString(fmt.Sprintf("\n子节点列表 (%d个):\n", len(node.Nodes)))
		for i, child := range node.Nodes {
			if i < 20 { // 只显示前20个子节点
				infoText.WriteString(fmt.Sprintf("- %s [%s]\n", child.Text, child.Type))
			} else {
				infoText.WriteString(fmt.Sprintf("... 还有 %d 个子节点\n", len(node.Nodes)-20))
				break
			}
		}
	}

	label := widget.NewLabel(infoText.String())
	label.Wrapping = fyne.TextWrapWord
	return container.NewScroll(label)
}

// updateHexTab 更新十六进制选项卡
func (cv *ContentViewer) updateHexTab(nodeType string, nodeValue interface{}, node *wzlib.WzNode) {
	var content fyne.CanvasObject

	if nodeValue != nil {
		// 尝试将数据转换为字节数组进行十六进制显示
		var data []byte

		switch v := nodeValue.(type) {
		case string:
			data = []byte(v)
		case []byte:
			data = v
		case int:
			data = []byte(strconv.Itoa(v))
		case int32:
			data = []byte(strconv.FormatInt(int64(v), 10))
		case int64:
			data = []byte(strconv.FormatInt(v, 10))
		case float32:
			data = []byte(strconv.FormatFloat(float64(v), 'f', -1, 32))
		case float64:
			data = []byte(strconv.FormatFloat(v, 'f', -1, 64))
		default:
			data = []byte(fmt.Sprintf("%v", v))
		}

		if len(data) > 0 {
			hexText := cv.formatHexDump(data)
			label := widget.NewLabel(hexText)
			label.TextStyle.Monospace = true
			content = container.NewScroll(label)
		} else {
			content = widget.NewLabel("无法转换为十六进制显示")
		}
	} else {
		content = widget.NewLabel("无数据")
	}

	cv.hexTab.Content = content
	cv.tabs.Refresh()
}

// formatHexDump 格式化十六进制转储
func (cv *ContentViewer) formatHexDump(data []byte) string {
	var result strings.Builder

	for i := 0; i < len(data); i += 16 {
		// 地址
		result.WriteString(fmt.Sprintf("%08X: ", i))

		// 十六进制字节
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				result.WriteString(fmt.Sprintf("%02X ", data[i+j]))
			} else {
				result.WriteString("   ")
			}

			if j == 7 {
				result.WriteString(" ")
			}
		}

		result.WriteString(" |")

		// ASCII 字符
		for j := 0; j < 16 && i+j < len(data); j++ {
			b := data[i+j]
			if b >= 32 && b <= 126 {
				result.WriteString(string(b))
			} else {
				result.WriteString(".")
			}
		}

		result.WriteString("|\n")

		// 限制显示长度，避免界面卡顿
		if i > 1024 {
			result.WriteString("... (数据过长，已截断)\n")
			break
		}
	}

	return result.String()
}

// GetContent 获取内容
func (cv *ContentViewer) GetContent() fyne.CanvasObject {
	return cv.content
}
