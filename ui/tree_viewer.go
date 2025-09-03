package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/luoxk/wzlib"
)

// TreeViewer WZ文件树视图
type TreeViewer struct {
	content        *fyne.Container
	tree           *widget.Tree
	wzStructure    *wzlib.WzStructure
	OnNodeSelected func(nodeType string, nodeValue interface{}, node *wzlib.WzNode)
	OnWzFileLoaded func(wzStructure interface{})
	statusLabel    *widget.Label
	searchEntry    *widget.Entry
	expandAllBtn   *widget.Button
	collapseAllBtn *widget.Button

	// 文件管理功能
	fileManager *FileManager

	// 性能优化缓存
	nodeCache   map[string]*wzlib.WzNode
	childCache  map[string][]widget.TreeNodeID
	branchCache map[string]bool
}

// NewTreeViewer 创建新的树视图
func NewTreeViewer() *TreeViewer {
	tv := &TreeViewer{
		statusLabel: widget.NewLabel("No WZ file loaded"),
		nodeCache:   make(map[string]*wzlib.WzNode),
		childCache:  make(map[string][]widget.TreeNodeID),
		branchCache: make(map[string]bool),
	}

	tv.createContent()
	return tv
}

// createContent 创建树视图内容
func (tv *TreeViewer) createContent() {
	// 现代风格搜索框
	tv.searchEntry = widget.NewEntry()
	tv.searchEntry.SetPlaceHolder("🔍 搜索节点名称、类型或路径...")
	tv.searchEntry.OnChanged = tv.onSearchChanged

	// 创建现代风格的文件操作按钮
	loadBtn := widget.NewButtonWithIcon("📁 加载文件", theme.FolderOpenIcon(), tv.loadWzFile)
	loadBtn.Importance = widget.HighImportance

	removeBtn := widget.NewButtonWithIcon("🗑️ 移除", theme.DeleteIcon(), tv.removeWzFile)
	removeBtn.Importance = widget.MediumImportance

	clearBtn := widget.NewButtonWithIcon("🧹 清空", theme.ContentClearIcon(), tv.clearWzFiles)
	clearBtn.Importance = widget.LowImportance

	// 创建树操作按钮
	tv.expandAllBtn = widget.NewButtonWithIcon("📂 展开全部", theme.ViewFullScreenIcon(), tv.expandAll)
	tv.expandAllBtn.Importance = widget.MediumImportance

	tv.collapseAllBtn = widget.NewButtonWithIcon("📁 折叠全部", theme.ViewRestoreIcon(), tv.collapseAll)
	tv.collapseAllBtn.Importance = widget.MediumImportance

	// 美观的按钮布局：使用网格布局
	fileButtonContainer := container.NewGridWithColumns(3, loadBtn, removeBtn, clearBtn)
	treeButtonContainer := container.NewGridWithColumns(2, tv.expandAllBtn, tv.collapseAllBtn)
	buttonContainer := container.NewVBox(
		fileButtonContainer,
		treeButtonContainer,
	)

	// 创建树控件
	tv.tree = widget.NewTree(
		tv.childUIDs,
		tv.isBranch,
		tv.createNode,
		tv.updateNode,
	)

	tv.tree.OnSelected = tv.onNodeSelected

	// 设置树节点的最小高度
	tv.tree.Resize(fyne.NewSize(400, 600))

	// 创建美观的标题区域
	titleLabel := widget.NewLabel("🌳 WZ文件结构")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	// 创建带间距的按钮区域
	buttonSection := container.NewVBox(
		widget.NewCard("", "", buttonContainer),
	)

	// 组装内容 - 使用 Border 布局让树视图占用剩余全部高度
	topContent := container.NewVBox(
		container.NewPadded(titleLabel),
		widget.NewSeparator(),
		container.NewPadded(tv.searchEntry),
		buttonSection,
		widget.NewSeparator(),
		container.NewPadded(tv.statusLabel),
	)

	tv.content = container.NewBorder(
		topContent,                   // 顶部：标题、搜索框、按钮、状态标签
		nil,                          // 底部：无
		nil,                          // 左侧：无
		nil,                          // 右侧：无
		container.NewScroll(tv.tree), // 中心：树视图占用剩余全部空间
	)
}

// LoadWzStructure 加载WZ结构
func (tv *TreeViewer) LoadWzStructure(wzStructure interface{}) {
	if ws, ok := wzStructure.(*wzlib.WzStructure); ok {
		tv.wzStructure = ws

		// 清除缓存
		tv.clearCache()

		// 清除所有选择和展开状态
		tv.tree.UnselectAll()
		// 刷新树视图
		tv.tree.Refresh()

		// 更新状态标签
		if ws.WzNode != nil {
			tv.statusLabel.SetText(fmt.Sprintf("已加载: %s (%d个子节点)", ws.WzNode.Text, len(ws.WzNode.Nodes)))
		} else {
			tv.statusLabel.SetText("WZ文件结构已加载，但无数据")
		}
	} else {
		tv.statusLabel.SetText("加载WZ结构失败")
	}
}

// childUIDs 获取子节点ID列表
func (tv *TreeViewer) childUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	if tv.wzStructure == nil {
		return []widget.TreeNodeID{}
	}

	// 检查缓存
	if cached, exists := tv.childCache[string(uid)]; exists {
		return cached
	}

	var childIDs []widget.TreeNodeID

	if uid == "" {
		// 根节点 - 返回WZ文件的直接子节点
		if tv.wzStructure.WzNode != nil {
			for _, child := range tv.wzStructure.WzNode.Nodes {
				childPath := tv.getNodePath(child)
				childIDs = append(childIDs, childPath)
			}
		}
	} else {
		// 查找节点
		node := tv.findNodeByPath(uid)
		if node == nil {
			// 缓存空结果
			tv.childCache[string(uid)] = childIDs
			return childIDs
		}

		// 如果是图像节点，尝试提取
		if img, ok := node.Value.(*wzlib.WzImage); ok {
			img.TryExtract()
		}

		// 返回子节点ID
		for _, child := range node.Nodes {
			childPath := tv.getNodePath(child)
			childIDs = append(childIDs, childPath)
		}
	}

	// 缓存结果
	tv.childCache[string(uid)] = childIDs
	return childIDs
}

// clearCache 清除所有缓存
func (tv *TreeViewer) clearCache() {
	tv.nodeCache = make(map[string]*wzlib.WzNode)
	tv.childCache = make(map[string][]widget.TreeNodeID)
	tv.branchCache = make(map[string]bool)
}

// isBranch 判断是否为分支节点
func (tv *TreeViewer) isBranch(uid widget.TreeNodeID) bool {
	if tv.wzStructure == nil {
		return false
	}

	// 检查缓存
	if cached, exists := tv.branchCache[string(uid)]; exists {
		return cached
	}

	var result bool

	if uid == "" {
		// 根节点总是分支节点
		result = tv.wzStructure.WzNode != nil && len(tv.wzStructure.WzNode.Nodes) > 0
	} else {
		node := tv.findNodeByPath(uid)
		if node == nil {
			result = false
		} else {
			result = len(node.Nodes) > 0
		}
	}

	// 缓存结果
	tv.branchCache[string(uid)] = result
	return result
}

// createNode 创建节点
func (tv *TreeViewer) createNode(branch bool) fyne.CanvasObject {
	label := widget.NewLabel("Node")
	// 设置文字样式，增加可读性
	label.Wrapping = fyne.TextWrapOff
	label.Alignment = fyne.TextAlignLeading
	// 设置更紧凑的最小高度和宽度
	label.Resize(fyne.NewSize(200, 20))

	return label
}

// updateNode 更新节点显示
func (tv *TreeViewer) updateNode(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
	// 直接获取标签
	label := obj.(*widget.Label)

	if tv.wzStructure == nil {
		label.SetText("Not loaded")
		return
	}

	node := tv.findNodeByPath(uid)
	if node == nil {
		label.SetText("Node not found")
		return
	}

	// 根据节点类型设置显示文本和图标
	displayText := node.Text
	if node.Type != "" {
		displayText = fmt.Sprintf("%s [%s]", node.Text, node.Type)
	}

	label.SetText(displayText)
}

// onNodeSelected 节点选择事件
func (tv *TreeViewer) onNodeSelected(uid widget.TreeNodeID) {
	if tv.wzStructure == nil {
		return
	}

	node := tv.findNodeByPath(uid)
	if node == nil {
		return
	}

	// 如果是图像节点，尝试提取
	if img, ok := node.Value.(*wzlib.WzImage); ok {
		img.TryExtract()
	}

	// 通知选择事件
	if tv.OnNodeSelected != nil {
		tv.OnNodeSelected(node.Type, node.Value, node)
	}

	tv.statusLabel.SetText(fmt.Sprintf("已选择: %s", tv.getNodePath(node)))
}

// findNodeByPath 根据路径查找节点
func (tv *TreeViewer) findNodeByPath(path string) *wzlib.WzNode {
	if tv.wzStructure == nil || tv.wzStructure.WzNode == nil {
		return nil
	}

	// 检查缓存
	if cached, exists := tv.nodeCache[path]; exists {
		return cached
	}

	var result *wzlib.WzNode

	if path == tv.wzStructure.WzNode.Text {
		result = tv.wzStructure.WzNode
	} else {
		// 移除根节点前缀
		rootPrefix := tv.wzStructure.WzNode.Text + "/"
		if strings.HasPrefix(path, rootPrefix) {
			relativePath := strings.TrimPrefix(path, rootPrefix)
			result = tv.wzStructure.WzNode.GetNode(relativePath)
		}
	}

	// 缓存结果（包括nil结果）
	tv.nodeCache[path] = result
	return result
}

// getNodePath 获取节点路径
func (tv *TreeViewer) getNodePath(node *wzlib.WzNode) string {
	return node.GetFullPath()
}

// onSearchChanged 搜索变化事件
func (tv *TreeViewer) onSearchChanged(text string) {
	// TODO: 实现搜索功能
	if text == "" {
		tv.tree.Refresh()
		return
	}

	// 简单的搜索实现，可以后续优化
	tv.tree.Refresh()
}

// expandAll 展开全部节点
func (tv *TreeViewer) expandAll() {
	if tv.wzStructure == nil {
		return
	}

	// 递归展开所有节点
	tv.expandNodeRecursive("")
}

// collapseAll 折叠全部节点
func (tv *TreeViewer) collapseAll() {
	if tv.wzStructure == nil {
		return
	}

	// 递归折叠所有节点
	tv.collapseNodeRecursive("")
}

// expandNodeRecursive 递归展开节点（限制深度）
func (tv *TreeViewer) expandNodeRecursive(uid widget.TreeNodeID) {
	tv.expandNodeRecursiveWithDepth(uid, 0, 2) // 限制展开深度为2层
}

// expandNodeRecursiveWithDepth 递归展开节点（带深度限制）
func (tv *TreeViewer) expandNodeRecursiveWithDepth(uid widget.TreeNodeID, currentDepth, maxDepth int) {
	if currentDepth >= maxDepth {
		return
	}

	tv.tree.OpenBranch(uid)

	childIDs := tv.childUIDs(uid)
	for _, childID := range childIDs {
		if tv.isBranch(childID) {
			tv.expandNodeRecursiveWithDepth(childID, currentDepth+1, maxDepth)
		}
	}
}

// collapseNodeRecursive 递归折叠节点
func (tv *TreeViewer) collapseNodeRecursive(uid widget.TreeNodeID) {
	childIDs := tv.childUIDs(uid)
	for _, childID := range childIDs {
		if tv.isBranch(childID) {
			tv.collapseNodeRecursive(childID)
		}
	}

	tv.tree.CloseBranch(uid)
}

// loadWzFile 加载WZ文件
func (tv *TreeViewer) loadWzFile() {
	if tv.fileManager != nil {
		tv.fileManager.loadWzFile()
	} else {
		tv.statusLabel.SetText("文件管理器未初始化")
	}
}

// removeWzFile 移除选中的WZ文件
func (tv *TreeViewer) removeWzFile() {
	if tv.fileManager != nil {
		tv.fileManager.removeSelectedFile()
	} else {
		tv.statusLabel.SetText("文件管理器未初始化")
	}
}

// clearWzFiles 清空所有WZ文件
func (tv *TreeViewer) clearWzFiles() {
	if tv.fileManager != nil {
		tv.fileManager.clearFileList()
		// 清空树视图
		tv.wzStructure = nil
		tv.clearCache()
		tv.tree.Refresh()
		tv.statusLabel.SetText("已清空所有WZ文件")
	} else {
		tv.statusLabel.SetText("文件管理器未初始化")
	}
}

// GetContent 获取树视图内容
func (tv *TreeViewer) GetContent() fyne.CanvasObject {
	return tv.content
}
