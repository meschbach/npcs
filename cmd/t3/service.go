package main

import (
	"context"
	"fmt"
	"github.com/meschbach/npcs/t3/bots"
	"github.com/meschbach/npcs/t3/network"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"net"
)

func serviceCommands() *cobra.Command {
	fillIn := &cobra.Command{Use: "fill-in", Short: "starts fill in service",
		Run: func(cmd *cobra.Command, args []string) {
			hub := network.NewHub(func(ctx context.Context) (network.Session, error) {
				return bots.NewFillInBot(), nil
			})
			s := grpc.NewServer()
			network.RegisterT3Server(s, hub)

			l, err := net.Listen("tcp", "localhost:3333")
			if err != nil {
				fmt.Printf("failed to listen: %v", err)
				return
			}
			if err := s.Serve(l); err != nil {
				fmt.Printf("failed to serve: %v", err)
			}
		},
	}
	serviceCmd := &cobra.Command{
		Use: "services",
	}
	serviceCmd.AddCommand(fillIn)

	return serviceCmd
}
