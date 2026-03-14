package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	_ "go.uber.org/automaxprocs"

	"github.com/anhnmt/sentra/internal/logger"
	"github.com/anhnmt/sentra/internal/runner"
	"github.com/anhnmt/sentra/internal/util"
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
	logger.Init()

}

func main() {
	opts := runner.ParseOptions()

	if opts.UpdateSignatures {
		err := util.UpdateSignatures()
		if err != nil {
			log.Fatal().Msgf("UpdateSignatures error: %v", err)
			return
		}
		log.Info().Msgf("Shutting down...")
		os.Exit(1)
	}

	runner, err := runner.New(opts)
	if err != nil {
		log.Info().Msgf("Failed to create runner: %v", err)
	}
	defer runner.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		defer cancel()
		if err := runner.Run(ctx); err != nil {
			log.Fatal().Msgf("Runner error: %v", err)
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	log.Info().Msgf("Shutting down...")
}
