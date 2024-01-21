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

// NewTagCommandWithOptions 创建一个基于选项的 tag 命令
func NewTagCommandWithOptions(opts *options.TagOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tag",
		Short:   "Create, list or delete a tag",
		GroupID: groupWork,
		Annotations: map[string]string{
			cmdutil.AnnotationRunAsRoot:      cmdutil.AnnotationValueTrue,
			cmdutil.AnnotationRequireManager: cmdutil.AnnotationValueTrue,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// 获取管理器
			mgr := cmdutil.ManagerFromContext(ctx)

			// 找到当前目录对应 workspace
			ws, err := mgr.GetWorkspaceFromPath(ctx, ".")
			if err != nil {
				return fmt.Errorf("get workspace from path \".\" error: %w", err)
			}

			switch {
			case opts.List:
				// 列出标签
				return runListTags(ctx, ws)
			case opts.Delete:
				// 删除标签
				return runDeleteTag(ctx, ws, args)
			default:
				if len(args) == 0 {
					// 列出标签
					return runListTags(ctx, ws)
				}
				// 添加标签
				return runAddTag(ctx, ws, args, opts)
			}
		},
	}

	// 绑定选项到命令行参数
	opts.AddPFlags(cmd.Flags())

	return cmd
}

// runListTags 列出标签
func runListTags(ctx context.Context, ws workspaces.Workspace) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)
	logger.V(1).Info("list tags")
	for _, t := range ws.Tags() {
		fmt.Println(t)
	}
	return nil
}

// runDeleteTag 删除标签
func runDeleteTag(ctx context.Context, ws workspaces.Workspace, args []string) error {
	for _, tagName := range args {
		logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)
		logger.V(1).Info(fmt.Sprintf("delete tag %q", tagName))
		if err := ws.DeleteTag(ctx, tagName); err != nil {
			return err
		}
	}
	return nil
}

// runAddTag 添加标签
func runAddTag(ctx context.Context, ws workspaces.Workspace, args []string, opts *options.TagOptions) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)
	// 获取参数
	tagName := args[0]
	ref := "HEAD"
	if len(args) > 1 {
		ref = args[1]
	}
	logger.V(1).Info(fmt.Sprintf("add tag %q to %q", tagName, ref))
	return ws.AddTag(ctx, tagName, ref, opts.Force)
}
