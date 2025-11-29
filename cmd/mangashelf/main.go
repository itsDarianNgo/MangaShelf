package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/mangashelf/mangashelf/internal/api"
	"github.com/mangashelf/mangashelf/internal/config"
	"github.com/mangashelf/mangashelf/internal/database"
	"github.com/mangashelf/mangashelf/internal/library"
	"github.com/mangashelf/mangashelf/internal/scraper"
	"github.com/mangashelf/mangashelf/internal/scraper/mangadex"
)

var (
	cfgFile        string
	dataDir        string
	hostFlag       string
	portFlag       int
	workersFlag    int
	verbose        bool
	disableUpdates bool
	skipScan       bool
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mangashelf",
		Short: "Self-hosted manga downloader and reader",
		RunE:  run,
	}

	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Path to config file")
	cmd.Flags().StringVarP(&dataDir, "data", "d", "./data", "Data directory for library and database")
	cmd.Flags().StringVarP(&hostFlag, "host", "H", "", "Server host override")
	cmd.Flags().IntVarP(&portFlag, "port", "p", 0, "Server port override")
	cmd.Flags().IntVarP(&workersFlag, "workers", "w", 0, "Number of download workers")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging")
	cmd.Flags().BoolVar(&disableUpdates, "no-update", false, "Disable automatic updates")
	cmd.Flags().BoolVar(&skipScan, "no-scan", false, "Skip library scan on startup")

	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	applyFlagOverrides(cmd, cfg)
	logger := buildLogger(cfg)

	if err := os.MkdirAll(filepath.Dir(cfg.Database.Path), 0o755); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	db, err := database.Open(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	logger.Info().Str("path", cfg.Database.Path).Msg("database initialized")

	scraperMgr := scraper.NewManager(logger)
	scraperMgr.Register(mangadex.New(cfg.Sources.Mangadex.Language))

	queries := database.New(db)
	libService := library.NewService(queries, scraperMgr, logger)

	router := api.NewRouter(logger, scraperMgr, libService)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	srv := &http.Server{ //nolint:exhaustruct
		Addr:    addr,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("failed to shutdown server")
		}
	}()

	logger.Info().Msgf("starting server on %s", addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}

func applyFlagOverrides(cmd *cobra.Command, cfg *config.Config) {
	if cmd.Flags().Changed("data") {
		cfg.Library.Path = filepath.Join(dataDir, "manga")
		cfg.Sources.CustomPath = filepath.Join(dataDir, "scrapers")
		cfg.Database.Path = filepath.Join(dataDir, "mangashelf.db")
		cfg.Logging.File.Path = filepath.Join(dataDir, "logs", "mangashelf.log")
	}

	if cmd.Flags().Changed("host") {
		cfg.Server.Host = hostFlag
	}

	if cmd.Flags().Changed("port") {
		cfg.Server.Port = portFlag
	}

	if cmd.Flags().Changed("workers") {
		cfg.Downloader.Workers = workersFlag
	}

	if verbose {
		cfg.Logging.Level = "debug"
	}

	if disableUpdates {
		cfg.Updates.Enabled = false
	}

	if skipScan {
		cfg.Library.ScanOnStartup = false
	}
}

func buildLogger(cfg *config.Config) zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	output := os.Stdout
	if cfg.Logging.Format == "text" {
		cw := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		return zerolog.New(cw).With().Timestamp().Logger()
	}

	return log.Output(output).With().Timestamp().Logger()
}
