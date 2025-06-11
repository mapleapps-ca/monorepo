// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/cmd/daemon/daemon.go
package daemon

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/manifold"
)

func DaemonCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "daemon",
		Short: "Run the cloud-services backend",
		Run: func(cmd *cobra.Command, args []string) {
			doRunDaemon()
		},
	}
	return cmd
}

func doRunDaemon() {
	// Create context that listens for OS signals
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer cancel()

	// Create the FX application with enhanced configuration
	app := fx.New(
		// Timeouts
		fx.StartTimeout(5*time.Minute),
		fx.StopTimeout(2*time.Minute),

		// Logging - Use environment-based logger selection
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),

		// Core providers
		fx.Provide(newEnvironmentLogger),
		fx.Provide(config.NewProvider),

		// Main application modules (includes pkg.Module() which has observability)
		manifold.Module(),

		// Lifecycle management
		fx.Invoke(registerLifecycleHooks),
	)

	// Start the application
	if err := app.Start(ctx); err != nil {
		app.Done()
		os.Exit(1)
	}

	// Wait for shutdown signal
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.Stop(shutdownCtx); err != nil {
		// Log error but don't exit with error code as we're shutting down
		// The logger might not be available at this point
		os.Exit(1)
	}
}

// newEnvironmentLogger creates appropriate logger based on environment
func newEnvironmentLogger(cfg *config.Configuration) (*zap.Logger, error) {
	if cfg.App.Environment == "production" {
		return newProductionLogger()
	}
	return zap.NewDevelopment()
}

// newProductionLogger creates a production-ready logger
func newProductionLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()

	// Customize based on environment variables
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		var zapLevel zap.AtomicLevel
		if err := zapLevel.UnmarshalText([]byte(level)); err == nil {
			config.Level = zapLevel
		}
	}

	// Set encoding based on LOG_FORMAT
	if format := os.Getenv("LOG_FORMAT"); format == "console" {
		config.Encoding = "console"
		config.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	logger, err := config.Build(zap.AddCaller())
	if err != nil {
		return nil, err
	}

	// Add service information
	return logger.With(
		zap.String("service", "mapleapps-backend"),
		zap.String("version", getVersion()),
	), nil
}

// registerLifecycleHooks sets up application lifecycle management
func registerLifecycleHooks(
	lc fx.Lifecycle,
	logger *zap.Logger,
	cfg *config.Configuration,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("MapleApps Backend starting",
				zap.String("version", getVersion()),
				zap.String("main_port", cfg.App.Port),
				zap.String("observability_port", "8080"),
				zap.String("environment", cfg.App.Environment),
				zap.String("log_level", cfg.Logging.Level),
			)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("MapleApps Backend shutting down gracefully")
			return nil
		},
	})
}

func getVersion() string {
	version := os.Getenv("SERVICE_VERSION")
	if version == "" {
		return "1.0.0"
	}
	return version
}
