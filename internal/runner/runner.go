package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/anhnmt/sentra/internal/core"
	"github.com/anhnmt/sentra/internal/detectors/yara"
	"github.com/anhnmt/sentra/internal/logger"
	"github.com/anhnmt/sentra/internal/progress"
	"github.com/anhnmt/sentra/internal/report"
	"github.com/anhnmt/sentra/internal/store"
	"github.com/anhnmt/sentra/internal/worker"
)

const defaultDBPath = "sentra.db"

type Runner struct {
	opts     *Options
	detector *yara.YaraDetector
	ioPool   *worker.Pool
	scanPool *worker.Pool
	skipDirs map[string]struct{}
	store    *store.Store
	session  *store.ScanSession
}

type scanResult struct {
	matches []core.MatchResult
	err     error
}

// New creates a new Runner with optional bbolt store.
func New(opts *Options) (*Runner, error) {
	dbPath := opts.DBPath
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	var db *store.Store
	if dbPath != "" && dbPath != ":memory:" {
		var err error
		db, err = store.Open(dbPath)
		if err != nil {
			log.Warn().Err(err).Msgf("failed to open db, running without history")
		}
	}

	detector, err := yara.New(opts.RulesDir)
	if err != nil {
		return nil, fmt.Errorf("init yara: %w", err)
	}

	ioPool, err := worker.New(&worker.Options{Size: opts.Workers})
	if err != nil {
		return nil, fmt.Errorf("init io pool: %w", err)
	}

	scanPool, err := worker.NewScanPool()
	if err != nil {
		return nil, fmt.Errorf("init scan pool: %w", err)
	}

	skipDirs := make(map[string]struct{}, len(opts.SkipDirs))
	for _, d := range opts.SkipDirs {
		skipDirs[d] = struct{}{}
	}

	return &Runner{
		opts:     opts,
		detector: detector,
		ioPool:   ioPool,
		scanPool: scanPool,
		skipDirs: skipDirs,
		store:    db,
	}, nil
}

// Store returns the underlying bbolt store (may be nil).
func (r *Runner) Store() *store.Store {
	return r.store
}

func (r *Runner) Run(ctx context.Context) error {
	if r.opts.Target == "" {
		return fmt.Errorf("--target is required")
	}
	if abs, err := filepath.Abs(r.opts.Target); err == nil {
		r.opts.Target = abs
	}

	log.Info().
		Str("target", r.opts.Target).
		Str("rules_dir", r.opts.RulesDir).
		Int("workers", r.opts.Workers).
		Int("max_depth", r.opts.MaxDepth).
		Strs("skip_dirs", r.opts.SkipDirs).
		Msg("scan starting")

	// Generate scan ID and DB path with timestamp
	scanID := time.Now().Format("20060102150405")

	// If output is requested, generate DB and report paths with scan ID
	if r.opts.OutputPath != "" {
		dbPath := fmt.Sprintf("sentra-%s.db", scanID)
		r.opts.DBPath = dbPath

		// Close existing store if any
		if r.store != nil {
			r.store.Close()
		}

		// Open new store with scan ID in path
		var err error
		r.store, err = store.Open(dbPath)
		if err != nil {
			log.Warn().Err(err).Msgf("failed to open db %s, running without history", dbPath)
			r.store = nil
		}
	}

	// Begin scan session if store is available
	if r.store != nil {
		rec := store.ScanRecord{
			ID:        scanID,
			StartedAt: time.Now(),
			Target:    r.opts.Target,
			RulesDir:  r.opts.RulesDir,
			Workers:   r.opts.Workers,
			Status:    "running",
		}
		session, err := r.store.BeginScan(rec)
		if err != nil {
			log.Warn().Err(err).Msg("failed to begin scan session")
		} else {
			r.session = session
			// Log scan start
			session.RecordLog(store.LevelInfo, "scan started", map[string]any{
				"target":    r.opts.Target,
				"rules_dir": r.opts.RulesDir,
				"workers":   r.opts.Workers,
			})
		}
	}

	bar := progress.New(progress.Options{Workers: r.opts.Workers})
	logger.InitWithWriter(bar.Writer)
	start := time.Now()

	ch := make(chan scanResult, r.opts.Workers)
	var wg sync.WaitGroup

	errs, errCount, done := r.startConsumer(ch, bar)

	var walkErr error
	go func() {
		walkErr = r.walkFiles(ctx, ch, &wg, bar)
		wg.Wait()
		close(ch)
	}()

	<-done
	bar.Done()
	logger.InitWithWriter(bar.Writer)

	canceled := errors.Is(ctx.Err(), context.Canceled)
	r.logScanSummary(ctx, bar, start, *errCount)

	// Finish scan session
	if r.session != nil {
		status := "completed"
		if canceled {
			status = "canceled"
		}
		r.session.Finish(bar.Files(), bar.Skipped(), bar.Matches(), *errCount, status)
		r.session.RecordLog(store.LevelInfo, "scan "+status, map[string]any{
			"scanned":  bar.Files(),
			"skipped":  bar.Skipped(),
			"matches":  bar.Matches(),
			"errors":   *errCount,
			"duration": time.Since(start).Round(time.Second).String(),
		})

		// Auto-generate HTML report if OutputPath is set
		if r.opts.OutputPath != "" {
			// Close store first to release database lock
			if r.store != nil {
				r.store.Close()
				r.store = nil
			}

			// Generate output path with scan ID
			outputPath := r.opts.OutputPath
			if !strings.HasSuffix(outputPath, ".html") {
				// Assume it's a directory or prefix, append scan ID
				outputPath = fmt.Sprintf("%s/report-%s.html", outputPath, r.session.Record.ID)
			}

			cmdLine := strings.Join(os.Args, " ")
			gen := report.NewGenerator(r.opts.DBPath)
			if err := gen.Generate(r.session.Record.ID, outputPath, cmdLine); err != nil {
				log.Warn().Err(err).Msg("failed to generate report")
			} else {
				log.Info().Str("report", outputPath).Msg("HTML report generated")
			}
		}
	}

	if walkErr != nil && !errors.Is(walkErr, context.Canceled) {
		return fmt.Errorf("walk %s: %w", r.opts.Target, walkErr)
	}
	if len(*errs) > 0 && !canceled {
		return fmt.Errorf("scan errors: %v", *errs)
	}
	return nil
}

func (r *Runner) startConsumer(ch <-chan scanResult, bar *progress.Bar) (*[]error, *int64, <-chan struct{}) {
	var (
		mu         sync.Mutex
		errs       []error
		errorCount int64
		done       = make(chan struct{})
	)

	go func() {
		defer close(done)
		for res := range ch {
			if res.err != nil {
				if !errors.Is(res.err, os.ErrPermission) {
					errorCount++
					mu.Lock()
					errs = append(errs, res.err)
					mu.Unlock()
					// Record error to DB
					if r.session != nil {
						r.session.RecordLog(store.LevelError, "scan error", map[string]any{
							"error": res.err.Error(),
						})
					}
				}
				continue
			}
			for _, match := range res.matches {
				bar.IncrementMatch()
				log.Warn().
					Str("detector", match.DetectorName).
					Str("rule", match.RuleName).
					Str("file", match.Target).
					Fields(match.Metadata).
					Msg("match detected")

				// Record match to DB
				if r.session != nil {
					rec := store.MatchRecord{
						ScanID:       r.session.Record.ID,
						DetectorName: match.DetectorName,
						RuleName:     match.RuleName,
						Target:       match.Target,
						Metadata:     match.Metadata,
						DetectedAt:   time.Now(),
					}
					r.session.RecordMatch(rec)
					// Also record as log
					r.session.RecordLog(store.LevelWarn, "match detected", map[string]any{
						"detector": match.DetectorName,
						"rule":     match.RuleName,
						"file":     match.Target,
					})
				}
			}
		}
	}()

	return &errs, &errorCount, done
}

func (r *Runner) logScanSummary(ctx context.Context, bar *progress.Bar, start time.Time, errorCount int64) {
	dur := time.Since(start).Round(time.Second).String()
	canceled := errors.Is(ctx.Err(), context.Canceled)

	if canceled {
		log.Warn().
			Int64("scanned", bar.Files()).
			Int64("skipped", bar.Skipped()).
			Int64("matches", bar.Matches()).
			Str("duration", dur).
			Msg("scan canceled")
		return
	}

	log.Info().
		Int64("scanned", bar.Files()).
		Int64("skipped", bar.Skipped()).
		Int64("matches", bar.Matches()).
		Int64("errors", errorCount).
		Str("duration", dur).
		Msg("scan complete")
}

func (r *Runner) Close() {
	r.scanPool.Close()
	r.ioPool.Close()
	r.detector.Close()
	if r.store != nil {
		r.store.Close()
	}
	log.Info().Msg("Goodbye!")
}
