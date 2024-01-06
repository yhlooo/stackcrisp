//go:build windows

package sudo

import (
	"context"
	"fmt"
	"runtime"
)

// IsRoot 返回是否 root 用户
// TODO: 暂未实现
func IsRoot() bool {
	return true
}

// RunAsRoot 在子进程中以 root 身份运行当前命令
func RunAsRoot(context.Context, ...string) error {
	return fmt.Errorf("sudo is not supported on %s", runtime.GOOS)
}
