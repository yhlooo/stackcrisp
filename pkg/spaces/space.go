package spaces

import (
	"context"

	"github.com/yhlooo/stackcrisp/pkg/mounts"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

// SpaceOptions 空间选项
type SpaceOptions struct {
	// 空间数据存储根目录
	SpaceDataRoot string
}

// Space 存储空间
//
//nolint:lll
type Space interface {
	// ID 返回空间 ID
	ID() uid.UID
	// Init 初始化
	Init(ctx context.Context) error
	// Load 加载数据
	Load(ctx context.Context) error
	// Save 将数据持久化
	Save(ctx context.Context) error
	// CreateMount 创建一个该空间的挂载
	CreateMount(ctx context.Context, revision string, mountID uid.UID, mountOpts mounts.MountOptions) (mount mounts.Mount, head uid.UID, err error)
}
