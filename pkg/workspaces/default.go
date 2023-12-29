package workspaces

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"

	"github.com/yhlooo/stackcrisp/pkg/mounts"
	"github.com/yhlooo/stackcrisp/pkg/spaces"
	fsutil "github.com/yhlooo/stackcrisp/pkg/utils/fs"
	"github.com/yhlooo/stackcrisp/pkg/utils/uid"
)

const (
	loggerName = "workspaces"
)

// New 创建一个工作空间
func New(path string, head uid.UID, space spaces.Space, mount mounts.Mount) Workspace {
	return &defaultWorkspace{
		path:  path,
		head:  head,
		space: space,
		mount: mount,
	}
}

// defaultWorkspace 是 Workspace 的一个默认实现
type defaultWorkspace struct {
	path  string
	head  uid.UID
	space spaces.Space
	mount mounts.Mount
}

var _ Workspace = &defaultWorkspace{}

// Path 返回工作空间路径
func (ws *defaultWorkspace) Path() string {
	return ws.path
}

// Head 返回头指针
func (ws *defaultWorkspace) Head() uid.UID {
	return ws.head
}

// Space 返回工作空间对应空间
func (ws *defaultWorkspace) Space() spaces.Space {
	return ws.space
}

// Mount 返回工作空间对应挂载
func (ws *defaultWorkspace) Mount() mounts.Mount {
	return ws.mount
}

// Expand 展开工作空间
func (ws *defaultWorkspace) Expand(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	// 挂载
	logger.Info("mounting ...")
	if err := ws.Mount().Mount(ctx); err != nil {
		return fmt.Errorf("mount error: %w", err)
	}

	// 确保目标路径上什么也没有
	if fsutil.IsExists(ws.Path()) {
		logger.V(1).Info(fmt.Sprintf("target path %q exists, remove it", ws.Path()))
		if err := os.Remove(ws.Path()); err != nil {
			return fmt.Errorf("clear path %q error: %w", ws.Path(), err)
		}
	}

	// 先移动到根目录
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get pwd error: %w", err)
	}
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("change working directory to \"/\" error: %w", err)
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
