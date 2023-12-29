//go:build linux

package mounts

import (
	"context"
	"fmt"
	"strings"
	"syscall"

	"github.com/go-logr/logr"
)

// CreateOverlayMount 创建一个 Overlay 挂载
func CreateOverlayMount(ctx context.Context, opts OverlayMountOptions) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 挂载标记
	var flags uintptr
	var showOpts string
	if opts.ReadOnly {
		flags |= syscall.MS_RDONLY
		showOpts = "ro,"
	}
	// 挂载名
	source := opts.Source
	if source == "" {
		source = "overlay"
	}
	// 挂载参数
	data := fmt.Sprintf(
		"lowerdir=%s,upperdir=%s,workdir=%s",
		strings.Join(opts.LowerDir, ":"),
		opts.UpperDir,
		opts.WorkDir,
	)

	showOpts += data
	logger.V(1).Info(fmt.Sprintf("mount -t overlay %q -o %q %q", source, showOpts, opts.MountPath))
	return syscall.Mount(source, opts.MountPath, "overlay", flags, data)
}
