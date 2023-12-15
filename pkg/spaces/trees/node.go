package trees

import (
	"sync"

	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

// Node 树上的节点
type Node interface {
	// ID 返回节点 ID
	ID() uid.UID
	// Parent 返回父节点
	Parent() Node
	// Children 返回子节点 ID 与对应节点的映射关系的一个只读副本
	// 键为节点 ID 的十六进制表示
	Children() map[string]Node
	// HasChild 返回是否有 ID 为 childID 的子节点
	HasChild(childID uid.UID) bool
	// IsRoot 返回当前是否根节点
	IsRoot() bool
	// IsLeaf 返回当前节点是否叶子节点
	IsLeaf() bool

	// SetParent 设置父节点
	SetParent(parent Node)
	// AddChild 添加子节点
	// 添加成功则返回 true 、已经存在则返回 false
	AddChild(child Node) bool
	// DeleteChild 删除子节点
	// 删除成功则返回 true 、不存在则返回 false
	DeleteChild(childID uid.UID) bool
}

// NewNode 创建一个 Node
func NewNode(id uid.UID) Node {
	return &defaultNode{id: id}
}

// defaultNode 是 Node 的一个默认实现
type defaultNode struct {
	id           uid.UID
	parent       Node
	children     map[string]Node
	childrenLock sync.RWMutex
}

var _ Node = &defaultNode{}

// ID 返回节点 ID
func (node *defaultNode) ID() uid.UID {
	return node.id
}

// Parent 返回父节点
func (node *defaultNode) Parent() Node {
	return node.parent
}

// Children 返回子节点 ID 与对应节点的映射关系的一个只读副本
func (node *defaultNode) Children() map[string]Node {
	node.childrenLock.RLock()
	defer node.childrenLock.RUnlock()

	if node.children == nil {
		return nil
	}

	// 拷贝一份作为输出
	ret := make(map[string]Node, len(node.children))
	for k, v := range node.children {
		ret[k] = v
	}

	return ret
}

// HasChild 返回是否有 ID 为 childID 的子节点
func (node *defaultNode) HasChild(childID uid.UID) bool {
	node.childrenLock.RLock()
	defer node.childrenLock.RUnlock()

	_, ok := node.children[childID.Hex()]
	return ok
}

// IsRoot 返回当前是否根节点
func (node *defaultNode) IsRoot() bool {
	// 没有父节点就是根节点
	return node.Parent() == nil
}

// IsLeaf 返回当前节点是否叶子节点
func (node *defaultNode) IsLeaf() bool {
	node.childrenLock.RLock()
	defer node.childrenLock.RUnlock()
	// 没有子节点就是叶子节点
	return len(node.children) == 0
}

// SetParent 设置父节点
func (node *defaultNode) SetParent(parent Node) {
	node.parent = parent
}

// AddChild 添加子节点
// 添加成功则返回 true 、已经存在则返回 false
func (node *defaultNode) AddChild(child Node) bool {
	node.childrenLock.Lock()
	defer node.childrenLock.Unlock()

	// 检查是否已经存在
	_, ok := node.children[child.ID().Hex()]
	if ok {
		return false
	}

	// 添加子节点
	if node.children == nil {
		node.children = make(map[string]Node)
	}
	node.children[child.ID().Hex()] = child

	return true
}

// DeleteChild 删除子节点
// 删除成功则返回 true 、不存在则返回 false
func (node *defaultNode) DeleteChild(childID uid.UID) bool {
	node.childrenLock.Lock()
	defer node.childrenLock.Unlock()

	// 检查是否存在
	_, ok := node.children[childID.Hex()]
	if !ok {
		return false
	}

	// 删除子节点
	delete(node.children, childID.Hex())

	return true
}
