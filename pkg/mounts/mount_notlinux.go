//go:build !linux

package mounts

import (
	"fmt"
	"runtime"
)

// Umount 卸载挂载
func (m *defaultMount) Umount() error {
	return fmt.Errorf("umount is not supported on %s", runtime.GOOS)
}
