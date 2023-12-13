//go:build !linux

package mounts

import (
	"fmt"
	"runtime"
)

// CreateOverlayMount 创建一个 Overlay 挂载
func CreateOverlayMount(OverlayMountOptions) error {
	return fmt.Errorf("overla is not supported on %s", runtime.GOOS)
}
