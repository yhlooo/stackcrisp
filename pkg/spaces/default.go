package spaces

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/go-logr/logr"

	"github.com/yhlooo/stackcrisp/pkg/layers"
	"github.com/yhlooo/stackcrisp/pkg/mounts"
	"github.com/yhlooo/stackcrisp/pkg/spaces/trees"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

const (
	spaceDataSubPathTree = "tree.json"
	loggerName           = "spaces"

	// RootTag 根节点标签
	RootTag = "ROOT"
)

// New 创建一个 Space
func New(id uid.UID, spaceDataRoot string, layerManager layers.LayerManager) Space {
	return &defaultSpace{
		id:            id,
		spaceDataRoot: spaceDataRoot,
		layerManger:   layerManager,
	}
}

// defaultSpace 是 Space 的一个默认实现
type defaultSpace struct {
	id            uid.UID
	spaceDataRoot string

	layerTree   trees.Tree
	layerManger layers.LayerManager
}

var _ Space = &defaultSpace{}

// ID 返回空间 ID
func (space *defaultSpace) ID() uid.UID {
	return space.id
}

// Init 初始化
func (space *defaultSpace) Init(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 创建树
	space.layerTree = trees.NewTree()
	// 创建根节点
	logger.Info("creating root layer ...")
	rootLayer, err := space.layerManger.Create(ctx)
	if err != nil {
		return fmt.Errorf("create root layer error: %w", err)
	}
	logger.V(1).Info(fmt.Sprintf("add root layer %s to tree", rootLayer.ID()))
	if err := space.layerTree.AddNode(nil, trees.NewNode(rootLayer.ID())); err != nil {
		return fmt.Errorf("add root layer to tree error: %w", err)
	}
	logger.V(1).Info("add root layer tag")
	if err := space.layerTree.AddTag(RootTag, rootLayer.ID()); err != nil {
		return fmt.Errorf("add root tag error: %w", err)
	}
	return nil
}

// Load 加载数据
func (space *defaultSpace) Load(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 读取树
	logger.V(1).Info(fmt.Sprintf("reading tree dump from %q ...", space.treeDumpSavePath()))
	raw, err := os.ReadFile(space.treeDumpSavePath())
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read tree dump error: %w", err)
	}
	// json 反序列化
	var dump trees.TreeDump
	if err := json.Unmarshal(raw, &dump); err != nil {
		return fmt.Errorf("unmarshal tree dump from json error: %w", err)
	}
	// 加载
	space.layerTree, err = trees.Load(dump)
	if err != nil {
		return fmt.Errorf("load tree from dump error: %w", err)
	}

	return nil
}

// Save 将数据持久化
func (space *defaultSpace) Save(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 导出树
	logger.V(1).Info("dumping tree ...")
	dump := trees.Dump(space.layerTree)
	// json 序列化
	raw, err := json.Marshal(&dump)
	if err != nil {
		return fmt.Errorf("marshal tree dump to json error: %w", err)
	}
	// 写文件
	logger.V(1).Info(fmt.Sprintf("writing tree dump to %q ...", space.treeDumpSavePath()))
	if err := os.WriteFile(space.treeDumpSavePath(), raw, 0644); err != nil {
		return fmt.Errorf("write tree dump error: %w", err)
	}

	return nil
}

// CreateMount 创建一个该空间的挂载
func (space *defaultSpace) CreateMount(
	ctx context.Context,
	revision string,
	mountID uid.UID,
	mountOpts mounts.MountOptions,
) (mounts.Mount, uid.UID, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 找到 lower 的最上层节点
	lowerNode, ok := space.layerTree.Search(revision)
	if !ok {
		return nil, nil, fmt.Errorf("layer %q not found", revision)
	}
	// 创建一个作为 upper 层的节点
	logger.Info("creating upper layer ...")
	upper, err := space.layerManger.Create(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("create upper layer error: %w", err)
	}
	logger.V(1).Info(fmt.Sprintf("add upper layer %s to tree", upper.ID()))
	if err := space.layerTree.AddNode(lowerNode.ID(), trees.NewNode(upper.ID())); err != nil {
		return nil, nil, fmt.Errorf("add upper layer to tree error: %w", err)
	}
	// 找到所有 lower 层
	var layerSet []layers.Layer
	cur := lowerNode
	for cur != nil {
		layer, err := space.layerManger.Get(ctx, cur.ID())
		if err != nil {
			return nil, nil, fmt.Errorf("get layer %q error: %w", cur.ID(), err)
		}
		layerSet = append(layerSet, layer)
		cur = cur.Parent()
	}
	slices.Reverse(layerSet)
	// 加上 upper 层
	layerSet = append(layerSet, upper)

	logger.V(1).Info(fmt.Sprintf("mount layers: %v", layerSet))
	mount, err := mounts.New(ctx, mountID, layerSet, mountOpts)
	return mount, upper.ID(), err
}

// treeDumpSavePath 返回导出的树存储路径
func (space *defaultSpace) treeDumpSavePath() string {
	return filepath.Join(space.spaceDataRoot, spaceDataSubPathTree)
}
