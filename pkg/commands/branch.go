package commands

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
	cmdutil "github.com/yhlooo/stackcrisp/pkg/utils/cmd"
	"github.com/yhlooo/stackcrisp/pkg/workspaces"
)

// NewBranchCommandWithOptions 创建一个基于选项的 branch 命令
func NewBranchCommandWithOptions(opts *options.BranchOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "branch",
		Short:   "List, create, or delete branches",
		GroupID: groupWork,
		Annotations: map[string]string{
			cmdutil.AnnotationRunAsRoot:      cmdutil.AnnotationValueTrue,
			cmdutil.AnnotationRequireManager: cmdutil.AnnotationValueTrue,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)
			_ = logger

			// 获取管理器
			mgr := cmdutil.ManagerFromContext(ctx)

			// 找到当前目录对应 workspace
			ws, err := mgr.GetWorkspaceFromPath(ctx, ".")
			if err != nil {
				return fmt.Errorf("get workspace from path \".\" error: %w", err)
			}

			switch {
			case opts.List:
				// 列出分支
				var branches []workspaces.Branch
				switch {
				case opts.Remotes:
					branches = ws.RemoteBranches()
				case opts.All:
					branches = ws.AllBranches()
				default:
					branches = ws.LocalBranches()
				}
				for _, name := range branches {
					fmt.Println(name.LocalName())
				}
			case opts.ShowCurrent:
				// 显示当前分支
				if branch := ws.Branch(); branch != nil {
					fmt.Println(branch.LocalName())
				}
			case opts.Move:
				// TODO: ...
			case opts.Copy:
				// TODO: ...
			case opts.Delete:
				// TODO: ...
			}
			return nil
		},
	}

	// 绑定选项到命令行参数
	opts.AddPFlags(cmd.Flags())

	return cmd
}
