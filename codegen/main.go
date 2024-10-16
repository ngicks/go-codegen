package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/ngicks/go-codegen/codegen/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	cmd.Execute(ctx)
}
