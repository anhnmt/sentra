package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	_ "go.uber.org/automaxprocs"

	"github.com/anhnmt/sentra/internal/runner"
)

const BANNER = `
  _________              __                 
 /   _____/ ____   _____/  |_____________   
 \_____  \_/ __ \ /    \   __\_  __ \__  \  
 /        \  ___/|   |  \  |  |  | \// __ \_
/_______  /\___  >___|  /__|  |__|  (____  /
        \/     \/     \/                 \/ 
`

func init() {
	fmt.Printf("%s\n", BANNER)

}

func main() {
	opts := runner.ParseOptions()

	runner, err := runner.New(opts)
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		if err := runner.Run(ctx); err != nil {
			log.Fatalf("Runner error: %v", err)
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	fmt.Println("\nShutting down...")

	runner.Close()
}
