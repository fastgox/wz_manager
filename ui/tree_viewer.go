package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/luoxk/wzlib"
)

// TreeViewer WZ文件树视图
type TreeViewer struct {
	content        *fyne.Container
	tree           *widget.Tree
	wzStructure    *wzlib.WzStructure
	OnNodeSelected func(nodeType string, nodeValue interface{})
	statusLabel    *widget.Label
	searchEntry    *widget.Entry
	expandAllBtn   *widget.Button
	collapseAllBtn *widget.Button
}

// NewTreeViewer 创建新的树视图
func NewTreeViewer() *TreeViewer {
	tv := &TreeViewer{
		statusLabel: widget.NewLabel("No WZ file loaded"),
	}

	tv.createContent()
	return tv
}

// createContent 创建树视图内容
func (tv *TreeViewer) createContent() {
	// 搜索框
	tv.searchEntry = widget.NewEntry()
	tv.searchEntry.SetPlaceHolder("搜索节点...")
	tv.searchEntry.OnChanged = tv.onSearchChanged

	// 控制按钮
	tv.expandAllBtn = widget.NewButton("展开全部", tv.expandAll)
	tv.collapseAllBtn = widget.NewButton("折叠全部", tv.collapseAll)

	buttonContainer := container.NewHBox(tv.expandAllBtn, tv.collapseAllBtn)

	// 创建树控件
	tv.tree = widget.NewTree(
		tv.childUIDs,
		tv.isBranch,
		tv.createNode,
		tv.updateNode,
	)

	tv.tree.OnSelected = tv.onNodeSelected

	// 组装内容
	tv.content = container.NewVBox(
		widget.NewLabel("WZ文件结构"),
		widget.NewSeparator(),
		tv.searchEntry,
		buttonContainer,
		widget.NewSeparator(),
		tv.statusLabel,
		tv.tree,
	)
}

// LoadWzStructure 加载WZ结构
func (tv *TreeViewer) LoadWzStructure(wzStructure interface{}) {
	if ws, ok := wzStructure.(*wzlib.WzStructure); ok {
		tv.wzStructure = ws
		tv.tree.Refresh()
		tv.statusLabel.SetText("WZ file structure loaded")
	}
}

// childUIDs 获取子节点ID列表
func (tv *TreeViewer) childUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	if tv.wzStructure == nil {
		return []widget.TreeNodeID{}
	}

	if uid == "" {
		// 根节点 - 返回WZ文件的直接子节点
		if tv.wzStructure.WzNode != nil {
			var childIDs []widget.TreeNodeID
			for _, child := range tv.wzStructure.WzNode.Nodes {
				childPath := tv.getNodePath(child)
				childIDs = append(childIDs, childPath)
			}
			return childIDs
		}
		return []widget.TreeNodeID{}
	}

	// 查找节点
	node := tv.findNodeByPath(uid)
	if node == nil {
		return []widget.TreeNodeID{}
	}

	// 如果是图像节点，尝试提取
	if img, ok := node.Value.(*wzlib.WzImage); ok {
		img.TryExtract()
	}

	// 返回子节点ID
	var childIDs []widget.TreeNodeID
	for _, child := range node.Nodes {
		childPath := tv.getNodePath(child)
		childIDs = append(childIDs, childPath)
	}

	return childIDs
}

// isBranch 判断是否为分支节点
func (tv *TreeViewer) isBranch(uid widget.TreeNodeID) bool {
	if tv.wzStructure == nil {
		return false
	}

	if uid == "" {
		// 根节点总是分支节点
		return tv.wzStructure.WzNode != nil && len(tv.wzStructure.WzNode.Nodes) > 0
	}

	node := tv.findNodeByPath(uid)
	if node == nil {
		return false
	}

	return len(node.Nodes) > 0 || node.Value != nil
}

// createNode 创建节点
func (tv *TreeViewer) createNode(branch bool) fyne.CanvasObject {
	return widget.NewLabel("Node")
}

// updateNode 更新节点显示
func (tv *TreeViewer) updateNode(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
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
		tv.OnNodeSelected(node.Type, node.Value)
	}

	tv.statusLabel.SetText(fmt.Sprintf("已选择: %s", tv.getNodePath(node)))
}

// findNodeByPath 根据路径查找节点
func (tv *TreeViewer) findNodeByPath(path string) *wzlib.WzNode {
	if tv.wzStructure == nil || tv.wzStructure.WzNode == nil {
		return nil
	}

	if path == tv.wzStructure.WzNode.Text {
		return tv.wzStructure.WzNode
	}

	// 移除根节点前缀
	rootPrefix := tv.wzStructure.WzNode.Text + "/"
	if strings.HasPrefix(path, rootPrefix) {
		relativePath := strings.TrimPrefix(path, rootPrefix)
		return tv.wzStructure.WzNode.GetNode(relativePath)
	}

	return nil
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

// expandNodeRecursive 递归展开节点
func (tv *TreeViewer) expandNodeRecursive(uid widget.TreeNodeID) {
	tv.tree.OpenBranch(uid)

	childIDs := tv.childUIDs(uid)
	for _, childID := range childIDs {
		if tv.isBranch(childID) {
			tv.expandNodeRecursive(childID)
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

// GetContent 获取树视图内容
func (tv *TreeViewer) GetContent() fyne.CanvasObject {
	return tv.content
}
