package mounts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"

	"github.com/yhlooo/stackcrisp/pkg/layers"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

const (
	mountDataSubPathMountPath = "merged"
	mountDataSubPathWorkDir   = "work"

	loggerName = "mounts"
)

// MountOptions 挂载选项
type MountOptions struct {
	MountDataRoot string
	ChownUID      int
	ChownGID      int
}

// New 创建一个挂载
//
// layers 中第 0 到 n-2 个元素是 lower 层，其中 n-2 层是最顶层。第 n-1 层是 upper 层。
func New(ctx context.Context, id uid.UID, layers []layers.Layer, opts MountOptions) (Mount, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 最少需要两层，一层 lower 一层 upper
	if len(layers) < 2 {
		return nil, fmt.Errorf("length of layers is %d, too few, no less than 2", len(layers))
	}

	// 0 - n-2 是 lower
	lowerDir := make([]string, len(layers)-1)
	for i, l := range layers[:len(layers)-1] {
		lowerDir[len(layers)-i-2] = l.DiffDir()
	}
	// 最上层是 upper
	upperDir := layers[len(layers)-1].DiffDir()

	// 挂载参数
	mountPath := filepath.Join(opts.MountDataRoot, mountDataSubPathMountPath)
	workDir := filepath.Join(opts.MountDataRoot, mountDataSubPathWorkDir)
	ovlOpts := OverlayMountOptions{
		MountPath: mountPath,
		LowerDir:  lowerDir,
		UpperDir:  upperDir,
		WorkDir:   workDir,
		ReadOnly:  false,
	}

	logger.V(1).Info(fmt.Sprintf("overlay mount options: %#v", ovlOpts))
	return &defaultMount{
		mountedMount: mountedMount{
			id:        id,
			mountPath: mountPath,
			chownUID:  opts.ChownUID,
			chownGID:  opts.ChownGID,
		},
		ovlOpts: ovlOpts,
	}, nil
}

// NewMountedMount 创建一个已经挂载的挂载
func NewMountedMount(id uid.UID, opts MountOptions) Mount {
	mountPath := filepath.Join(opts.MountDataRoot, mountDataSubPathMountPath)
	return &mountedMount{
		id:        id,
		mountPath: mountPath,
		chownUID:  opts.ChownUID,
		chownGID:  opts.ChownGID,
	}
}

// Mount 挂载
type Mount interface {
	// ID 返回挂载 ID
	ID() uid.UID
	// MountPath 返回挂载点绝对路径
	MountPath() string
	// Mount 挂载
	Mount(ctx context.Context) error
	// Umount 卸载
	Umount(ctx context.Context) error
	// CreateSymlink 在指定路径创建访问挂载点的软链
	CreateSymlink(ctx context.Context, path string) error
}

// defaultMount 是 Mount 的一个默认实现
type defaultMount struct {
	mountedMount
	ovlOpts OverlayMountOptions
}

var _ Mount = &defaultMount{}

// mountedMount 是 Mount 的一个实现，但是已经挂载不能再挂载
type mountedMount struct {
	id        uid.UID
	mountPath string
	chownUID  int
	chownGID  int
}

var _ Mount = &mountedMount{}

// ID 返回挂载 ID
func (m *mountedMount) ID() uid.UID {
	return m.id
}

// Mount 挂载
func (m *mountedMount) Mount(context.Context) error {
	return fmt.Errorf("already mount")
}

// MountPath 返回挂载点绝对路径
func (m *mountedMount) MountPath() string {
	return m.mountPath
}

// CreateSymlink 在指定路径创建访问挂载点的软链
func (m *mountedMount) CreateSymlink(ctx context.Context, path string) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)
	logger.V(1).Info(fmt.Sprintf("ln -s %q %q", m.MountPath(), path))
	if err := os.Symlink(m.MountPath(), path); err != nil {
		return err
	}
	if err := os.Lchown(path, m.chownUID, m.chownGID); err != nil {
		return fmt.Errorf("chown %q to \"%d:%d\" error: %w", path, m.chownUID, m.chownGID, err)
	}
	return nil
}
