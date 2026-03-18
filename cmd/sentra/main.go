package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	_ "go.uber.org/automaxprocs"

	"github.com/anhnmt/sentra/internal/logger"
	"github.com/anhnmt/sentra/internal/runner"
	"github.com/anhnmt/sentra/internal/store"
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

	// Collect system info once for logging and storage
	sysInfo, err := sysinfo.Collect()
	if err != nil {
		log.Warn().Err(err).Msg("could not collect system info")
	}

	if opts.UpdateSignatures {
		if err := updater.UpdateSignatures(); err != nil {
			log.Fatal().Msgf("UpdateSignatures error: %v", err)
		}
		log.Info().Msg("Shutting down...")
		os.Exit(0)
	}

	// Validate scan options
	if opts.Target == "" {
		log.Fatal().Msg("target is required (use --target)")
	}

	r, err := runner.New(opts)
	if err != nil {
		log.Fatal().Msgf("Failed to create runner: %v", err)
	}
	defer r.Close()

	// Save device info to store if available
	if s := r.Store(); s != nil && sysInfo != nil {
		hostname, _ := os.Hostname()
		devInfo := store.DeviceInfo{
			Hostname:    hostname,
			OS:          sysInfo.OS,
			Arch:        sysInfo.Arch,
			CPUModel:    sysInfo.CPUModel,
			CPUCores:    sysInfo.CPUCores,
			TotalRAMMB:  sysInfo.TotalRAMMB,
			CollectedAt: time.Now(),
		}
		if err := s.SaveDevice(devInfo); err != nil {
			log.Warn().Err(err).Msg("failed to save device info")
		}
	}

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
