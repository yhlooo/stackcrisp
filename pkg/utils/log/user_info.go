package log

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
)

// UserInfo 日志输出用户信息
func UserInfo(logger logr.Logger) {
	logger.Info(fmt.Sprintf("uid: %d, gid: %d", os.Getuid(), os.Getgid()))
}
