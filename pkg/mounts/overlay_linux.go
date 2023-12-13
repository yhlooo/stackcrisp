//go:build linux

package mounts

import (
	"fmt"
	"strings"
	"syscall"
)

// CreateOverlayMount 创建一个 Overlay 挂载
func CreateOverlayMount(opts OverlayMountOptions) error {
	// 挂载标记
	var flags uintptr
	if opts.ReadOnly {
		flags |= syscall.MS_RDONLY
	}
	// 挂载点
	mountPath := opts.MountPath
	if mountPath == "" {
		mountPath = "overlay"
	}
	// 挂载参数
	data := fmt.Sprintf(
		"lowerdir=%s,upperdir=%s,workdir=%s",
		strings.Join(opts.LowerDir, ":"),
		opts.UpperDir,
		opts.WorkDir,
	)

	return syscall.Mount(opts.Source, mountPath, "overlay", flags, data)
}
