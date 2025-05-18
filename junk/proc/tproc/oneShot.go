package tproc

import (
	"context"
	"github.com/meschbach/npcs/junk/proc"
	"log/slog"
)

func RunOnce(name string, run func(ctx context.Context) error) error {
	return proc.RunOnce(func(ctx context.Context) error {
		var err error
		slog.InfoContext(ctx, "Starting OTEL system")
		shutdownOtel, err := setupOTelSDK(ctx, name)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to start OTEL system", "error", err)
			return err
		}
		defer shutdownOtel(ctx)
		slog.InfoContext(ctx, "Starting application")
		return run(ctx)
	})
}
