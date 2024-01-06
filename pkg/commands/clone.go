package commands

import (
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/stackcrisp/pkg/commands/options"
	cmdutil "github.com/yhlooo/stackcrisp/pkg/utils/cmd"
)

// NewCloneCommandWithOptions 创建一个基于选项的 clone 命令
func NewCloneCommandWithOptions(_ options.CloneOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone <workspace> [<directory>]",
		Short: "Clone a workspace into a new directory.",
		Annotations: map[string]string{
			cmdutil.AnnotationRunAsRoot:      cmdutil.AnnotationValueTrue,
			cmdutil.AnnotationRequireManager: cmdutil.AnnotationValueTrue,
		},
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)
			_ = logger
			// TODO: ...
			return nil
		},
	}

	return cmd
}
