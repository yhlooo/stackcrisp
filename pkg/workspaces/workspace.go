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
func (ws *defaultWorkspace) Branch() Branch {
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

// AllBranches 返回本地分支和全局分支列表
func (ws *defaultWorkspace) AllBranches() []Branch {
	var ret []Branch
	for fullName := range ws.Space().Tree().Branches() {
		b, err := ParseBranchFullName(fullName)
		if err != nil {
			continue
		}
		if b.IsLocal() && b.WorkspaceID().Base32() != ws.id.Base32() {
			// 其它 workspace 的本地分支
			continue
		}
		ret = append(ret, b)
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
	for fullName := range ws.Space().Tree().Branches() {
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
		ret = append(ret, b)
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
	for fullName := range ws.Space().Tree().Branches() {
		b, err := ParseBranchFullName(fullName)
		if err != nil {
			continue
		}
		if !b.IsGlobal() {
			// 本地分支
			continue
		}
		ret = append(ret, b)
	}
	// 排序
	sort.Slice(ret, func(i, j int) bool {
		// 按字典序
		return ret[i].Name() < ret[j].Name()
	})
	return ret
}

// SetBranch 设置分支
func (ws *defaultWorkspace) SetBranch(localName string) error {
	if err := ws.space.Tree().AddBranch(NewLocalBranch(ws.id, localName).FullName(), ws.head.ID()); err != nil {
		return err
	}
	ws.branch = localName
	return nil
}

// Search 通过 key 搜索节点
//
// key 可以是各种形式的节点 ID 、分支名、标签名
func (ws *defaultWorkspace) Search(key string) (trees.Node, trees.KeyType, bool) {
	// 首先是 HEAD
	if key == headTag {
		return ws.Head().Parent(), trees.Commit, true
	}
	// 首先直接搜
	// 包括 commit 、 tag 、 分支完整名
	if node, keyType, ok := ws.space.Tree().Search(key); ok {
		return node, keyType, true
	}
	// 然后搜索分支本地名
	for _, b := range ParseBranchLocalName(ws.id, key) {
		if node, ok := ws.space.Tree().GetByBranch(b.FullName()); ok {
			return node, trees.Branch, true
		}
	}
	// 实在没有了
	return nil, "", false
}
