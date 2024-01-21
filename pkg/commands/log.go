package commands

import (
	"fmt"
	"strings"
	"time"

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

			ref := "HEAD"
			if len(args) > 0 {
				ref = args[0]
			}
			logger.V(1).Info(fmt.Sprintf("revision: %q", ref))

			// 获取管理器
			mgr := cmdutil.ManagerFromContext(ctx)

			// 获取工作空间
			ws, err := mgr.GetWorkspaceFromPath(ctx, ".")
			if err != nil {
				return fmt.Errorf("get workspace from path \".\" error: %w", err)
			}

			// 获取提交历史
			commits, err := ws.GetHistory(ref)
			if err != nil {
				return fmt.Errorf("get history of revision %q error: %w", ref, err)
			}

			// 打印
			for _, c := range commits {
				var pointers []string
				for _, t := range c.Tags() {
					pointers = append(pointers, fmt.Sprintf("\033[33mtag: %s\033[0m", t))
				}
				for _, b := range c.Branches() {
					if ws.Branch().FullName() == b.FullName() {
						pointers = append(
							[]string{fmt.Sprintf("\033[34mHEAD -> \033[32m%s\033[0m", b.LocalName())},
							pointers...,
						)
					} else {
						if b.IsLocal() {
							pointers = append(pointers, fmt.Sprintf("\033[34m%s\033[0m", b.LocalName()))
						} else {
							pointers = append(pointers, fmt.Sprintf("\033[31m%s\033[0m", b.LocalName()))
						}
					}
				}

				if pointers != nil {
					fmt.Printf(
						"\033[33mcommit %s\033[0m (%s)\n",
						c.ID().Hex(), strings.Join(pointers, "\033[33m, \033[0m"),
					)
				} else {
					fmt.Printf("\033[33mcommit %s\033[0m\n", c.ID().Hex())
				}
				if c.Date != nil {
					fmt.Printf("Date:  %s\n", c.Date().Format(time.ANSIC+" -0700"))
				}
				if c.Message() != "" {
					fmt.Println()
					fmt.Println("    " + strings.ReplaceAll(strings.TrimRight(c.Message(), "\r\n "), "\n", "\n    "))
					fmt.Println()
				}
			}

			return nil
		},
	}
	return cmd
}
