//go:build linux

package mounts

import (
	"fmt"
	"os"
	"syscall"

	fsutil "github.com/yhlooo/stackcrisp/pkg/utils/fs"
)

// Mount 挂载
func (m *defaultMount) Mount() error {
	mountPath := m.ovlOpts.MountPath
	workDir := m.ovlOpts.WorkDir

	// 确保相关目录存在
	if !fsutil.IsDir(mountPath) {
		if err := os.Mkdir(mountPath, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("mkdir %q error: %w", mountPath, err)
		}
	}
	if !fsutil.IsDir(workDir) {
		if err := os.Mkdir(workDir, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("mkdir %q error: %w", workDir, err)
		}
	}

	return CreateOverlayMount(m.ovlOpts)
}

// Umount 卸载挂载
func (m *defaultMount) Umount() error {
	return syscall.Unmount(m.MountPath(), 0)
}
