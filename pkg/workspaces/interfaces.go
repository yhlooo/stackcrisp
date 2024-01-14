package workspaces

import (
	"context"

	"github.com/yhlooo/stackcrisp/pkg/mounts"
	"github.com/yhlooo/stackcrisp/pkg/spaces"
	"github.com/yhlooo/stackcrisp/pkg/spaces/trees"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

// Workspace 工作空间
type Workspace interface {
	// ID 返回 Workspace ID
	ID() uid.UID

	// Path 返回工作空间路径
	Path() string
	// Head 返回头指针
	Head() trees.Node
	// Branch 返回当前分支
	Branch() Branch
	// Space 返回工作空间对应空间
	Space() spaces.Space
	// Mount 返回工作空间对应挂载
	Mount() mounts.Mount

	// Expand 展开工作空间
	Expand(ctx context.Context) error
	// AllBranches 返回本地分支和全局分支列表
	AllBranches() []Branch
	// LocalBranches 返回本地分支列表
	LocalBranches() []Branch
	// RemoteBranches 返回远程分支列表
	RemoteBranches() []Branch
	// SetBranch 设置分支
	SetBranch(localName string) error
	// Search 通过 key 搜索节点
	//
	// key 可以是各种形式的节点 ID 、分支名、标签名
	Search(key string) (trees.Node, trees.KeyType, bool)
}

// Branch 分支
type Branch interface {
	// Name 返回分支名
	Name() string
	// FullName 返回在 spaces.Space 中存储的完整名
	FullName() string
	// LocalName 返回在 Workspace 中的本地名
	LocalName() string

	// IsGlobal 返回该分支是否全局分支
	IsGlobal() bool
	// IsLocal 返回该分支是否 Workspace 本地分支
	IsLocal() bool
	// WorkspaceID 对于 Workspace 本地分支，返回所属 Workspace ID ，否则返回 nil
	WorkspaceID() uid.UID
}
