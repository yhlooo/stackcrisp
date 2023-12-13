package mounts

// OverlayMountOptions OverlayFS 挂载选项
type OverlayMountOptions struct {
	// 挂载来源（挂载点名）
	// 默认为 overlay
	Source string
	// 挂载点路径
	MountPath string
	// OverlayFS lowerdir
	LowerDir []string
	// OverlayFS upperdir
	UpperDir string
	// OverlayFS workdir
	WorkDir string
	// 是否只读挂载
	ReadOnly bool
}
