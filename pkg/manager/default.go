package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-logr/logr"

	"github.com/yhlooo/stackcrisp/pkg/layers"
	"github.com/yhlooo/stackcrisp/pkg/mounts"
	"github.com/yhlooo/stackcrisp/pkg/spaces"
	fsutil "github.com/yhlooo/stackcrisp/pkg/utils/fs"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
	"github.com/yhlooo/stackcrisp/pkg/workspaces"
)

const (
	managerDataSubPathLayers = "overlay"
	managerDataSubPathSpaces = "spaces"
	managerDataSubPathMounts = "mounts"

	loggerName = "manager"
)

// New 创建一个 Manager
func New(opts Options) (Manager, error) {
	dataRoot, err := filepath.Abs(opts.DataRoot)
	if err != nil {
		return nil, fmt.Errorf("get absolute path of data root error: %w", err)
	}
	return &defaultManager{
		dataRoot: dataRoot,
		chownUID: opts.ChownUID,
		chownGID: opts.ChownGID,

		layerManager: nil,
	}, nil
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

// WorkspaceInfo 工作空间信息
type WorkspaceInfo struct {
	Path    string `json:"path"`
	Head    string `json:"head"`
	SpaceID string `json:"sapceID"`
	MountID string `json:"mountID"`
}

// CreateWorkspace 创建一个工作空间
func (mgr *defaultManager) CreateWorkspace(ctx context.Context, path string) (workspaces.Workspace, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("get absolute path of %q error: %w", path, err)
	}

	// 创建 space
	space, err := mgr.createSpace(ctx)
	if err != nil {
		return nil, fmt.Errorf("create space error: %w", err)
	}

	// 创建挂载
	mount, head, err := mgr.createMount(ctx, space, spaces.RootTag)
	if err != nil {
		return nil, fmt.Errorf("create mount error: %w", err)
	}

	ws := workspaces.New(absPath, head, space, mount)

	// 记录空间信息
	logger.Info(fmt.Sprintf("saving space %s ...", space.ID()))
	if err := space.Save(ctx); err != nil {
		return nil, fmt.Errorf("save space error: %w", err)
	}
	// 记录工作空间信息
	if err := mgr.saveWorkspaceInfo(ctx, ws); err != nil {
		return nil, fmt.Errorf("save workspace info error: %w", err)
	}

	return ws, nil
}

// GetWorkspaceFromPath 从指定目录获取对应工作空间
func (mgr *defaultManager) GetWorkspaceFromPath(ctx context.Context, path string) (workspaces.Workspace, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 从挂载路径获取挂载 ID
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("get absolute path of %q error: %w", path, err)
	}
	mountPath := absPath
	if fsutil.IsSymlink(absPath) {
		mountPath, err = os.Readlink(absPath)
		if err != nil {
			return nil, fmt.Errorf("get workspace mount path error: %w", err)
		}
	}
	absMountPath, err := filepath.Abs(mountPath)
	if err != nil {
		return nil, fmt.Errorf("get absolute path of %q error: %w", mountPath, err)
	}
	relPath, err := filepath.Rel(filepath.Join(mgr.dataRoot, managerDataSubPathMounts), absMountPath)
	if err != nil {
		return nil, fmt.Errorf("get relative path of mount path %q error: %w", absMountPath, err)
	}
	divided := strings.Split(relPath, string(filepath.Separator))
	if len(divided) == 0 {
		return nil, fmt.Errorf("parse mount path error")
	}
	mountID, err := uid.DecodeUID128FromBase32(divided[0])
	if err != nil {
		return nil, fmt.Errorf("parse mount id %q error: %w", divided[0], err)
	}

	// 读取挂载点对应工作空间信息
	wsInfo, err := mgr.loadWorkspaceInfo(ctx, mountID)
	if err != nil {
		return nil, fmt.Errorf("load workspace info error: %w", err)
	}

	// 加载 space
	logger.Info(fmt.Sprintf("loading space %q ...", wsInfo.SpaceID))
	spaceID, err := uid.DecodeUID128FromBase32(wsInfo.SpaceID)
	if err != nil {
		return nil, fmt.Errorf("parse space id %q error: %w", wsInfo.SpaceID, err)
	}
	space := spaces.New(
		spaceID,
		filepath.Join(mgr.dataRoot, managerDataSubPathSpaces, wsInfo.SpaceID),
		mgr.layerManager,
	)
	if err := space.Load(ctx); err != nil {
		return nil, fmt.Errorf("load space error: %w", err)
	}
	logger.Info(fmt.Sprintf("loaded space %s", space.ID()))

	// 加载挂载
	mount := mounts.NewMountedMount(mountID, mounts.MountOptions{
		MountDataRoot: filepath.Join(mgr.dataRoot, managerDataSubPathMounts, wsInfo.MountID),
		ChownUID:      mgr.chownUID,
		ChownGID:      mgr.chownGID,
	})
	logger.Info(fmt.Sprintf("loaded mount %s", mount.ID()))

	// 加载头指针
	head, err := uid.DecodeUID128FromHex(wsInfo.Head)
	if err != nil {
		return nil, fmt.Errorf("parse workspace head id %q error: %w", wsInfo.Head, err)
	}

	return workspaces.New(absPath, head, space, mount), nil
}

// RemoveWorkspaceMount 删除工作空间挂载
func (mgr *defaultManager) RemoveWorkspaceMount(ctx context.Context, ws workspaces.Workspace) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 先移动到根目录
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get pwd error: %w", err)
	}
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("change working directory to \"/\" error: %w", err)
	}

	// 卸载挂载
	if err := ws.Mount().Umount(ctx); err != nil {
		logger.Info(fmt.Sprintf("WARN umount %q error: %v", ws.Mount().MountPath(), err))
	}

	// 删除挂载数据
	mountDataPath := filepath.Join(mgr.dataRoot, managerDataSubPathMounts, ws.Mount().ID().Base32())
	logger.V(1).Info(fmt.Sprintf("rm %q", mountDataPath))
	if err := os.RemoveAll(mountDataPath); err != nil {
		return fmt.Errorf("remove mount data error: %w", err)
	}

	// 删除工作空间信息
	wsInfoFile := filepath.Join(mgr.dataRoot, managerDataSubPathMounts, ws.Mount().ID().Base32()+".workspace")
	logger.V(1).Info(fmt.Sprintf("rm %q", wsInfoFile))
	if err := os.Remove(wsInfoFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove workspace info error: %w", err)
	}

	// 移动回去
	if err := os.Chdir(pwd); err != nil {
		return fmt.Errorf("change working directory to %q error: %w", pwd, err)
	}

	return nil
}

// Commit 提交工作空间变更
func (mgr *defaultManager) Commit(ctx context.Context, ws workspaces.Workspace) (workspaces.Workspace, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 获取 space
	space := ws.Space()

	// 基于当前头指针创新新挂载
	mount, head, err := mgr.createMount(ctx, space, ws.Head().Hex())
	if err != nil {
		return nil, fmt.Errorf("create mount error: %w", err)
	}
	logger.Info(fmt.Sprintf("forward to new head %q", ws.Head().Hex()))

	newWS := workspaces.New(ws.Path(), head, space, mount)

	// 记录空间信息
	logger.Info(fmt.Sprintf("saving space %s ...", space.ID()))
	if err := space.Save(ctx); err != nil {
		return nil, fmt.Errorf("save space error: %w", err)
	}
	// 记录工作空间信息
	if err := mgr.saveWorkspaceInfo(ctx, newWS); err != nil {
		return nil, fmt.Errorf("save workspace info error: %w", err)
	}

	return newWS, nil
}

// Clone 克隆工作空间
func (mgr *defaultManager) Clone(
	ctx context.Context,
	sourceWS workspaces.Workspace,
	targetPath string,
) (workspaces.Workspace, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 获取 space
	space := sourceWS.Space()

	// 获取头节点
	headNode, ok := space.Tree().Get(sourceWS.Head())
	if !ok {
		return nil, fmt.Errorf("get workspace head layer %q not found", sourceWS.Head().Hex())
	}

	// 基于当前已经提交的最新层创建新挂载
	mount, head, err := mgr.createMount(ctx, space, headNode.Parent().ID().Hex())
	if err != nil {
		return nil, fmt.Errorf("create mount error: %w", err)
	}
	logger.Info(fmt.Sprintf("forward to new head %q", headNode.Parent().ID().Hex()))

	newWS := workspaces.New(targetPath, head, space, mount)

	// 记录空间信息
	logger.Info(fmt.Sprintf("saving space %s ...", space.ID()))
	if err := space.Save(ctx); err != nil {
		return nil, fmt.Errorf("save space error: %w", err)
	}
	// 记录工作空间信息
	if err := mgr.saveWorkspaceInfo(ctx, newWS); err != nil {
		return nil, fmt.Errorf("save workspace info error: %w", err)
	}

	return newWS, nil

}

// saveWorkspaceInfo 保存工作空间信息
func (mgr *defaultManager) saveWorkspaceInfo(ctx context.Context, ws workspaces.Workspace) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 序列化
	wsInfoRaw, err := json.Marshal(&WorkspaceInfo{
		Path:    ws.Path(),
		Head:    ws.Head().Hex(),
		SpaceID: ws.Space().ID().Base32(),
		MountID: ws.Mount().ID().Base32(),
	})
	if err != nil {
		return fmt.Errorf("marshal workspace info to json error: %w", err)
	}
	wsInfoFile := filepath.Join(
		mgr.dataRoot, managerDataSubPathMounts,
		ws.Mount().ID().Base32()+".workspace",
	)

	// 写文件
	logger.V(1).Info(fmt.Sprintf("write workspace info to file %q", wsInfoFile))
	if err := os.WriteFile(wsInfoFile, wsInfoRaw, 0644); err != nil {
		return fmt.Errorf("write workspace info to file error: %w", err)
	}

	return nil
}

// loadWorkspaceInfo 加载工作空间信息
func (mgr *defaultManager) loadWorkspaceInfo(ctx context.Context, mountID uid.UID) (*WorkspaceInfo, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	wsInfoFile := filepath.Join(mgr.dataRoot, managerDataSubPathMounts, mountID.Base32()+".workspace")
	logger.V(1).Info(fmt.Sprintf("write workspace info to file %q", wsInfoFile))

	// 读文件
	wsInfoRaw, err := os.ReadFile(wsInfoFile)
	if err != nil {
		return nil, fmt.Errorf("read workspace info from file %q error: %w", wsInfoFile, err)
	}

	// 反序列化
	var wsInfo WorkspaceInfo
	if err := json.Unmarshal(wsInfoRaw, &wsInfo); err != nil {
		return nil, fmt.Errorf("unmarshal workspace info from json error: %w", err)
	}

	return &wsInfo, nil
}

// createSpace 创建一个存储空间
func (mgr *defaultManager) createSpace(ctx context.Context) (spaces.Space, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	spaceID := uid.NewUID128()
	logger.Info(fmt.Sprintf("creating space %s ...", spaceID))
	spaceDataRoot := filepath.Join(mgr.dataRoot, managerDataSubPathSpaces, spaceID.Base32())
	logger.V(1).Info(fmt.Sprintf("madir %q", spaceDataRoot))
	if err := os.Mkdir(spaceDataRoot, 0755); err != nil {
		return nil, fmt.Errorf("make directory %q for space data root error: %w", spaceDataRoot, err)
	}
	space := spaces.New(spaceID, spaceDataRoot, mgr.layerManager)
	logger.V(1).Info("initializing space ...")
	if err := space.Init(ctx); err != nil {
		return space, fmt.Errorf("init space error: %w", err)
	}
	return space, nil
}

// createMount 使用指定空间版本创建一个挂载
func (mgr *defaultManager) createMount(
	ctx context.Context,
	space spaces.Space,
	revision string,
) (mounts.Mount, uid.UID, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 创建挂载目录
	mountID := uid.NewUID128()
	logger.Info(fmt.Sprintf("creating mount %s ...", mountID))
	mountDataRoot := filepath.Join(mgr.dataRoot, managerDataSubPathMounts, mountID.Base32())
	logger.V(1).Info(fmt.Sprintf("madir %q", mountDataRoot))
	if err := os.Mkdir(mountDataRoot, 0755); err != nil {
		return nil, nil, fmt.Errorf("make directory %q for mount data root error: %w", mountDataRoot, err)
	}
	// 挂载
	mount, head, err := space.CreateMount(ctx, revision, mountID, mounts.MountOptions{
		MountDataRoot: mountDataRoot,
		ChownUID:      mgr.chownUID,
		ChownGID:      mgr.chownGID,
	})
	return mount, head, err
}
