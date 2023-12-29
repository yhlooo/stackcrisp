//go:build linux

package mounts

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/go-logr/logr"

	fsutil "github.com/yhlooo/stackcrisp/pkg/utils/fs"
)

// Mount 挂载
func (m *defaultMount) Mount(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	mountPath := m.ovlOpts.MountPath
	workDir := m.ovlOpts.WorkDir

	// 确保相关目录存在
	if !fsutil.IsDir(mountPath) {
		logger.V(1).Info(fmt.Sprintf("mkdir %q", mountPath))
		if err := os.Mkdir(mountPath, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("mkdir %q error: %w", mountPath, err)
		}
	}
	if !fsutil.IsDir(workDir) {
		logger.V(1).Info(fmt.Sprintf("mkdir %q", workDir))
		if err := os.Mkdir(workDir, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("mkdir %q error: %w", workDir, err)
		}
	}

	if err := CreateOverlayMount(ctx, m.ovlOpts); err != nil {
		return err
	}
	if err := os.Chown(mountPath, m.chownUID, m.chownGID); err != nil {
		return fmt.Errorf("chown %q to \"%d:%d\" error: %w", mountPath, m.chownUID, m.chownGID, err)
	}
	return nil
}

// Umount 卸载挂载
func (m *mountedMount) Umount(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)
	logger.V(1).Info(fmt.Sprintf("umount %q", m.MountPath()))
	return syscall.Unmount(m.MountPath(), 0)
}
