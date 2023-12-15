package trees

import (
	"fmt"
	"sync"

	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

// Tree 树
type Tree interface {
	// Get 通过节点 ID 获取节点
	Get(id uid.UID) (Node, bool)
	// GetByBranch 通过分支名获取节点
	GetByBranch(name string) (Node, bool)
	// GetByTag 通过标签名获取节点
	GetByTag(name string) (Node, bool)
	// Root 获取根节点
	Root() Node
	// Tags 获取标签与节点的映射关系的一个只读副本
	Tags() map[string]Node
	// Branches 获取分支与其头指针节点的映射关系的一个只读副本
	Branches() map[string]Node
	// Search 通过 key 搜索节点
	// key 可以是各种形式的节点 ID 、分支名、标签名
	Search(key string) (Node, bool)

	// AddNode 往树上添加节点
	AddNode(parentID uid.UID, node Node) error

	// TODO: DeleteNode(id uid.UID) bool

	// AddTag 添加标签
	// 可以覆盖同名标签
	AddTag(name string, nodeID uid.UID) error
	// DeleteTag 删除标签
	// 删除成功则返回 true 、不存在则返回 false
	DeleteTag(name string) bool
	// AddBranch 添加分支
	// 可以覆盖同名分支
	AddBranch(name string, nodeID uid.UID) error
	// UpdateBranch 更新分支头指针
	// force = true: 允许将分支头指针移动任何位置
	// force = false: 仅允许将分支头指针移动到当前位置的子节点
	UpdateBranch(name string, nodeID uid.UID, force bool) error
}

// NewTree 创建一个 Tree
func NewTree() Tree {
	return &defaultTree{}
}

// defaultTree 是 Tree 的一个默认实现
type defaultTree struct {
	root Node

	nodes     map[string]Node
	nodesLock sync.RWMutex

	branches     map[string]Node
	branchesLock sync.RWMutex

	tags     map[string]Node
	tagsLock sync.RWMutex
}

var _ Tree = &defaultTree{}

// Get 通过节点 ID 获取节点
func (tree *defaultTree) Get(id uid.UID) (Node, bool) {
	tree.nodesLock.RLock()
	defer tree.nodesLock.RUnlock()
	node, ok := tree.nodes[id.Hex()]
	return node, ok
}

// GetByBranch 通过分支名获取节点
func (tree *defaultTree) GetByBranch(name string) (Node, bool) {
	tree.branchesLock.RLock()
	defer tree.branchesLock.RUnlock()
	node, ok := tree.branches[name]
	return node, ok
}

// GetByTag 通过标签名获取节点
func (tree *defaultTree) GetByTag(name string) (Node, bool) {
	tree.tagsLock.RLock()
	defer tree.tagsLock.RUnlock()
	node, ok := tree.tags[name]
	return node, ok
}

// Root 获取根节点
func (tree *defaultTree) Root() Node {
	tree.nodesLock.RLock()
	defer tree.nodesLock.RUnlock()
	return tree.root
}

// Tags 获取标签与节点的映射关系的一个只读副本
func (tree *defaultTree) Tags() map[string]Node {
	tree.tagsLock.RLock()
	defer tree.tagsLock.RUnlock()

	if tree.tags == nil {
		return nil
	}

	// 拷贝一份作为输出
	ret := make(map[string]Node, len(tree.tags))
	for k, v := range tree.tags {
		ret[k] = v
	}
	return ret
}

// Branches 获取分支与其头指针节点的映射关系的一个只读副本
func (tree *defaultTree) Branches() map[string]Node {
	tree.branchesLock.RLock()
	defer tree.branchesLock.RUnlock()

	if tree.branches == nil {
		return nil
	}

	// 拷贝一份作为输出
	ret := make(map[string]Node, len(tree.branches))
	for k, v := range tree.branches {
		ret[k] = v
	}
	return ret
}

// Search 通过 key 搜索节点
// key 可以是各种形式的节点 ID 、分支名、标签名
func (tree *defaultTree) Search(key string) (Node, bool) {
	// 首先尝试从节点 ID 搜索
	var nodeID uid.UID
	if id, err := uid.DecodeUID128FromHex(key); err == nil {
		nodeID = id
	} else if id, err = uid.DecodeUID128FromBase32(key); err == nil {
		nodeID = id
	}
	if nodeID != nil {
		node, ok := tree.Get(nodeID)
		if ok {
			return node, true
		}
	}

	// 然后是标签
	if node, ok := tree.GetByTag(key); ok {
		return node, true
	}

	// 然后是分支
	if node, ok := tree.GetByBranch(key); ok {
		return node, true
	}

	return nil, false
}

// AddNode 往树上添加节点
func (tree *defaultTree) AddNode(parentID uid.UID, node Node) error {
	tree.nodesLock.Lock()
	defer tree.nodesLock.Unlock()

	// 如果没有父节点就是插入根节点
	if parentID == nil {
		if tree.root != nil {
			// 根节点不允许变更
			return fmt.Errorf("root node already exists (%s) and can not be change", tree.root.ID().Hex())
		}
		tree.root = node
		tree.updateIndexes(node, false)
		return nil
	}

	// 找到父节点
	parent, ok := tree.nodes[parentID.Hex()]
	if !ok {
		return fmt.Errorf("parent node %q not found", parentID.Hex())
	}

	// 设置父子关系
	parent.AddChild(node)
	node.SetParent(parent)
	// 更新索引
	tree.updateIndexes(node, false)

	return nil
}

// AddTag 添加标签
// 可以覆盖同名标签
func (tree *defaultTree) AddTag(name string, nodeID uid.UID) error {
	tree.tagsLock.Lock()
	defer tree.tagsLock.Unlock()
	tree.nodesLock.RLock()
	defer tree.nodesLock.RUnlock()

	// 找到节点
	node, ok := tree.nodes[nodeID.Hex()]
	if !ok {
		return fmt.Errorf("node %q not found", nodeID.Hex())
	}

	// 添加标签
	if tree.tags == nil {
		tree.tags = make(map[string]Node)
	}
	tree.tags[name] = node

	return nil
}

// DeleteTag 删除标签
// 删除成功则返回 true 、不存在则返回 false
func (tree *defaultTree) DeleteTag(name string) bool {
	tree.tagsLock.Lock()
	defer tree.tagsLock.Unlock()

	if _, ok := tree.tags[name]; !ok {
		return false
	}
	delete(tree.tags, name)

	return true
}

// AddBranch 添加分支
// 可以覆盖同名分支
func (tree *defaultTree) AddBranch(name string, nodeID uid.UID) error {
	tree.branchesLock.Lock()
	defer tree.branchesLock.Unlock()
	tree.nodesLock.RLock()
	defer tree.nodesLock.RUnlock()

	// 找到节点
	node, ok := tree.nodes[nodeID.Hex()]
	if !ok {
		return fmt.Errorf("node %q not found", nodeID.Hex())
	}

	// 添加分支头指针索引
	if tree.branches == nil {
		tree.branches = make(map[string]Node)
	}
	tree.branches[name] = node

	return nil

}

// UpdateBranch 更新分支头指针
// force = true: 允许将分支头指针移动任何位置
// force = false: 仅允许将分支头指针移动到当前位置的子节点
func (tree *defaultTree) UpdateBranch(name string, nodeID uid.UID, force bool) error {
	tree.branchesLock.Lock()
	defer tree.branchesLock.Unlock()
	tree.nodesLock.RLock()
	defer tree.nodesLock.RUnlock()

	// 找到节点
	node, ok := tree.nodes[nodeID.Hex()]
	if !ok {
		return fmt.Errorf("node %q not found", nodeID.Hex())
	}

	// 找到分支头指针
	head, ok := tree.branches[name]
	if !ok {
		return fmt.Errorf("branch %q not found", name)
	}

	// 原位，不用更新
	if head.ID().Hex() == node.ID().Hex() {
		return nil
	}

	// 强制更新
	if force {
		tree.branches[name] = node
		return nil
	}

	// 检查新位置到当前 head 的连通性
	cur := node
	connected := false
	for cur != nil {
		if cur.ID().Hex() == head.ID().Hex() {
			connected = true
			break
		}
	}
	if !connected {
		return fmt.Errorf(
			"can not move HEAD of the branch %q to %q, unreachable from the current HEAD %q",
			name, node.ID().Hex(), head.ID().Hex(),
		)
	}

	tree.branches[name] = node
	return nil
}

// updateIndexes 更新索引
// 将 node 及其子节点添加到索引中
func (tree *defaultTree) updateIndexes(node Node, lock bool) {
	if lock {
		tree.nodesLock.Lock()
		defer tree.nodesLock.Unlock()
	}

	if tree.nodes == nil {
		tree.nodes = make(map[string]Node)
	}

	roots := []Node{node}
	for len(roots) > 0 {
		// 取队列第一个节点
		item := roots[0]
		roots = roots[1:]
		// 添加到索引
		tree.nodes[item.ID().Hex()] = item
		// 将子节点添加到递归列表
		for _, child := range item.Children() {
			roots = append(roots, child)
		}
	}
}
