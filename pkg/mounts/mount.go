package mounts

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/yhlooo/stackcrisp/pkg/layers"
)

const (
	mountDataSubPathMountPath = "merged"
	mountDataSubPathWorkDir   = "work"
)

// MountOptions 挂载选项
type MountOptions struct {
	MountDataRoot string
}

// New 创建一个挂载
func New(ctx context.Context, layers []layers.Layer, opts MountOptions) (Mount, error) {
	// 最少需要两层，一层 lower 一层 upper
	if len(layers) < 2 {
		return nil, fmt.Errorf("length of layers is %d, too few, no less than 2", len(layers))
	}

	// 0 - n-1 是 lower
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

	return &defaultMount{
		ovlOpts: ovlOpts,
	}, nil
}

// Mount 挂载
type Mount interface {
	// MountPath 返回挂载点绝对路径
	MountPath() string
	// Mount 挂载
	Mount() error
	// Umount 卸载
	Umount() error
}

// defaultMount 是 Mount 的一个默认实现
type defaultMount struct {
	ovlOpts OverlayMountOptions
}

var _ Mount = &defaultMount{}

// MountPath 返回挂载点绝对路径
func (m *defaultMount) MountPath() string {
	return m.ovlOpts.MountPath
}
