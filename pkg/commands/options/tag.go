package options

import "github.com/spf13/pflag"

// NewDefaultTagOptions 创建一个默认 tag 命令选项
func NewDefaultTagOptions() TagOptions {
	return TagOptions{
		List:   false,
		Delete: false,
		Force:  false,
	}
}

// TagOptions tag 命令选项
type TagOptions struct {
	// 列出标签
	List bool `json:"list,omitempty" yaml:"list,omitempty"`
	// 删除标签
	Delete bool `json:"delete,omitempty" yaml:"delete,omitempty"`
	// 强制创建标签
	Force bool `json:"force,omitempty" yaml:"force,omitempty"`
}

// AddPFlags 将选项绑定到命令行参数
func (o *TagOptions) AddPFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&o.List, "list", "l", o.List, "List tags.")
	flags.BoolVarP(&o.Delete, "delete", "d", o.Delete, "Delete a tag.")
	flags.BoolVarP(
		&o.Force, "force", "f", o.Force,
		"Replace an existing tag with the given name (instead of failing).",
	)
}
