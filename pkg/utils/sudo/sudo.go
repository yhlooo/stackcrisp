//go:build !windows

package sudo

import (
	"context"
	"os"
	"os/exec"
)

// IsRoot 返回是否 root 用户
func IsRoot() bool {
	return os.Getuid() == 0
}

// RunAsRoot 在子进程中以 root 身份运行当前命令
func RunAsRoot(ctx context.Context, extraArgs ...string) error {
	// 命令前加个 sudo -E
	cmd := exec.CommandContext(ctx, "sudo", append([]string{"-E"}, append(os.Args, extraArgs...)...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
