package workspaces

import (
	"context"

	"github.com/yhlooo/stackcrisp/pkg/mounts"
	"github.com/yhlooo/stackcrisp/pkg/spaces"
	"github.com/yhlooo/stackcrisp/pkg/spaces/trees"
)

// Workspace 工作空间
type Workspace interface {
	// Path 返回工作空间路径
	Path() string
	// Head 返回头指针
	Head() trees.Node
	// Branch 返回当前分支
	Branch() string
	// Space 返回工作空间对应空间
	Space() spaces.Space
	// Mount 返回工作空间对应挂载
	Mount() mounts.Mount

	// Expand 展开工作空间
	Expand(ctx context.Context) error
	// LocalBranches 返回本地分支列表
	LocalBranches() []string
	// RemoteBranches 返回远程分支列表
	RemoteBranches() []string
}
