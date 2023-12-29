package spaces

import (
	"context"

	"github.com/yhlooo/stackcrisp/pkg/mounts"
)

// SpaceOptions 空间选项
type SpaceOptions struct {
	// 空间数据存储根目录
	SpaceDataRoot string
}

// Space 存储空间
type Space interface {
	// Init 初始化
	Init(ctx context.Context) error
	// Load 加载数据
	Load(ctx context.Context) error
	// Save 将数据持久化
	Save(ctx context.Context) error
	// CreateMount 创建一个该空间的挂载
	CreateMount(ctx context.Context, revision string, mountOpts mounts.MountOptions) (mounts.Mount, error)
}
