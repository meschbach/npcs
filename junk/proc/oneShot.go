package proc

import (
	"context"
	"os/signal"
	"syscall"
)

func RunOnce(run func(context.Context) error) error {
	procContext, closeProc := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP)
	defer closeProc()

	return run(procContext)
}
