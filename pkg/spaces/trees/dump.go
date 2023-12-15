package trees

import (
	"fmt"

	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

// TreeDump 导出的树，可 JSON 序列化
type TreeDump struct {
	// 节点
	Nodes *NodeDump `json:"nodes"`
	// 分支
	Branches map[string]string `json:"branches"`
	// 标签
	Tags map[string]string `json:"tags"`
}

// NodeDump 导出的节点，可 JSON 序列化
type NodeDump struct {
	// 节点 ID
	ID string `json:"id"`
	// 子节点
	Children []NodeDump `json:"children,omitempty"`
}

// Dump 导出树
func Dump(tree Tree) TreeDump {
	dump := TreeDump{
		Nodes:    nil,
		Branches: nil,
		Tags:     nil,
	}
	root := tree.Root()
	branches := tree.Branches()
	tags := tree.Tags()

	// 导出节点
	if root != nil {
		dump.Nodes = &NodeDump{
			ID: root.ID().Hex(),
		}
		nodes := map[*NodeDump]map[string]Node{
			nil: {root.ID().Hex(): root},
		}
		for len(nodes) > 0 {
			newNodes := map[*NodeDump]map[string]Node{}
			for parent, children := range nodes {
				if parent == nil {
					parent = dump.Nodes
				}
				if len(children) == 0 {
					continue
				}
				i := 0
				parent.Children = make([]NodeDump, len(children))
				for _, n := range children {
					parent.Children[i].ID = n.ID().Hex()
					newNodes[&parent.Children[i]] = n.Children()
					i++
				}
			}
			nodes = newNodes
		}
	}
	// 导出分支
	if branches != nil {
		dump.Branches = make(map[string]string)
		for k, v := range branches {
			dump.Branches[k] = v.ID().Hex()
		}
	}
	// 导出标签
	if tags != nil {
		dump.Tags = make(map[string]string)
		for k, v := range tags {
			dump.Tags[k] = v.ID().Hex()
		}
	}

	return dump
}

// Load 基于 dump 创建一个 Tree
func Load(dump TreeDump) (Tree, error) {
	tree := NewTree()
	if dump.Nodes == nil {
		return tree, nil
	}

	// 导入节点
	rootID, err := uid.DecodeUID128FromHex(dump.Nodes.ID)
	if err != nil {
		return tree, fmt.Errorf("decode id %q of the root node error: %w", dump.Nodes.ID, err)
	}
	if err := tree.AddNode(nil, NewNode(rootID)); err != nil {
		return tree, fmt.Errorf("add root node %q error: %w", rootID, err)
	}
	nodes := []NodeDump{*dump.Nodes}
	for len(nodes) > 0 {
		item := nodes[0]
		nodes = nodes[1:]
		parentID, err := uid.DecodeUID128FromHex(item.ID)
		if err != nil {
			return tree, fmt.Errorf("decode id %q of the node error: %w", item.ID, err)
		}
		for _, child := range item.Children {
			childID, err := uid.DecodeUID128FromHex(child.ID)
			if err != nil {
				return tree, fmt.Errorf("decode id %q of the node error: %w", child.ID, err)
			}
			if err := tree.AddNode(parentID, NewNode(childID)); err != nil {
				return nil, fmt.Errorf("add node %q as child of the node %q error: %w", childID, parentID, err)
			}
		}
		nodes = append(nodes, item.Children...)
	}

	// 导入分支
	for name, v := range dump.Branches {
		nodeID, err := uid.DecodeUID128FromHex(v)
		if err != nil {
			return tree, fmt.Errorf("decode id %q of the node error: %w", nodeID, err)
		}
		if err := tree.AddBranch(name, nodeID); err != nil {
			return tree, fmt.Errorf("add branch %q -> %q error: %w", name, nodeID)
		}
	}

	// 导入标签
	for name, v := range dump.Tags {
		nodeID, err := uid.DecodeUID128FromHex(v)
		if err != nil {
			return tree, fmt.Errorf("decode id %q of the node error: %w", nodeID, err)
		}
		if err := tree.AddTag(name, nodeID); err != nil {
			return tree, fmt.Errorf("add tag %q -> %q error: %w", name, nodeID)
		}
	}

	return tree, nil
}
