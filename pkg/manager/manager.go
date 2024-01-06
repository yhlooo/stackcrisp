package manager

import (
	"context"

	"github.com/yhlooo/stackcrisp/pkg/workspaces"
)

// Manager 管理器
type Manager interface {
	// Prepare 准备
	Prepare(ctx context.Context) error
	// CreateWorkspace 创建工作空间
	CreateWorkspace(ctx context.Context, path string) (workspaces.Workspace, error)
	// GetWorkspaceFromPath 从指定目录获取对应工作空间
	GetWorkspaceFromPath(ctx context.Context, path string) (workspaces.Workspace, error)
	// RemoveWorkspaceMount 删除工作空间挂载
	RemoveWorkspaceMount(ctx context.Context, ws workspaces.Workspace) error
	// Clone 克隆工作空间
	Clone(ctx context.Context, ws workspaces.Workspace, targetPath string) (workspaces.Workspace, error)
	// Commit 提交工作空间变更
	Commit(ctx context.Context, ws workspaces.Workspace) (workspaces.Workspace, error)
}

// Options 管理器选项
type Options struct {
	// 数据存储根目录
	DataRoot string
	// 修改空间中存储文件所属用户 ID ， -1 表示不修改
	ChownUID int
	// 修改空间中存储文件所属用户组 ID ， -1 表示不修改
	ChownGID int
}
