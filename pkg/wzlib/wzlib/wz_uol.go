package wzlib

import "strings"

// WzUol 表示WZ中的UOL（路径引用）对象
type WzUol struct {
	Uol string
}

// NewWzUol 创建一个新的WzUol对象
func NewWzUol(uol string) *WzUol {
	return &WzUol{Uol: uol}
}

// HandleUol 根据UOL路径解析并返回目标节点
func (u *WzUol) HandleUol(currentNode *WzNode) *WzNode {
	if currentNode == nil || currentNode.ParentNode == nil || u.Uol == "" {
		return nil
	}
	dirs := strings.Split(u.Uol, "/")
	current := currentNode.ParentNode
	outImg := false

	for _, dir := range dirs {
		if dir == ".." {
			if current.ParentNode == nil {
				// 尝试获取WzImage
				if img, ok := current.Value.(*WzImage); ok && img.Node != nil && img.Node.ParentNode != nil {
					current = img.Node.ParentNode
					outImg = true
				} else {
					current = nil
				}
			} else {
				current = current.ParentNode
			}
		} else {
			dirNode := current.FindChild(dir)
			// 如果没找到且刚刚跳出img，尝试 dir+".img"
			if dirNode == nil && outImg {
				dirNode = current.FindChild(dir + ".img")
				if dirNode != nil {
					if _, ok := dirNode.Value.(*WzImage); ok {
						outImg = false
					}
				}
			}
			current = dirNode
		}
		if current == nil {
			return nil
		}
	}
	return current
}
