package layers

import (
	"fmt"
	"os"
	"path/filepath"

	fsutil "github.com/yhlooo/stackcrisp/pkg/utils/fs"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

// NewLayer 创建一个 Layer
func NewLayer(id uid.UID, layerDataRoot string) Layer {
	return &defaultLayer{
		id:            id,
		layerDataRoot: layerDataRoot,
	}
}

// Layer 层
type Layer interface {
	// ID 返回层 ID
	ID() uid.UID
	// DiffDir 返回 diff 目录路径
	DiffDir() string
	// Save 将相关信息持久化
	Save() error
}

const layerDataSubPathDiff = "diff"

// defaultLayer 是 Layer 的一个默认实现
type defaultLayer struct {
	id            uid.UID
	layerDataRoot string
}

var _ Layer = &defaultLayer{}

// ID 返回层 ID
func (l *defaultLayer) ID() uid.UID {
	return l.id
}

// DiffDir 返回 diff 目录路径
func (l *defaultLayer) DiffDir() string {
	return filepath.Join(l.layerDataRoot, layerDataSubPathDiff)
}

// Save 将相关信息持久化
func (l *defaultLayer) Save() error {
	// 创建 diff 目录
	if !fsutil.IsDir(l.DiffDir()) {
		if err := os.Mkdir(l.DiffDir(), 0755); err != nil {
			return fmt.Errorf("make dir %q for layer diff error: %w", l.DiffDir(), err)
		}
	}
	return nil
}
