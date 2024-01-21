package options

import "github.com/spf13/pflag"

// NewDefaultBranchOptions 创建一个默认 branch 命令选项
func NewDefaultBranchOptions() BranchOptions {
	return BranchOptions{
		List:        false,
		ShowCurrent: false,
		Move:        false,
		Copy:        false,
		Delete:      false,
		Remotes:     false,
		All:         false,
	}
}

// BranchOptions branch 命令选项
type BranchOptions struct {
	// 列出分支
	List bool `json:"list,omitempty" yaml:"list,omitempty"`
	// 显示当前分支
	ShowCurrent bool `json:"showCurrent,omitempty" yaml:"showCurrent,omitempty"`
	// 移动分支
	Move bool `json:"move,omitempty" yaml:"move,omitempty"`
	// 拷贝分支
	Copy bool `json:"copy,omitempty" yaml:"copy,omitempty"`
	// 删除分支
	Delete bool `json:"delete,omitempty" yaml:"delete,omitempty"`
	// 列出或删除远程分支
	Remotes bool `json:"remotes,omitempty" yaml:"remotes,omitempty"`
	// 列出所有分支
	All bool `json:"all,omitempty" yaml:"all,omitempty"`
	// 强制创建分支
	Force bool `json:"force,omitempty" yaml:"force,omitempty"`
}

// AddPFlags 将选项绑定到命令行参数
func (o *BranchOptions) AddPFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&o.List, "list", "l", o.List, "List branches.")
	flags.BoolVar(
		&o.ShowCurrent, "show-current", o.ShowCurrent,
		"Print the name of the current branch. In detached HEAD state, nothing is printed.",
	)
	flags.BoolVarP(&o.Move, "move", "m", o.Move, "Move/rename a branch, together with its config and reflog.")
	flags.BoolVarP(&o.Copy, "copy", "c", o.Copy, "Copy a branch, together with its config and reflog.")
	flags.BoolVarP(&o.Delete, "delete", "d", o.Delete, "Delete a branch.")
	flags.BoolVarP(&o.Remotes, "remotes", "r", o.Remotes, "List or delete the remote-tracking branches.")
	flags.BoolVarP(&o.All, "all", "a", o.All, "List both remote-tracking branches and local branches.")
	flags.BoolVarP(
		&o.Force, "force", "f", o.Force,
		"Replace an existing branch with the given name (instead of failing).",
	)
}
