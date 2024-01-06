package commands

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
	cmdutil "github.com/yhlooo/stackcrisp/pkg/utils/cmd"
)

func NewLogCommandWithOptions(_ *options.LogOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "log [<revision>]",
		Short:   "Show commit logs",
		GroupID: groupState,
		Annotations: map[string]string{
			cmdutil.AnnotationRequireManager: cmdutil.AnnotationValueTrue,
		},
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

			revision := "HEAD"
			if len(args) > 0 {
				revision = args[0]
			}
			logger.V(1).Info(fmt.Sprintf("revision: %q", revision))

			// 获取管理器
			mgr := cmdutil.ManagerFromContext(ctx)

			// 获取工作空间
			ws, err := mgr.GetWorkspaceFromPath(ctx, ".")
			if err != nil {
				return fmt.Errorf("get workspace from path \".\" error: %w", err)
			}

			// 获取提交历史
			commits, err := mgr.GetHistory(ctx, ws, revision)
			if err != nil {
				return fmt.Errorf("get history of revision %q error: %w", err)
			}

			// 打印
			for _, c := range commits {
				fmt.Printf("commit %s\n", c.ID.Hex())
			}

			return nil
		},
	}
	return cmd
}
