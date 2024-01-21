package workspaces

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/go-logr/logr"

	"github.com/yhlooo/stackcrisp/pkg/mounts"
	"github.com/yhlooo/stackcrisp/pkg/spaces"
	"github.com/yhlooo/stackcrisp/pkg/spaces/trees"
	fsutil "github.com/yhlooo/stackcrisp/pkg/utils/fs"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

const (
	loggerName = "workspaces"

	rootTag = "ROOT"
	headTag = "HEAD"
)

// New 创建一个工作空间
func New(id uid.UID, path string, space spaces.Space, mount mounts.Mount, head trees.Node, branch string) Workspace {
	return &defaultWorkspace{
		id:     id,
		path:   path,
		space:  space,
		mount:  mount,
		head:   head,
		branch: branch,
	}
}

// defaultWorkspace 是 Workspace 的一个默认实现
type defaultWorkspace struct {
	id     uid.UID
	path   string
	space  spaces.Space
	mount  mounts.Mount
	head   trees.Node
	branch string
}

var _ Workspace = &defaultWorkspace{}

// ID 返回 Workspace ID
func (ws *defaultWorkspace) ID() uid.UID {
	return ws.id
}

// Path 返回工作空间路径
func (ws *defaultWorkspace) Path() string {
	return ws.path
}

// Space 返回工作空间对应空间
func (ws *defaultWorkspace) Space() spaces.Space {
	return ws.space
}

// Mount 返回工作空间对应挂载
func (ws *defaultWorkspace) Mount() mounts.Mount {
	return ws.mount
}

// Head 返回头指针
func (ws *defaultWorkspace) Head() trees.Node {
	return ws.head
}

// Branch 返回当前分支
func (ws *defaultWorkspace) Branch() BranchInfo {
	return NewLocalBranch(ws.id, ws.branch)
}

// Expand 展开工作空间
func (ws *defaultWorkspace) Expand(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 挂载
	logger.Info("mounting ...")
	if err := ws.Mount().Mount(ctx); err != nil {
		return fmt.Errorf("mount error: %w", err)
	}

	// 先移动到根目录
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get pwd error: %w", err)
	}
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("change working directory to \"/\" error: %w", err)
	}

	// 确保目标路径上什么也没有
	if fsutil.IsExists(ws.Path()) {
		logger.V(1).Info(fmt.Sprintf("target path %q exists, remove it", ws.Path()))
		if err := os.Remove(ws.Path()); err != nil {
			return fmt.Errorf("clear path %q error: %w", ws.Path(), err)
		}
	}

	// 创建软链
	logger.Info(fmt.Sprintf("link mount point to %q", ws.Path()))
	if err := ws.Mount().CreateSymlink(ctx, ws.Path()); err != nil {
		return fmt.Errorf("create syslink %q to mount path error: %w", ws.Path(), err)
	}

	// 移动回去
	if err := os.Chdir(pwd); err != nil {
		return fmt.Errorf("change working directory to %q error: %w", pwd, err)
	}

	return nil
}

// GetHistory 获取提交历史
func (ws *defaultWorkspace) GetHistory(ref string) ([]Commit, error) {
	// 获取指定节点
	node, _, _ := ws.Search(ref)
	if node == nil {
		return nil, fmt.Errorf("revision %q not found", ref)
	}

	// 追溯提交历史
	var commits []Commit
	cur := node
	for !cur.IsRoot() {
		commits = append(commits, GetCommitFromNode(ws, cur))
		cur = cur.Parent()
	}

	return commits, nil
}

// AllBranches 返回本地分支和全局分支列表
func (ws *defaultWorkspace) AllBranches() []Branch {
	var ret []Branch
	for fullName, head := range ws.Space().Tree().Branches() {
		b, err := ParseBranchFullName(fullName)
		if err != nil {
			continue
		}
		if b.IsLocal() && b.WorkspaceID().Base32() != ws.id.Base32() {
			// 其它 workspace 的本地分支
			continue
		}
		ret = append(ret, b.Branch(head))
	}
	// 排序
	sort.Slice(ret, func(i, j int) bool {
		switch {
		case ret[i].IsLocal() && ret[j].IsGlobal():
			// 本地分支优先
			return true
		case ret[i].IsGlobal() && ret[j].IsLocal():
			// 本地分支优先
			return false
		default:
			// 按字典序
			return ret[i].Name() < ret[j].Name()
		}
	})
	return ret
}

// LocalBranches 返回本地分支列表
func (ws *defaultWorkspace) LocalBranches() []Branch {
	var ret []Branch
	for fullName, head := range ws.Space().Tree().Branches() {
		b, err := ParseBranchFullName(fullName)
		if err != nil {
			continue
		}
		if !b.IsLocal() {
			// 全局分支
			continue
		}
		if b.WorkspaceID().Base32() != ws.id.Base32() {
			// 其它 workspace 的本地分支
			continue
		}
		ret = append(ret, b.Branch(head))
	}
	// 排序
	sort.Slice(ret, func(i, j int) bool {
		// 按字典序
		return ret[i].Name() < ret[j].Name()
	})
	return ret
}

// RemoteBranches 返回远程分支列表
func (ws *defaultWorkspace) RemoteBranches() []Branch {
	var ret []Branch
	for fullName, head := range ws.Space().Tree().Branches() {
		b, err := ParseBranchFullName(fullName)
		if err != nil {
			continue
		}
		if !b.IsGlobal() {
			// 本地分支
			continue
		}
		ret = append(ret, b.Branch(head))
	}
	// 排序
	sort.Slice(ret, func(i, j int) bool {
		// 按字典序
		return ret[i].Name() < ret[j].Name()
	})
	return ret
}

// SetBranch 设置当前分支
func (ws *defaultWorkspace) SetBranch(localName string) error {
	if err := ws.space.Tree().AddBranch(
		NewLocalBranch(ws.id, localName).FullName(),
		ws.head.Parent().ID(),
	); err != nil {
		return err
	}
	ws.branch = localName
	return nil
}

// AddBranch 添加分支
func (ws *defaultWorkspace) AddBranch(ctx context.Context, branchLocalName string, ref string, force bool) error {
	node, _, ok := ws.Search(ref)
	if !ok {
		return fmt.Errorf("failed to resolve %q as valid ref", ref)
	}

	branch := NewLocalBranch(ws.id, branchLocalName)

	// 检查是否已经存在该分支
	if existsNode, ok := ws.Space().Tree().GetByBranch(branch.FullName()); ok && !force {
		return fmt.Errorf("branch %q already exists at %q", branch.LocalName(), existsNode.ID().Hex())
	}

	// 添加分支
	if err := ws.Space().Tree().AddBranch(branch.FullName(), node.ID()); err != nil {
		return err
	}

	// 保存
	if err := ws.Space().Save(ctx); err != nil {
		return fmt.Errorf("save space info error: %w", err)
	}

	return nil
}

// Tags 返回标签列表
func (ws *defaultWorkspace) Tags() []string {
	tagsMap := ws.Space().Tree().Tags()
	if len(tagsMap) == 0 {
		return nil
	}
	ret := make([]string, 0, len(tagsMap))
	for t := range tagsMap {
		if t == rootTag {
			continue
		}
		ret = append(ret, t)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i] < ret[j]
	})
	return ret
}

// DeleteTag 删除标签
func (ws *defaultWorkspace) DeleteTag(ctx context.Context, tagName string) error {
	// 删除标签
	if ok := ws.Space().Tree().DeleteTag(tagName); !ok {
		return fmt.Errorf("tag %q not found", tagName)
	}

	// 保存
	if err := ws.Space().Save(ctx); err != nil {
		return fmt.Errorf("save space info error: %w", err)
	}
	return nil
}

// AddTag 添加标签
func (ws *defaultWorkspace) AddTag(ctx context.Context, tagName string, ref string, force bool) error {
	node, _, ok := ws.Search(ref)
	if !ok {
		return fmt.Errorf("failed to resolve %q as valid ref", ref)
	}

	// 检查是否已经存在该标签
	if existsNode, ok := ws.Space().Tree().GetByTag(tagName); ok && !force {
		return fmt.Errorf("tag %q already exists at %q", tagName, existsNode.ID().Hex())
	}

	// 添加标签
	if err := ws.Space().Tree().AddTag(tagName, node.ID()); err != nil {
		return err
	}

	// 保存
	if err := ws.Space().Save(ctx); err != nil {
		return fmt.Errorf("save space info error: %w", err)
	}

	return nil
}

// Search 通过 ref 搜索节点
//
// ref 可以是各种形式的节点 ID 、分支名、标签名
func (ws *defaultWorkspace) Search(ref string) (trees.Node, trees.KeyType, bool) {
	// 首先是 HEAD
	if ref == headTag {
		return ws.Head().Parent(), trees.Commit, true
	}
	// 首先直接搜
	// 包括 commit 、 tag 、 分支完整名
	if node, keyType, ok := ws.space.Tree().Search(ref); ok {
		return node, keyType, true
	}
	// 然后搜索分支本地名
	for _, b := range ParseBranchLocalName(ws.id, ref) {
		if node, ok := ws.space.Tree().GetByBranch(b.FullName()); ok {
			return node, trees.Branch, true
		}
	}
	// 实在没有了
	return nil, "", false
}
