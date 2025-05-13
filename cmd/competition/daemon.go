package main

import (
	"context"
	"fmt"
	"github.com/meschbach/npcs/competition"
	"github.com/meschbach/npcs/junk/proc"
	"github.com/spf13/cobra"
	"github.com/thejerf/suture/v4"
	"os"
)

type Daemon struct {
	c       *competition.System
	sysCtx  context.Context
	sysDone func()
	sys     *suture.Supervisor
	sysOut  <-chan error
}

func (d *Daemon) Start(init context.Context, run context.Context) error {
	sysCtx, cancel := context.WithCancel(run)
	d.sysCtx = sysCtx
	d.sysDone = cancel
	d.sys = suture.NewSimple("competitiond")
	d.sys.Add(d.c)
	d.sysOut = d.sys.ServeBackground(sysCtx)
	return nil
}

func (d *Daemon) Stop(ctx context.Context) error {
	d.sysDone()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-d.sysOut:
		return err
	}
}

func daemonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Launches daemon for competitions",
		Run: func(cmd *cobra.Command, args []string) {
			c := competition.NewCompetitionSystem(nil, "localhost:11234", competition.NetworkedGRPC, nil)
			if err := proc.Run(&Daemon{c: c}); err != nil {
				_, e := fmt.Fprintf(os.Stderr, "failed to run process because %e", err)
				if e != nil {
					panic(e)
				}
			}
		},
	}

	return cmd
}
