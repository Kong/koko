package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/kong/koko/internal/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM,
		syscall.SIGINT)
	defer stop()
	cmd.ExecuteContext(ctx)
}
