package commands

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
	cmdutil "github.com/yhlooo/stackcrisp/pkg/utils/cmd"
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
		Args: cobra.MaximumNArgs(2),
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

			switch {
			case opts.List || len(args) == 0:
				// 列出标签
				for t := range ws.Space().Tree().Tags() {
					if t == "ROOT" {
						continue
					}
					fmt.Println(t)
				}
			case opts.Delete:
				// 删除标签
				tag := args[0]
				logger.V(1).Info(fmt.Sprintf("delete tag %q", tag))
				if ok := ws.Space().Tree().DeleteTag(tag); !ok {
					return fmt.Errorf("tag %q not found", tag)
				}
			default:
				// 添加标签

				// 获取参数
				tag := args[0]
				revision := "HEAD"
				if len(args) > 1 {
					revision = args[1]
				}
				logger.V(1).Info(fmt.Sprintf("add tag %q to %q", tag, revision))

				// 查找节点
				node, _, ok := ws.Search(revision)
				if !ok {
					return fmt.Errorf("failed to resolve %q as valid ref", revision)
				}

				// 检查是否已经存在该标签
				if existsNode, ok := ws.Space().Tree().GetByTag(tag); ok && !opts.Force {
					return fmt.Errorf("tag %q already exists at %q", tag, existsNode.ID().Hex())
				}

				if err := ws.Space().Tree().AddTag(tag, node.ID()); err != nil {
					return err
				}
				if err := ws.Space().Save(ctx); err != nil {
					return fmt.Errorf("save space info error: %w", err)
				}
			}
			return nil
		},
	}

	// 绑定选项到命令行参数
	opts.AddPFlags(cmd.Flags())

	return cmd
}
