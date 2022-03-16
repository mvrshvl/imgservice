package main

import (
	"context"
	"log"

	"imgservice/config"
	"imgservice/core/runner"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := runner.NewRunner()

	cfg, err := config.New(config.ConfigPath)
	if err != nil {
		log.Fatalf("can't create config: %v", err)
	}

	ctx, err = r.Run(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	r.Wait(ctx)
}
