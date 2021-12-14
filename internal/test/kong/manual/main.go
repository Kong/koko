package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/kong/koko/internal/test/kong"
)

func main() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	err := kong.RunDP(ctx, kong.DockerInput{Image: "2.5.0"})
	if err != nil {
		fmt.Printf("%v\n", err)
		defer os.Exit(1)
	}
}
