package commands

import (
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
	cmdutil "github.com/yhlooo/stackcrisp/pkg/utils/cmd"
)

// NewCloneCommandWithOptions 创建一个基于选项的 clone 命令
func NewCloneCommandWithOptions(_ *options.CloneOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clone <workspace> [<directory>]",
		Short:   "Clone a workspace into a new directory",
		GroupID: groupStart,
		Annotations: map[string]string{
			cmdutil.AnnotationRunAsRoot:      cmdutil.AnnotationValueTrue,
			cmdutil.AnnotationRequireManager: cmdutil.AnnotationValueTrue,
		},
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

			source := args[0]
			target := "."
			if len(args) > 1 {
				target = args[1]
			}
			targetAbsPath, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("get absolute path of %q error: %w", target, err)
			}
			logger.V(1).Info(fmt.Sprintf("source: %q, target path: %q", source, target))

			// 获取管理器
			mgr := cmdutil.ManagerFromContext(ctx)

			// 获取源工作空间
			// TODO: 应该还要支持从 space id / 名获取
			sourceWS, err := mgr.GetWorkspaceFromPath(ctx, source)
			if err != nil {
				return fmt.Errorf("get workspace from path %q error: %w", source, err)
			}

			// 克隆
			logger.Info("cloning workspace ...")
			targetWS, err := mgr.Clone(ctx, sourceWS, targetAbsPath)
			if err != nil {
				return fmt.Errorf("clone workspace error: %w", err)
			}

			// 展开 workspace
			logger.Info("expanding workspace ...")
			if err := targetWS.Expand(ctx); err != nil {
				return fmt.Errorf("expand workspace error: %w", err)
			}

			return nil
		},
	}

	return cmd
}
