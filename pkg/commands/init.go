package commands

import (
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
	"github.com/yhlooo/stackcrisp/pkg/manager"
	fsutil "github.com/yhlooo/stackcrisp/pkg/utils/fs"
)

// NewInitCommandWithOptions 创建一个基于选项的 init 命令
func NewInitCommandWithOptions(_ options.InitOptions, globalOptions options.GlobalOptionsGetter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [PATH]",
		Short: "Create an empty Space or reinitialize an existing one.",
		Annotations: map[string]string{
			AnnotationRunAsRoot: AnnotationValueTrue,
		},
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

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
			if err := mgr.Prepare(ctx); err != nil {
				return fmt.Errorf("prepare manager error: %w", err)
			}
			// 创建 workspace
			logger.Info("creating workspace ...")
			ws, err := mgr.CreateWorkspace(ctx, targetAbsPath)
			if err != nil {
				return fmt.Errorf("create workspacespace error: %w", err)
			}

			// 展开 workspace
			logger.Info("expanding workspace ...")
			if err := ws.Expand(ctx); err != nil {
				return fmt.Errorf("expand workspace error: %w", err)
			}

			return nil
		},
	}
	return cmd
}
