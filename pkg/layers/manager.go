package layers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"

	"github.com/yhlooo/stackcrisp/pkg/utils/errors"
	fsutil "github.com/yhlooo/stackcrisp/pkg/utils/fs"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

// NewLayerManager 创建一个 LayerManager
func NewLayerManager(layersDataRoot string) LayerManager {
	return &defaultLayerManager{
		layersDataRoot: layersDataRoot,
	}
}

// LayerManager 层管理器
type LayerManager interface {
	// Create 创建一个层
	Create(ctx context.Context) (Layer, error)
	// Get 通过层 ID 获取层
	Get(ctx context.Context, id uid.UID) (Layer, error)
	// List 列出所有层
	List(ctx context.Context) ([]Layer, error)
	// Delete 通过层 ID 删除指定层
	Delete(ctx context.Context, id uid.UID) (Layer, error)
}

// defaultLayerManager 是 LayerManager 的默认实现
type defaultLayerManager struct {
	// 层数据存储根目录
	layersDataRoot string
}

var _ LayerManager = &defaultLayerManager{}

// Create 创建一个层
func (mgr *defaultLayerManager) Create(_ context.Context) (Layer, error) {
	// 创建层目录
	id := uid.NewUID128()
	rootDir := mgr.getLayerDataRoot(id)
	if err := os.Mkdir(rootDir, 0755); err != nil {
		return nil, fmt.Errorf("make dir %q for layer data root error: %w", rootDir, err)
	}
	// 创建层
	l := NewLayer(id, rootDir)
	if err := l.Save(); err != nil {
		return nil, err
	}
	return l, nil
}

// Get 通过层 ID 获取层
func (mgr *defaultLayerManager) Get(_ context.Context, id uid.UID) (Layer, error) {
	rootDir := mgr.getLayerDataRoot(id)
	if !fsutil.IsDir(rootDir) {
		return nil, errors.New(ErrReasonLayerNotFound, fmt.Sprintf("layer %q not found", id))
	}
	return &defaultLayer{
		id:            id,
		layerDataRoot: rootDir,
	}, nil
}

// List 列出所有层
func (mgr *defaultLayerManager) List(ctx context.Context) ([]Layer, error) {
	logger := logr.FromContextOrDiscard(ctx)

	// 列出目录
	entries, err := os.ReadDir(mgr.layersDataRoot)
	if err != nil {
		return nil, fmt.Errorf("read layers data root dir %q error: %w", mgr.layersDataRoot, err)
	}

	// 从目录名获取层 id
	layers := make([]Layer, 0, len(entries))
	for _, item := range entries {
		if !item.IsDir() {
			logger.Info(fmt.Sprintf(
				"WARN: unexpected file in layers data root: %q",
				filepath.Join(mgr.layersDataRoot, item.Name()),
			))
			continue
		}
		id, err := uid.DecodeUID128FromBase32(item.Name())
		if err != nil {
			logger.Error(err, fmt.Sprintf("parse uid from layer data dirname %q error", item.Name()))
			continue
		}
		layers = append(layers, &defaultLayer{id: id, layerDataRoot: mgr.getLayerDataRoot(id)})
	}

	return layers, nil
}

// Delete 通过层 ID 删除指定层
func (mgr *defaultLayerManager) Delete(_ context.Context, id uid.UID) (Layer, error) {
	rootDir := mgr.getLayerDataRoot(id)
	if !fsutil.IsDir(rootDir) {
		return nil, errors.New(ErrReasonLayerNotFound, fmt.Sprintf("layer %q not found", id))
	}
	if err := os.RemoveAll(rootDir); err != nil {
		return nil, fmt.Errorf("remove layer data %q error: %w", rootDir, err)
	}
	return &defaultLayer{
		id:            id,
		layerDataRoot: rootDir,
	}, nil
}

// getLayerDataRoot 获取层数据目录
func (mgr *defaultLayerManager) getLayerDataRoot(id uid.UID) string {
	return filepath.Join(mgr.layersDataRoot, id.Base32())
}
