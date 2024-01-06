package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
)

const (
	loggerName = "commands"
)

// NewStackCrispCommandWithOptions 创建一个基于选项的 stackcrisp 命令
func NewStackCrispCommandWithOptions(opts options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "stackcrisp",
		Short:        "Manage OverlayFS mounts with git-like commands.",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 校验全局选项
			if err := opts.Global.Validate(); err != nil {
				return err
			}
			// 设置日志
			logger := setLogger(cmd, opts.Global.Verbosity)
			// 切换到 root
			if need, err := switchToRoot(cmd); need {
				return err
			}
			// 设置工作目录
			if err := changeWorkingDirectory(cmd, opts.Global.Chdir); err != nil {
				return err
			}

			logger.V(1).Info(fmt.Sprintf("command: %q, args: %#v", cmd.Name(), args))
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// 绑定选项到命令行参数
	opts.Global.AddPFlags(cmd.PersistentFlags())

	// 添加子命令
	cmd.AddCommand(
		NewInitCommandWithOptions(opts.Init, &opts.Global),
		NewCommitCommandWithOptions(opts.Commit, &opts.Global),
	)

	return cmd
}

// NewStackCrispCommand 使用默认选项创建一个 stackcrisp 命令
func NewStackCrispCommand() *cobra.Command {
	return NewStackCrispCommandWithOptions(options.NewDefaultOptions())
}
