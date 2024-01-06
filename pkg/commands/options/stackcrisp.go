package options

// NewDefaultOptions 创建一个默认运行选项
func NewDefaultOptions() Options {
	return Options{
		Global:   NewDefaultGlobalOptions(),
		Init:     NewDefaultInitOptions(),
		Clone:    NewDefaultCloneOptions(),
		Commit:   NewDefaultCommitOptions(),
		Checkout: NewDefaultCheckoutOptions(),
		Log:      NewDefaultLogOptions(),
	}
}

// Options stackcrisp 运行选项
type Options struct {
	// 全局选项
	Global GlobalOptions `json:"global,omitempty" yaml:"global,omitempty"`
	// init 命令选项
	Init InitOptions `json:"init,omitempty" yaml:"init,omitempty"`
	// clone 命令选项
	Clone CloneOptions `json:"clone,omitempty" yaml:"clone,omitempty"`
	// commit 命令选项
	Commit CommitOptions `json:"commit,omitempty" yaml:"commit,omitempty"`
	// checkout 命令选项
	Checkout CheckoutOptions `json:"checkout,omitempty" yaml:"checkout,omitempty"`
	// log 命令选项
	Log LogOptions `json:"log,omitempty" yaml:"log,omitempty"`
}
