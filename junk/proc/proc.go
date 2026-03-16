// Package proc provides process lifecycle management and signal handling.
package proc

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Component interface {
	Start(init context.Context, run context.Context) error
	Stop(context.Context) error
}

// AsService initializes and starts the given component, manages lifecycle signals, and handles proper shutdown operations.
func AsService(c Component) error {
	//nolint // background context is reasonable for root of component hierarchy
	procContext, closeProc := context.WithCancel(context.Background())
	defer closeProc()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	startupProblem := make(chan error, 1)
	startupContext, startupCancel := context.WithCancel(procContext)
	go func() {
		defer startupCancel()
		startupProblem <- c.Start(startupContext, procContext)
	}()

	select {
	case err := <-startupProblem:
		if err != nil {
			return err
		}
	case <-signals:
		startupCancel()
		return <-startupProblem
	}

	sig := <-signals
	fmt.Printf("Received signal: %v, exiting\n", sig)

	// nolint
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.Stop(shutdownCtx)
}
