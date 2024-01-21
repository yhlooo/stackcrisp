package workspaces

import (
	"context"
	"time"

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
	Branch() BranchInfo
	// Space 返回工作空间对应空间
	Space() spaces.Space
	// Mount 返回工作空间对应挂载
	Mount() mounts.Mount

	// Expand 展开工作空间
	Expand(ctx context.Context) error

	// GetHistory 获取提交历史
	GetHistory(ref string) ([]Commit, error)

	// AllBranches 返回本地分支和全局分支列表
	AllBranches() []Branch
	// LocalBranches 返回本地分支列表
	LocalBranches() []Branch
	// RemoteBranches 返回远程分支列表
	RemoteBranches() []Branch
	// SetBranch 设置当前分支
	SetBranch(localName string) error
	// AddBranch 添加分支
	AddBranch(ctx context.Context, branchLocalName string, ref string, force bool) error

	// Tags 返回标签列表
	Tags() []string
	// DeleteTag 删除标签
	DeleteTag(ctx context.Context, tagName string) error
	// AddTag 添加标签
	AddTag(ctx context.Context, tagName string, ref string, force bool) error

	// Search 通过 ref 搜索节点
	//
	// ref 可以是各种形式的节点 ID 、分支名、标签名
	Search(ref string) (trees.Node, trees.KeyType, bool)
}

// BranchInfo 分支信息
type BranchInfo interface {
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

	// Branch 返回带有指定头指针的分支
	Branch(head trees.Node) Branch
}

// Branch 分支
type Branch interface {
	BranchInfo
	// Head 返回分支头指针
	Head() trees.Node
}

// CommitInfo 提交信息
type CommitInfo interface {
	// Date 返回提交日期时间
	Date() *time.Time
	// Message 返回提交信息
	Message() string
	// SetToNode 设置提交信息到节点
	SetToNode(node trees.Node)
}

// Commit 提交
type Commit interface {
	CommitInfo

	// ID 返回提交 ID
	ID() uid.UID
	// Branches 返回提交对应分支头指针的分支
	Branches() []Branch
	// Tags 返回提交对应标签
	Tags() []string
}
