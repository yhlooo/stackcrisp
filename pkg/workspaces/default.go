package workspaces

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"

	"github.com/yhlooo/stackcrisp/pkg/mounts"
	"github.com/yhlooo/stackcrisp/pkg/spaces"
	"github.com/yhlooo/stackcrisp/pkg/spaces/trees"
	fsutil "github.com/yhlooo/stackcrisp/pkg/utils/fs"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

const (
	loggerName = "workspaces"

	remoteBranchPrefixInTree = "global/"
	remoteBranchPrefix       = "origin/"
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
func (ws *defaultWorkspace) Branch() string {
	return ws.branch
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

// LocalBranches 返回本地分支列表
func (ws *defaultWorkspace) LocalBranches() []string {
	var ret []string
	for branch := range ws.Space().Tree().Branches() {
		if strings.HasPrefix(branch, remoteBranchPrefixInTree) {
			continue
		}
		ret = append(ret, strings.TrimPrefix(branch, ws.id.Base32()+"/"))
	}
	return ret
}

// RemoteBranches 返回远程分支列表
func (ws *defaultWorkspace) RemoteBranches() []string {
	var ret []string
	for branch := range ws.Space().Tree().Branches() {
		if !strings.HasPrefix(branch, remoteBranchPrefixInTree) {
			continue
		}
		ret = append(ret, remoteBranchPrefix+strings.TrimPrefix(branch, remoteBranchPrefixInTree))
	}
	return ret
}
