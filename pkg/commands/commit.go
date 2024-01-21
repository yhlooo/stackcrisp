package commands

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
	cmdutil "github.com/yhlooo/stackcrisp/pkg/utils/cmd"
	"github.com/yhlooo/stackcrisp/pkg/workspaces"
)

// NewCommitCommandWithOptions 创建一个基于选项的 commit 命令
func NewCommitCommandWithOptions(opts *options.CommitOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "commit",
		Short:   "Record changes to the space",
		GroupID: groupWork,
		Annotations: map[string]string{
			cmdutil.AnnotationRunAsRoot:      cmdutil.AnnotationValueTrue,
			cmdutil.AnnotationRequireManager: cmdutil.AnnotationValueTrue,
		},
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

			// 获取管理器
			mgr := cmdutil.ManagerFromContext(ctx)

			// 找到当前目录对应 workspace
			ws, err := mgr.GetWorkspaceFromPath(ctx, ".")
			if err != nil {
				return fmt.Errorf("get workspace from path \".\" error: %w", err)
			}

			// commit
			newWS, err := mgr.Commit(ctx, ws, workspaces.NewCommitInfo(opts.Message))
			if err != nil {
				return fmt.Errorf("commit error: %w", err)
			}

			// 展开 workspace
			logger.Info("expanding workspace ...")
			if err := newWS.Expand(ctx); err != nil {
				return fmt.Errorf("expand workspace error: %w", err)
			}

			// 回收旧的 workspace
			logger.Info("removing old workspace mount ...")
			if err := mgr.RemoveWorkspaceMount(ctx, ws); err != nil {
				return fmt.Errorf("remove old workspace mount error: %w", err)
			}

			return nil
		},
	}

	// 绑定选项到命令行参数
	opts.AddPFlags(cmd.Flags())

	return cmd
}
