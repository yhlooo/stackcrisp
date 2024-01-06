package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// NewDefaultGlobalOptions 返回默认全局选项
func NewDefaultGlobalOptions() GlobalOptions {
	return GlobalOptions{
		Verbosity: 0,
		Chdir:     "",
		DataRoot:  "/var/lib/stackcrisp",
		UID:       -1,
		GID:       -1,
	}
}

// GlobalOptions 全局选项
type GlobalOptions struct {
	// 日志数量级别（ 0 / 1 / 2 ）
	Verbosity uint32 `json:"verbosity" yaml:"verbosity"`
	// 改变工作目录
	Chdir string `json:"chdir,omitempty" yaml:"chdir,omitempty"`
	// 数据存储根目录
	DataRoot string `json:"dataRoot" yaml:"dataRoot"`
	// 执行命令的原始用户 ID
	UID int `json:"uid" yaml:"uid"`
	// 执行命令的原始用户组 ID
	GID int `json:"gid" yaml:"gid"`
}

// Validate 校验选项是否合法
func (o *GlobalOptions) Validate() error {
	if o.Verbosity > 2 {
		return fmt.Errorf("invalid log verbosity: %d (expected: 0, 1 or 2)", o.Verbosity)
	}
	return nil
}

// AddPFlags 将选项绑定到命令行参数
func (o *GlobalOptions) AddPFlags(flags *pflag.FlagSet) {
	flags.Uint32VarP(&o.Verbosity, "verbose", "v", o.Verbosity, "Number for the log level verbosity (0, 1, or 2)")
	flags.StringVarP(&o.Chdir, "chdir", "C", o.Chdir, "Change to directory before doing anything")

	flags.StringVar(&o.DataRoot, "data-root", o.DataRoot, "Root directory of persistent data")
	flags.IntVar(&o.UID, "uid", o.UID, "The uid of the user who executed the original command")
	flags.IntVar(&o.GID, "gid", o.GID, "The uid of the user who executed the original command")
}

// GlobalOptionsGetter 全局选项查看器
type GlobalOptionsGetter interface {
	// GetDataRoot 数据存储根目录
	GetDataRoot() string
	// GetUID 执行命令的原始用户 ID
	GetUID() int
	// GetGID 执行命令的原始用户组 ID
	GetGID() int
}

var _ GlobalOptionsGetter = &GlobalOptions{}

// GetDataRoot 数据存储根目录
func (o *GlobalOptions) GetDataRoot() string {
	if o == nil {
		return ""
	}
	return o.DataRoot
}

// GetUID 执行命令的原始用户 ID
func (o *GlobalOptions) GetUID() int {
	return o.UID
}

// GetGID 执行命令的原始用户组 ID
func (o *GlobalOptions) GetGID() int {
	return o.GID
}
