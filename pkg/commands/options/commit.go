package options

import "github.com/spf13/pflag"

// NewDefaultCommitOptions 创建一个默认 commit 命令选项
func NewDefaultCommitOptions() CommitOptions {
	return CommitOptions{
		Message: "",
	}
}

// CommitOptions commit 命令选项
type CommitOptions struct {
	// commit 信息
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

// AddPFlags 将选项绑定到命令行参数
func (o *CommitOptions) AddPFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&o.Message, "message", "m", o.Message,
		"Use the given message as the commit message. "+
			"If multiple -m options are given, their values are concatenated as separate paragraphs.",
	)
}
