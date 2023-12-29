package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-logr/logr"

	"github.com/yhlooo/stackcrisp/pkg/layers"
	"github.com/yhlooo/stackcrisp/pkg/mounts"
	"github.com/yhlooo/stackcrisp/pkg/spaces"
	fsutil "github.com/yhlooo/stackcrisp/pkg/utils/fs"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

const (
	managerDataSubPathLayers = "overlay"
	managerDataSubPathSpaces = "spaces"
	managerDataSubPathMounts = "mounts"
)

// New 创建一个 Manager
func New(opts Options) Manager {
	return &defaultManager{
		dataRoot:     opts.DataRoot,
		layerManager: nil,
	}
}

// defaultManager 是 Manager 的一个默认实现
type defaultManager struct {
	dataRoot string

	prepareOnce  sync.Once
	layerManager layers.LayerManager
}

var _ Manager = &defaultManager{}

// Prepare 准备
func (mgr *defaultManager) Prepare(ctx context.Context) error {
	var err error
	mgr.prepareOnce.Do(func() {
		err = mgr.doPrepare(ctx)
	})
	return err
}

// doPrepare 准备
func (mgr *defaultManager) doPrepare(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)

	// 确保根目录
	logger.V(1).Info("preparing data root")
	if err := os.MkdirAll(mgr.dataRoot, 0755); err != nil {
		return fmt.Errorf("make directory for data root error: %w", err)
	}
	// 确保层管理器
	logger.V(1).Info("preparing layer manager")
	layersDataRoot := filepath.Join(mgr.dataRoot, managerDataSubPathLayers)
	if !fsutil.IsDir(layersDataRoot) {
		if err := os.Mkdir(layersDataRoot, 0755); err != nil {
			return fmt.Errorf("make directory for layers data root error: %w", err)
		}
	}
	if mgr.layerManager == nil {
		mgr.layerManager = layers.NewLayerManager(layersDataRoot)
	}
	// 确保 spaces 目录
	logger.V(1).Info("preparing spaces data root")
	spacesDataRoot := filepath.Join(mgr.dataRoot, managerDataSubPathSpaces)
	if !fsutil.IsDir(spacesDataRoot) {
		if err := os.Mkdir(spacesDataRoot, 0755); err != nil {
			return fmt.Errorf("make directory for spaces data root error: %w", err)
		}
	}
	// 确保 mounts 目录
	logger.V(1).Info("preparing mounts data root")
	mountsDataRoot := filepath.Join(mgr.dataRoot, managerDataSubPathMounts)
	if !fsutil.IsDir(mountsDataRoot) {
		if err := os.Mkdir(mountsDataRoot, 0755); err != nil {
			return fmt.Errorf("make directory for mounts data root error: %w", err)
		}
	}

	return nil
}

// CreateSpace 创建一个存储空间
func (mgr *defaultManager) CreateSpace(ctx context.Context) (spaces.Space, error) {
	spaceDataRoot := filepath.Join(mgr.dataRoot, managerDataSubPathSpaces, uid.NewUID128().Base32())
	if err := os.Mkdir(spaceDataRoot, 0755); err != nil {
		return nil, fmt.Errorf("make directory %q for space data root error: %w", spaceDataRoot, err)
	}
	space := spaces.New(spaceDataRoot, mgr.layerManager)
	if err := space.Init(ctx); err != nil {
		return space, fmt.Errorf("init space error: %w", err)
	}
	if err := space.Save(ctx); err != nil {
		return space, fmt.Errorf("save space error: %w", err)
	}
	return space, nil
}

// CreateMount 使用指定空间版本创建一个挂载
func (mgr *defaultManager) CreateMount(
	ctx context.Context,
	space spaces.Space,
	revision string,
) (mounts.Mount, error) {
	// 创建挂载目录
	mountDataRoot := filepath.Join(mgr.dataRoot, managerDataSubPathMounts, uid.NewUID128().Base32())
	if err := os.Mkdir(mountDataRoot, 0755); err != nil {
		return nil, fmt.Errorf("make directory %q for mount data root error: %w", mountDataRoot, err)
	}
	// 挂载
	return space.CreateMount(ctx, revision, mounts.MountOptions{
		MountDataRoot: mountDataRoot,
	})
}
