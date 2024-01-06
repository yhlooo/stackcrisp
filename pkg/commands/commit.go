package commands

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
	"github.com/yhlooo/stackcrisp/pkg/manager"
	cmdutil "github.com/yhlooo/stackcrisp/pkg/utils/cmd"
)

// NewCommitCommandWithOptions 创建一个基于选项的 commit 命令
func NewCommitCommandWithOptions(opts options.CommitOptions, globalOptions options.GlobalOptionsGetter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Record changes to the space.",
		Annotations: map[string]string{
			cmdutil.AnnotationRunAsRoot: cmdutil.AnnotationValueTrue,
		},
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

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

			// 找到当前目录对应 workspace
			ws, err := mgr.GetWorkspaceFromPath(ctx, ".")
			if err != nil {
				return fmt.Errorf("get workspace from path \".\" error: %w", err)
			}

			// commit
			newWS, err := mgr.Commit(ctx, ws)
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
