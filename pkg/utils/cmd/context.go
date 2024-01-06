package cmd

import (
	"context"

	"github.com/yhlooo/stackcrisp/pkg/manager"
)

type contextKeyManager struct{}

// NewContextWithManager 将 mgr 注入到上下文中
func NewContextWithManager(parent context.Context, mgr manager.Manager) context.Context {
	return context.WithValue(parent, contextKeyManager{}, mgr)
}

// ManagerFromContext 从上下文获取 manager.Manager
func ManagerFromContext(ctx context.Context) manager.Manager {
	mgr, ok := ctx.Value(contextKeyManager{}).(manager.Manager)
	if ok {
		return mgr
	}
	return nil
}
