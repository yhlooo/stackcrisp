package mounts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"

	"github.com/yhlooo/stackcrisp/pkg/layers"
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
func New(ctx context.Context, layers []layers.Layer, opts MountOptions) (Mount, error) {
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
		ovlOpts:  ovlOpts,
		chownUID: opts.ChownUID,
		chownGID: opts.ChownGID,
	}, nil
}

// Mount 挂载
type Mount interface {
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
	ovlOpts  OverlayMountOptions
	chownUID int
	chownGID int
}

var _ Mount = &defaultMount{}

// MountPath 返回挂载点绝对路径
func (m *defaultMount) MountPath() string {
	return m.ovlOpts.MountPath
}

// CreateSymlink 在指定路径创建访问挂载点的软链
func (m *defaultMount) CreateSymlink(ctx context.Context, path string) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)
	logger.V(1).Info(fmt.Sprintf("ln -s %q %q", m.ovlOpts.MountPath, path))
	if err := os.Symlink(m.ovlOpts.MountPath, path); err != nil {
		return err
	}
	if err := os.Lchown(path, m.chownUID, m.chownGID); err != nil {
		return fmt.Errorf("chown %q to \"%d:%d\" error: %w", path, m.chownUID, m.chownGID, err)
	}
	return nil
}
