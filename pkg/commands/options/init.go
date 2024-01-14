package options

import "github.com/spf13/pflag"

// NewDefaultInitOptions 创建一个默认 init 命令选项
func NewDefaultInitOptions() InitOptions {
	return InitOptions{
		InitialBranch: "main",
	}
}

// InitOptions init 命令选项
type InitOptions struct {
	// 初始分支
	InitialBranch string `json:"initialBranch,omitempty" yaml:"initialBranch,omitempty"`
}

// AddPFlags 将选项绑定到命令行参数
func (o *InitOptions) AddPFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&o.InitialBranch,
		"initial-branch", "b", o.InitialBranch,
		"Use the specified name for the initial branch in the newly created space.",
	)
}
