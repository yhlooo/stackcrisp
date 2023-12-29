package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// NewDefaultGlobalOptions 返回默认全局选项
func NewDefaultGlobalOptions() GlobalOptions {
	return GlobalOptions{
		Verbosity: 0,
		DataRoot:  "/var/lib/stackcrisp",
	}
}

// GlobalOptions 全局选项
type GlobalOptions struct {
	// 日志数量级别（ 0 / 1 / 2 ）
	Verbosity uint32 `json:"verbosity" yaml:"verbosity"`
	// 数据存储根目录
	DataRoot string `json:"dataRoot" yaml:"dataRoot"`
}

// Validate 校验选项是否合法
func (o *GlobalOptions) Validate() error {
	if o.Verbosity < 0 || o.Verbosity > 2 {
		return fmt.Errorf("invalid log verbosity: %d (expected: 0, 1 or 2)", o.Verbosity)
	}
	return nil
}

// AddPFlags 将选项绑定到命令行参数
func (o *GlobalOptions) AddPFlags(flags *pflag.FlagSet) {
	flags.Uint32VarP(&o.Verbosity, "verbose", "v", o.Verbosity, "Number for the log level verbosity (0, 1, or 2)")
	flags.StringVar(&o.DataRoot, "data-root", o.DataRoot, "Root directory of persistent data")
}

// GlobalOptionsGetter 全局选项查看器
type GlobalOptionsGetter interface {
	// GetDataRoot 数据存储根目录
	GetDataRoot() string
}

var _ GlobalOptionsGetter = &GlobalOptions{}

// GetDataRoot 数据存储根目录
func (o *GlobalOptions) GetDataRoot() string {
	if o == nil {
		return ""
	}
	return o.DataRoot
}
