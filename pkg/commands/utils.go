package commands

import (
	"os"
	"strconv"
)

// sudoExtraArgs 切换为 root 用户时需要额外指定的参数
func sudoExtraArgs() []string {
	return []string{
		"--uid",
		strconv.FormatInt(int64(os.Getuid()), 10),
		"--gid",
		strconv.FormatInt(int64(os.Getgid()), 10),
	}
}
