package tproc

import (
	"context"
	"errors"
	"log/slog"

	"github.com/meschbach/npcs/junk/proc"
)

func RunOnce(name string, run func(context.Context) error) error {
	return proc.RunOnce(func(ctx context.Context) (problem error) {
		var err error
		slog.InfoContext(ctx, "Starting OTEL system")
		shutdownOtel, err := setupOTelSDK(ctx, name)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to start OTEL system", "error", err)
			return err
		}
		defer func() {
			problem = errors.Join(shutdownOtel(ctx), err)
		}()
		slog.InfoContext(ctx, "Starting application")
		return run(ctx)
	})
}
