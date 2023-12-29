package commands

import (
	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
)

// NewStackCrispCommandWithOptions 创建一个基于选项的 stackcrisp 命令
func NewStackCrispCommandWithOptions(opts options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stackcrisp",
		Short: "Manage OverlayFS mounts with git-like commands.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 校验全局选项
			if err := opts.Global.Validate(); err != nil {
				return err
			}
			// 设置日志级别
			logrusLogger := logrus.New()
			switch opts.Global.Verbosity {
			case 1:
				logrusLogger.SetLevel(logrus.DebugLevel)
			case 2:
				logrusLogger.SetLevel(logrus.TraceLevel)
			default:
				logrusLogger.SetLevel(logrus.InfoLevel)
			}
			// 将 logger 注入上下文
			logger := logrusr.New(logrusLogger)
			cmd.SetContext(logr.NewContext(cmd.Context(), logger))
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
	)

	return cmd
}

// NewStackCrispCommand 使用默认选项创建一个 stackcrisp 命令
func NewStackCrispCommand() *cobra.Command {
	return NewStackCrispCommandWithOptions(options.NewDefaultOptions())
}
