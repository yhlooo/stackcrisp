package commands

import (
	"context"
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
				return runListBranch(ctx, ws, opts)
			case opts.ShowCurrent:
				// 显示当前分支
				if branch := ws.Branch(); branch != nil {
					fmt.Println(branch.LocalName())
				}
			case opts.Delete:
				// 删除分支
				return runDeleteBranch(ctx, ws, args, opts)
			default:
				if len(args) == 0 {
					// 列出分支
					return runListBranch(ctx, ws, opts)
				}
				// 创建分支
				return runAddBranch(ctx, ws, args, opts)
			}
			return nil
		},
	}

	// 绑定选项到命令行参数
	opts.AddPFlags(cmd.Flags())

	return cmd
}

// runListBranch 列出分支
func runListBranch(ctx context.Context, ws workspaces.Workspace, opts *options.BranchOptions) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)
	logger.V(1).Info(fmt.Sprintf("list branches, remote: %t, all: %t", opts.Remotes, opts.All))
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
	return nil
}

// runAddBranch 添加分支
func runAddBranch(ctx context.Context, ws workspaces.Workspace, args []string, opts *options.BranchOptions) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)
	branchLocalName := args[0]
	ref := "HEAD"
	if len(args) > 1 {
		ref = args[1]
	}
	logger.V(1).Info(fmt.Sprintf("add branch %q to %q", branchLocalName, ref))
	return ws.AddBranch(ctx, branchLocalName, ref, opts.Force)
}

// runDeleteBranch 删除分支
func runDeleteBranch(ctx context.Context, ws workspaces.Workspace, args []string, opts *options.BranchOptions) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)
	for _, branchName := range args {
		logger.V(1).Info(fmt.Sprintf("delete branch %q, remote: %t", branchName, opts.Remotes))
		if err := ws.DeleteBranch(ctx, branchName, opts.Remotes); err != nil {
			return err
		}
	}
	return nil
}
