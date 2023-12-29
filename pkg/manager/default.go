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

	loggerName = "manager"
)

// New 创建一个 Manager
func New(opts Options) Manager {
	return &defaultManager{
		dataRoot: opts.DataRoot,
		chownUID: opts.ChownUID,
		chownGID: opts.ChownGID,

		layerManager: nil,
	}
}

// defaultManager 是 Manager 的一个默认实现
type defaultManager struct {
	dataRoot string
	chownUID int
	chownGID int

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
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 确保根目录
	logger.V(1).Info("preparing data root ...")
	if !fsutil.IsDir(mgr.dataRoot) {
		logger.V(1).Info(fmt.Sprintf("madir %q", mgr.dataRoot))
		if err := os.MkdirAll(mgr.dataRoot, 0755); err != nil {
			return fmt.Errorf("make directory for data root error: %w", err)
		}
	}
	// 确保层管理器
	logger.V(1).Info("preparing layer manager ...")
	layersDataRoot := filepath.Join(mgr.dataRoot, managerDataSubPathLayers)
	if !fsutil.IsDir(layersDataRoot) {
		logger.V(1).Info(fmt.Sprintf("madir %q", layersDataRoot))
		if err := os.Mkdir(layersDataRoot, 0755); err != nil {
			return fmt.Errorf("make directory for layers data root error: %w", err)
		}
	}
	if mgr.layerManager == nil {
		logger.V(1).Info(fmt.Sprintf("create layer manager, dataRoot: %q", layersDataRoot))
		mgr.layerManager = layers.NewLayerManager(layersDataRoot)
	}
	// 确保 spaces 目录
	logger.V(1).Info("preparing spaces data root")
	spacesDataRoot := filepath.Join(mgr.dataRoot, managerDataSubPathSpaces)
	if !fsutil.IsDir(spacesDataRoot) {
		logger.V(1).Info(fmt.Sprintf("madir %q", spacesDataRoot))
		if err := os.Mkdir(spacesDataRoot, 0755); err != nil {
			return fmt.Errorf("make directory for spaces data root error: %w", err)
		}
	}
	// 确保 mounts 目录
	logger.V(1).Info("preparing mounts data root")
	mountsDataRoot := filepath.Join(mgr.dataRoot, managerDataSubPathMounts)
	if !fsutil.IsDir(mountsDataRoot) {
		logger.V(1).Info(fmt.Sprintf("madir %q", mountsDataRoot))
		if err := os.Mkdir(mountsDataRoot, 0755); err != nil {
			return fmt.Errorf("make directory for mounts data root error: %w", err)
		}
	}

	return nil
}

// CreateSpace 创建一个存储空间
func (mgr *defaultManager) CreateSpace(ctx context.Context) (spaces.Space, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	spaceID := uid.NewUID128()
	logger.Info(fmt.Sprintf("creating space %s ...", spaceID))

	spaceDataRoot := filepath.Join(mgr.dataRoot, managerDataSubPathSpaces, spaceID.Base32())
	logger.V(1).Info(fmt.Sprintf("madir %q", spaceDataRoot))
	if err := os.Mkdir(spaceDataRoot, 0755); err != nil {
		return nil, fmt.Errorf("make directory %q for space data root error: %w", spaceDataRoot, err)
	}
	space := spaces.New(spaceDataRoot, mgr.layerManager)
	logger.V(1).Info("initializing space ...")
	if err := space.Init(ctx); err != nil {
		return space, fmt.Errorf("init space error: %w", err)
	}
	logger.V(1).Info("saving space ...")
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
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 创建挂载目录
	mountDataRoot := filepath.Join(mgr.dataRoot, managerDataSubPathMounts, uid.NewUID128().Base32())
	logger.V(1).Info(fmt.Sprintf("madir %q", mountDataRoot))
	if err := os.Mkdir(mountDataRoot, 0755); err != nil {
		return nil, fmt.Errorf("make directory %q for mount data root error: %w", mountDataRoot, err)
	}
	// 挂载
	mount, err := space.CreateMount(ctx, revision, mounts.MountOptions{
		MountDataRoot: mountDataRoot,
		ChownUID:      mgr.chownUID,
		ChownGID:      mgr.chownGID,
	})
	if err != nil {
		return mount, err
	}
	return mount, mount.Mount(ctx)
}
