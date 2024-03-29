package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
	"github.com/yhlooo/stackcrisp/pkg/manager"
	logutil "github.com/yhlooo/stackcrisp/pkg/utils/log"
	"github.com/yhlooo/stackcrisp/pkg/utils/sudo"
)

const (
	loggerName = "utils.cmd"
)

// SetLogger 设置命令日志，并返回 logr.Logger
func SetLogger(cmd *cobra.Command, verbosity uint32) logr.Logger {
	// 设置日志级别
	logrusLogger := logrus.New()
	switch verbosity {
	case 1:
		logrusLogger.SetLevel(logrus.DebugLevel)
	case 2:
		logrusLogger.SetLevel(logrus.TraceLevel)
	default:
		logrusLogger.SetLevel(logrus.InfoLevel)
	}
	// 将 logger 注入上下文
	logger := logrusr.New(logrusLogger)
	cmd.SetContext(logr.NewContext(cmd.Context(), logger))

	return logger
}

// SwitchToRootIfNecessary 如果需要的话切换到 root 用户运行
func SwitchToRootIfNecessary(cmd *cobra.Command) (bool, error) {
	logger := logr.FromContextOrDiscard(cmd.Context()).WithName(loggerName)
	logutil.UserInfo(logger.V(1))
	if cmd.Annotations[AnnotationRunAsRoot] == AnnotationValueTrue && !sudo.IsRoot() {
		return true, runAsRoot(cmd)
	}
	return false, nil
}

// runAsRoot 设置使用 root 用户运行
func runAsRoot(cmd *cobra.Command) error {
	logger := logr.FromContextOrDiscard(cmd.Context()).WithName(loggerName)

	logger.Info("switch to root")
	cmd.Run = nil
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return sudo.RunAsRoot(cmd.Context(), sudoExtraArgs()...)
	}
	cmd.PrintErr()

	return nil
}

// sudoExtraArgs 切换为 root 用户时需要额外指定的参数
func sudoExtraArgs() []string {
	return []string{
		"--uid",
		strconv.FormatInt(int64(os.Getuid()), 10),
		"--gid",
		strconv.FormatInt(int64(os.Getgid()), 10),
	}
}

// ChangeWorkingDirectory 切换命令工作目录
func ChangeWorkingDirectory(cmd *cobra.Command, path string) error {
	defer func() {
		pwd, _ := os.Getwd()
		logger := logr.FromContextOrDiscard(cmd.Context())
		logger.V(1).Info(fmt.Sprintf("working directory: %q", pwd))
	}()

	if path == "" {
		return nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("get absolute path of %q error: %w", path, err)
	}
	if err := os.Chdir(absPath); err != nil {
		return fmt.Errorf("change working directory to %q error: %w", absPath, err)
	}
	// chdir 之后需要更新一下 PWD 变量，否则 os.Getwd 会判断错误
	if err := os.Setenv("PWD", absPath); err != nil {
		return fmt.Errorf("set env PWD to %q error: %w", absPath, err)
	}

	return nil
}

// InjectManagerIfNecessary 如果需要的话注入 manager.Manager
func InjectManagerIfNecessary(cmd *cobra.Command, globalOptions options.GlobalOptionsGetter) error {
	if cmd.Annotations[AnnotationRequireManager] != AnnotationValueTrue {
		return nil
	}

	logger := logr.FromContextOrDiscard(cmd.Context()).WithName(loggerName)

	// 创建管理器
	logger.V(1).Info(fmt.Sprintf("new manager, dataRoot: %q", globalOptions.GetDataRoot()))
	mgr, err := manager.New(manager.Options{
		DataRoot: globalOptions.GetDataRoot(),
		ChownUID: globalOptions.GetUID(),
		ChownGID: globalOptions.GetGID(),
	})
	if err != nil {
		return fmt.Errorf("create manager error: %w", err)
	}
	if err := mgr.Prepare(cmd.Context()); err != nil {
		return fmt.Errorf("prepare manager error: %w", err)
	}

	// 注入到上下文
	cmd.SetContext(NewContextWithManager(cmd.Context(), mgr))

	return nil
}
