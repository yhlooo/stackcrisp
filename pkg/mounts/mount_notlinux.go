//go:build !linux

package mounts

import (
	"context"
	"fmt"
	"runtime"
)

// Mount 挂载
func (m *defaultMount) Mount(context.Context) error {
	return fmt.Errorf("mount is not supported on %s", runtime.GOOS)
}

// Umount 卸载挂载
func (m *defaultMount) Umount(context.Context) error {
	return fmt.Errorf("umount is not supported on %s", runtime.GOOS)
}
