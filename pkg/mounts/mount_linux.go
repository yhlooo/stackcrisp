//go:build linux

package mounts

import "syscall"

// Umount 卸载挂载
func (m *defaultMount) Umount() error {
	return syscall.Unmount(m.MountPath(), 0)
}
