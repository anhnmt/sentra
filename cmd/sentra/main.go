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
	"github.com/anhnmt/sentra/internal/sysinfo"
	"github.com/anhnmt/sentra/internal/updater"
)

const banner = `
  _________              __                 
 /   _____/ ____   _____/  |_____________   
 \_____  \_/ __ \ /    \   __\_  __ \__  \  
 /        \  ___/|   |  \  |  |  | \// __ \_
/_______  /\___  >___|  /__|  |__|  (____  /
        \/     \/     \/                 \/ 
`

func init() {
	fmt.Printf("%s\n", banner)
	logger.Init()
}

func main() {
	opts, err := runner.ParseOptions()
	if err != nil {
		log.Warn().Err(err).Msg("error parse")
	}

	if _, err := sysinfo.Collect(); err != nil {
		log.Warn().Err(err).Msg("could not collect system info")
	}

	if opts.UpdateSignatures {
		if err := updater.UpdateSignatures(); err != nil {
			log.Fatal().Msgf("UpdateSignatures error: %v", err)
		}
		log.Info().Msg("Shutting down...")
		os.Exit(0)
	}

	r, err := runner.New(opts)
	if err != nil {
		log.Fatal().Msgf("Failed to create runner: %v", err)
	}
	defer r.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		defer cancel()
		if err := r.Run(ctx); err != nil {
			log.Fatal().Msgf("Runner error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Info().Msg("Shutting down...")
}
