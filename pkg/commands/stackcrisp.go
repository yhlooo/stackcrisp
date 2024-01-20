package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
	cmdutil "github.com/yhlooo/stackcrisp/pkg/utils/cmd"
)

const (
	loggerName = "commands"

	groupStart = "start"
	groupWork  = "work"
	groupState = "state"
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
			logger := cmdutil.SetLogger(cmd, opts.Global.Verbosity)
			// 切换到 root
			if need, err := cmdutil.SwitchToRootIfNecessary(cmd); need {
				return err
			}
			// 设置工作目录
			if err := cmdutil.ChangeWorkingDirectory(cmd, opts.Global.Chdir); err != nil {
				return err
			}
			// 创建并注入 manager
			if err := cmdutil.InjectManagerIfNecessary(cmd, &opts.Global); err != nil {
				return err
			}

			logger.V(1).Info(fmt.Sprintf("command: %q, args: %#v, options: %#v", cmd.Name(), args, opts))
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// 绑定选项到命令行参数
	opts.Global.AddPFlags(cmd.PersistentFlags())

	// 添加命令组
	cmd.AddGroup(
		&cobra.Group{ID: groupStart, Title: "Start a working area"},
		&cobra.Group{ID: groupWork, Title: "Grow, mark and tweak your common history"},
		&cobra.Group{ID: groupState, Title: "Examine the history and state"},
	)

	// 添加子命令
	cmd.AddCommand(
		NewInitCommandWithOptions(&opts.Init),
		NewCloneCommandWithOptions(&opts.Clone),
		NewCommitCommandWithOptions(&opts.Commit),
		NewCheckoutCommandWithOptions(&opts.Checkout),
		NewBranchCommandWithOptions(&opts.Branch),
		NewTagCommandWithOptions(&opts.Tag),
		NewLogCommandWithOptions(&opts.Log),
	)

	return cmd
}

// NewStackCrispCommand 使用默认选项创建一个 stackcrisp 命令
func NewStackCrispCommand() *cobra.Command {
	return NewStackCrispCommandWithOptions(options.NewDefaultOptions())
}
