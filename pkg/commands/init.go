package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
	"github.com/yhlooo/stackcrisp/pkg/manager"
	fsutil "github.com/yhlooo/stackcrisp/pkg/utils/fs"
)

// NewInitCommandWithOptions 创建一个基于选项的 init 命令
func NewInitCommandWithOptions(_ options.InitOptions, globalOptions options.GlobalOptionsGetter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create an empty Space or reinitialize an existing one.",
		RunE: func(cmd *cobra.Command, args []string) error {
			workingDir, err := filepath.Abs(".")
			if err != nil {
				return fmt.Errorf("get absolute path of \".\" error: %w", err)
			}

			ctx := cmd.Context()

			// 创建管理器
			mgr := manager.New(manager.Options{
				DataRoot: globalOptions.GetDataRoot(),
			})
			if err := mgr.Prepare(ctx); err != nil {
				return fmt.Errorf("prepare manager error: %w", err)
			}
			// 创建 space
			space, err := mgr.CreateSpace(ctx)
			if err != nil {
				return fmt.Errorf("create space error: %w", err)
			}

			// 挂载
			// TODO: 分支应当可以指定，而且应当支持分支在本地工作区的别名
			mount, err := mgr.CreateMount(
				ctx, space,
				"ROOT",
			)
			if err != nil {
				return fmt.Errorf("create mount error: %w", err)
			}

			// 确保目标路径上不是一个非空目录或文件
			if fsutil.IsExists(workingDir) && !fsutil.IsEmptyDir(workingDir) {
				return fmt.Errorf("path %q is not an empty dir", workingDir)
			}
			// 确保目标路径上什么也没有
			if fsutil.IsExists(workingDir) {
				if err := os.Remove(workingDir); err != nil {
					return fmt.Errorf("clear path %q error: %w", workingDir, err)
				}
			}
			// 创建软链
			if err := mount.CreateSymlink(workingDir); err != nil {
				return fmt.Errorf("create symlink %q to mount path error: %w", workingDir, err)
			}

			return nil
		},
	}
	return cmd
}
