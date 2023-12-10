package main

import (
	"context"
	"log"
	"syscall"

	"github.com/yhlooo/stackcrisp/pkg/commands"
	ctxutil "github.com/yhlooo/stackcrisp/pkg/utils/context"
)

func main() {
	// 将信号绑定到上下文
	ctx, cancel := ctxutil.Notify(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	// 创建命令
	cmd := commands.NewStackCrispCommand()
	// 执行命令
	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}
