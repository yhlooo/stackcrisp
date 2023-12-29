package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
	"github.com/yhlooo/stackcrisp/pkg/manager"
	fsutil "github.com/yhlooo/stackcrisp/pkg/utils/fs"
	logutil "github.com/yhlooo/stackcrisp/pkg/utils/log"
	"github.com/yhlooo/stackcrisp/pkg/utils/sudo"
)

// NewInitCommandWithOptions 创建一个基于选项的 init 命令
func NewInitCommandWithOptions(_ options.InitOptions, globalOptions options.GlobalOptionsGetter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [PATH]",
		Short: "Create an empty Space or reinitialize an existing one.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)
			logger.V(1).Info(fmt.Sprintf("command: \"init\", args: %#v", args))

			// 切换到 root
			logutil.UserInfo(logger.V(1))
			if !sudo.IsRoot() {
				logger.Info("switch to root")
				return sudo.RunAsRoot(ctx, sudoExtraArgs()...)
			}

			target := "."
			if len(args) > 0 {
				target = args[0]
			}
			targetAbsPath, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("get absolute path of %q error: %w", target, err)
			}
			logger.V(1).Info(fmt.Sprintf("target path: %q", targetAbsPath))

			// 确保目标路径上不是一个非空目录或文件
			if fsutil.IsExists(targetAbsPath) && !fsutil.IsEmptyDir(targetAbsPath) {
				return fmt.Errorf("path %q is not an empty dir", targetAbsPath)
			}
			// 确保目标路径上什么也没有
			if fsutil.IsExists(targetAbsPath) {
				logger.V(1).Info(fmt.Sprintf("target path %q exists, remove it", targetAbsPath))
				if err := os.Remove(targetAbsPath); err != nil {
					return fmt.Errorf("clear path %q error: %w", targetAbsPath, err)
				}
			}

			// 创建管理器
			logger.V(1).Info(fmt.Sprintf("new manager, dataRoot: %q", globalOptions.GetDataRoot()))
			mgr := manager.New(manager.Options{
				DataRoot: globalOptions.GetDataRoot(),
				ChownUID: globalOptions.GetUID(),
				ChownGID: globalOptions.GetGID(),
			})
			if err := mgr.Prepare(ctx); err != nil {
				return fmt.Errorf("prepare manager error: %w", err)
			}
			// 创建 space
			logger.Info("creating space ...")
			space, err := mgr.CreateSpace(ctx)
			if err != nil {
				return fmt.Errorf("create space error: %w", err)
			}

			// 挂载
			// TODO: 分支应当可以指定，而且应当支持分支在本地工作区的别名
			logger.Info("creating mount ...")
			mount, err := mgr.CreateMount(
				ctx, space,
				"ROOT",
			)
			if err != nil {
				return fmt.Errorf("create mount error: %w", err)
			}

			// 创建软链
			logger.Info(fmt.Sprintf("link mount point to %q", targetAbsPath))
			if err := mount.CreateSymlink(ctx, targetAbsPath); err != nil {
				return fmt.Errorf("create symlink %q to mount path error: %w", targetAbsPath, err)
			}

			return nil
		},
	}
	return cmd
}
