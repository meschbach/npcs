package systest

import (
	"context"
	"errors"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/meschbach/npcs/junk/inProc"
	"github.com/meschbach/npcs/t3"
	"github.com/meschbach/npcs/t3/bots"
	"github.com/meschbach/npcs/t3/network"
	"github.com/stretchr/testify/require"
	"github.com/thejerf/suture/v4"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/oauth"
	"sync"
	"testing"
	"time"
)

func TestPushGame(t *testing.T) {
	ctx, done := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(done)

	gateway := network.NewPush()

	net := inProc.NewTestGRPCLayer(t)
	physSrv := net.SpawnService(ctx, "push.npcs:54321", func(ctx context.Context, srv *grpc.Server) error {
		network.RegisterT3PushServer(srv, gateway)
		return nil
	})

	serviceContext, serviceContextDone := context.WithCancel(ctx)
	supervisor := suture.NewSimple("test")
	supervisor.Add(physSrv)
	supervisorResult := supervisor.ServeBackground(serviceContext)
	t.Cleanup(func() {
		serviceContextDone()
		select {
		case err := <-supervisorResult:
			require.NotNil(t, err)
		}
	})

	// configure game clients
	player1Token := faker.Jwt()
	player1GameID, err := uuid.NewV7()
	require.NoError(t, err)
	player1ServiceSide, err := gateway.Router.Register(player1Token, player1GameID.String(), 1)
	require.NoError(t, err)

	player2Token := faker.Jwt()
	player2GameID, err := uuid.NewV7()
	require.NoError(t, err)
	player2ServiceSide, err := gateway.Router.Register(player2Token, player2GameID.String(), 2)
	require.NoError(t, err)

	//launch clients
	gamePlayers := &sync.WaitGroup{}
	gamePlayers.Add(2)
	go func() {
		defer gamePlayers.Done()
		playerCtx, done := context.WithTimeout(ctx, time.Second)
		defer done()

		fillIn := bots.NewFillInBot()
		creds := oauth.TokenSource{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: player1Token})}
		player1Opts := append(net.ConnectOptions(), grpc.WithPerRPCCredentials(creds))
		outgoing := network.NewPushClient("in-proc://push.npcs:54321", player1GameID.String(), player1Token, fillIn, player1Opts...)
		require.NoError(t, outgoing.Serve(playerCtx))
	}()
	go func() {
		defer gamePlayers.Done()
		playerCtx, done := context.WithTimeout(ctx, time.Second)
		defer done()

		columnFiller := bots.NewFillColumn(t3.Move{
			Row:    0,
			Column: 2,
		})
		creds := oauth.TokenSource{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: player2Token})}
		player2Opts := append(net.ConnectOptions(), grpc.WithPerRPCCredentials(creds))
		outgoing := network.NewPushClient("in-proc://push.npcs:54321", player2GameID.String(), player2Token, columnFiller, player2Opts...)
		require.NoError(t, outgoing.Serve(playerCtx))
	}()

	g := t3.NewGame(player1ServiceSide, player2ServiceSide)
	for !g.Concluded() {
		require.NoError(t, g.Step(ctx))
	}
	_, winner := g.Result()
	//todo: consider pushing this down
	p1Err := gateway.Router.GameComplete(ctx, player1Token, player1GameID.String(), winner)
	p2Err := gateway.Router.GameComplete(ctx, player2Token, player2GameID.String(), winner)
	gamePlayers.Wait()
	require.NoError(t, errors.Join(p1Err, p2Err))
}
