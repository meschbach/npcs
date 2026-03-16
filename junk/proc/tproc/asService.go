// Package tproc provides test utilities for process lifecycle management.
package tproc

import (
	"context"
	"errors"
	"log/slog"

	"github.com/meschbach/npcs/junk/proc"
)

type tracedWrapper struct {
	c            proc.Component
	name         string
	shutdownOtel func(context.Context) error
}

func (t *tracedWrapper) Start(startup, run context.Context) error {
	var err error
	slog.InfoContext(startup, "Starting OTEL system")
	t.shutdownOtel, err = setupOTelSDK(run, t.name)
	if err != nil {
		slog.ErrorContext(startup, "Failed to start OTEL system", "error", err)
		return err
	}
	slog.InfoContext(startup, "Starting application")
	return t.c.Start(startup, run)
}

func (t *tracedWrapper) Stop(ctx context.Context) error {
	stopError := t.c.Stop(ctx)
	otelShutdownError := t.shutdownOtel(ctx)
	return errors.Join(stopError, otelShutdownError)
}

func AsService(c proc.Component) error {
	return proc.AsService(&tracedWrapper{c: c})
}
