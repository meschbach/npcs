package tui

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/npcs/t3"
	"github.com/meschbach/npcs/t3/bots"
	"github.com/meschbach/npcs/t3/network"
	"google.golang.org/grpc"
)

func RunGame(ctx context.Context, serviceURL string) (err error) {
	fmt.Printf("Connecting to bot...")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.NewClient(serviceURL, opts...)
	if err != nil {
		fmt.Printf("\nFailed to connect to bot because %s\n", err.Error())
		return nil
	}
	defer func() {
		err = errors.Join(err, conn.Close())
	}()

	hub := network.NewT3Client(conn)
	bot, err := network.NewRemotePlayer(ctx, hub, 1)
	if err != nil {
		fmt.Printf("\nFailed to spawn bot because %s\n", err.Error())
		return nil
	}
	fmt.Printf("done.\n")

	human := &simple{player: 2}

	game := t3.NewGame(bot, human)
	for !game.Concluded() {
		if err := game.Step(ctx); err != nil {
			return err
		}
	}
	fmt.Printf("Game concluded.\n")
	_, winner := game.Result()
	fmt.Printf("Winning player: %d\n", winner)
	return nil
}

func RunFillIn(ctx context.Context) (err error) {
	bot := bots.NewFillInBot()
	human := &simple{player: 2}

	game := t3.NewGame(network.WithLocalSession(bot), human)
	for !game.Concluded() {
		if err := game.Step(ctx); err != nil {
			return err
		}
	}
	fmt.Printf("Game concluded.\n")
	_, winner := game.Result()
	fmt.Printf("Winning player: %d\n", winner)
	return nil
}
