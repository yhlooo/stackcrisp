package manager

import (
	"context"

	"github.com/yhlooo/stackcrisp/pkg/mounts"
	"github.com/yhlooo/stackcrisp/pkg/spaces"
)

// Manager 管理器
type Manager interface {
	// Prepare 准备
	Prepare(ctx context.Context) error
	// CreateSpace 创建一个存储空间
	CreateSpace(ctx context.Context) (spaces.Space, error)
	// CreateMount 使用指定空间版本创建一个挂载
	CreateMount(ctx context.Context, space spaces.Space, revision string) (mounts.Mount, error)
}

// Options 管理器选项
type Options struct {
	// 数据存储根目录
	DataRoot string
}
