package wzlib

import "strings"

type WzNode struct {
	Value      any       // 节点的值
	Text       string    // 节点的名称
	ParentNode *WzNode   // 父节点
	Nodes      []*WzNode // 子节点集合
	Type       string    // 节点类型
}

// NewWzNode 创建一个新的 WzNode
func NewWzNode(text string) *WzNode {
	return &WzNode{
		Text:  text,
		Nodes: []*WzNode{},
	}
}

// GetFullPath 返回节点的完整路径
func (n *WzNode) GetFullPath() string {
	if n.ParentNode == nil {
		return n.Text
	}
	return n.ParentNode.GetFullPath() + "/" + n.Text
}

// AddChild 添加子节点
func (n *WzNode) AddChild(child *WzNode) *WzNode {
	child.ParentNode = n
	n.Nodes = append(n.Nodes, child)
	return child
}

// FindChild 根据名称查找子节点
func (n *WzNode) FindChild(name string) *WzNode {
	for _, child := range n.Nodes {
		if child.Text == name {
			return child
		}
	}
	return nil
}

// Clone 克隆当前节点
func (n *WzNode) Clone() *WzNode {
	newNode := NewWzNode(n.Text)
	newNode.Value = n.Value
	newNode.Type = n.Type
	for _, child := range n.Nodes {
		newNode.AddChild(child.Clone())
	}
	return newNode
}

// CompareTo 比较两个节点的名称
func (n *WzNode) CompareTo(other *WzNode) int {
	if n.Text < other.Text {
		return -1
	}
	if n.Text > other.Text {
		return 1
	}
	return 0
}

// ToString 返回节点的文本名称
func (n *WzNode) ToString() string {
	return n.Text
}

// GetNode 根据路径查找子节点，路径用'/'分隔
func (n *WzNode) GetNode(path string) *WzNode {
	if path == "" {
		return n
	}
	parts := strings.SplitN(path, "/", 2)
	child := n.FindChild(parts[0])
	if child == nil {
		return nil
	}
	// 如果child是*WzImage类型，尝试解压img文件
	if img, ok := child.Value.(*WzImage); ok {
		img.TryExtract() // 假设有TryExtract方法
	}
	if len(parts) == 1 {
		return child
	}
	return child.GetNode(parts[1])
}
