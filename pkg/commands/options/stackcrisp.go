package options

// NewDefaultOptions 创建一个默认运行选项
func NewDefaultOptions() Options {
	return Options{
		Global: NewDefaultGlobalOptions(),
		Init:   NewDefaultInitOptions(),
		Commit: NewDefaultCommitOptions(),
	}
}

// Options stackcrisp 运行选项
type Options struct {
	// 全局选项
	Global GlobalOptions `json:"global,omitempty" yaml:"global,omitempty"`
	// init 命令选项
	Init InitOptions `json:"init,omitempty" yaml:"init,omitempty"`
	// commit 命令选项
	Commit CommitOptions `json:"commit,omitempty" yaml:"commit,omitempty"`
}
