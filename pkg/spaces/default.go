package spaces

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/yhlooo/stackcrisp/pkg/layers"
	"github.com/yhlooo/stackcrisp/pkg/mounts"
	"github.com/yhlooo/stackcrisp/pkg/spaces/trees"
)

const (
	spaceDataSubPathTree = "tree.json"

	// RootTag 根节点标签
	RootTag = "ROOT"
)

// New 创建一个 Space
func New(spaceDataRoot string, layerManager layers.LayerManager) Space {
	return &defaultSpace{
		spaceDataRoot: spaceDataRoot,
		layerManger:   layerManager,
	}
}

// defaultSpace 是 Space 的一个默认实现
type defaultSpace struct {
	spaceDataRoot string

	layerTree   trees.Tree
	layerManger layers.LayerManager
}

var _ Space = &defaultSpace{}

// Init 初始化
func (space *defaultSpace) Init(ctx context.Context) error {
	// 创建树
	space.layerTree = trees.NewTree()
	// 创建根节点
	rootLayer, err := space.layerManger.Create(ctx)
	if err != nil {
		return fmt.Errorf("create root layer error: %w", err)
	}
	if err := space.layerTree.AddNode(nil, trees.NewNode(rootLayer.ID())); err != nil {
		return fmt.Errorf("add root layer to tree error: %w", err)
	}
	if err := space.layerTree.AddTag(RootTag, rootLayer.ID()); err != nil {
		return fmt.Errorf("add root tag error: %w", err)
	}
	return nil
}

// Load 加载数据
func (space *defaultSpace) Load(context.Context) error {
	// 读取树
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
func (space *defaultSpace) Save(context.Context) error {
	// 导出树
	dump := trees.Dump(space.layerTree)
	// json 序列化
	raw, err := json.Marshal(&dump)
	if err != nil {
		return fmt.Errorf("marshal tree dump to json error: %w", err)
	}
	// 写文件
	if err := os.WriteFile(space.treeDumpSavePath(), raw, 0644); err != nil {
		return fmt.Errorf("write tree dump error: %w", err)
	}

	return nil
}

// CreateMount 创建一个该空间的挂载
func (space *defaultSpace) CreateMount(
	ctx context.Context,
	revision string,
	mountOpts mounts.MountOptions,
) (mounts.Mount, error) {
	// 找到 lower 的最上层节点
	lowerNode, ok := space.layerTree.Search(revision)
	if !ok {
		return nil, fmt.Errorf("layer %q not found", revision)
	}
	// 创建一个作为 upper 层的节点
	upper, err := space.layerManger.Create(ctx)
	if err != nil {
		return nil, fmt.Errorf("create upper layer error: %w", err)
	}
	if err := space.layerTree.AddNode(lowerNode.ID(), trees.NewNode(upper.ID())); err != nil {
		return nil, fmt.Errorf("add upper layer to tree error: %w", err)
	}
	// 找到所有 lower 层
	var layerSet []layers.Layer
	cur := lowerNode
	for cur != nil {
		layer, err := space.layerManger.Get(ctx, cur.ID())
		if err != nil {
			return nil, fmt.Errorf("get layer %q error: %w", cur.ID(), err)
		}
		layerSet = append(layerSet, layer)
		cur = cur.Parent()
	}
	slices.Reverse(layerSet)
	// 加上 upper 层
	layerSet = append(layerSet, upper)

	return mounts.New(ctx, layerSet, mountOpts)
}

// treeDumpSavePath 返回导出的树存储路径
func (space *defaultSpace) treeDumpSavePath() string {
	return filepath.Join(space.spaceDataRoot, spaceDataSubPathTree)
}
