package proc

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Component interface {
	Start(init context.Context, run context.Context) error
	Stop(context.Context) error
}

// Run initializes and starts the given component, manages lifecycle signals, and handles proper shutdown operations.
func Run(c Component) error {
	procContext, closeProc := context.WithCancel(context.Background())
	defer closeProc()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP)

	if err := runStartup(procContext, c); err != nil {
		return err
	}

	<-signals

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.Stop(shutdownCtx)
}

func runStartup(procContext context.Context, c Component) error {
	startupContext, startupCancel := context.WithCancel(procContext)
	defer startupCancel()
	return c.Start(startupContext, procContext)
}
