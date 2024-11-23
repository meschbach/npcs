package systest

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/madflojo/testcerts"
	"github.com/meschbach/npcs/junk/inProc"
	"github.com/meschbach/npcs/t3"
	"github.com/meschbach/npcs/t3/bots"
	"github.com/meschbach/npcs/t3/network"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
	"sync"
	"testing"
	"time"
)

func TestPushGame(t *testing.T) {
	ctx, done := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(done)

	gateway := network.NewPush()

	net := inProc.NewNetwork()
	l, err := net.Listen(ctx, "push.npcs:54321")
	require.NoError(t, err)

	ca := testcerts.NewCA()
	service, err := ca.NewKeyPair("push.npcs")
	require.NoError(t, err)

	// spawn and register service
	cfg, err := service.ConfigureTLSConfig(&tls.Config{})
	require.NoError(t, err)
	s := grpc.NewServer(grpc.Creds(credentials.NewTLS(cfg)))
	network.RegisterT3PushServer(s, gateway)

	go func() {
		s.Serve(l)
		//todo: track the resulting error
		//err := s.Serve(l)
		//require.NoError(t, err)
	}()
	t.Cleanup(func() {
		require.NoError(t, l.Close())
	})

	//build grpc client stuff
	grpcDiscovery := manual.NewBuilderWithScheme("in-proc")
	grpcDiscovery.InitialState(resolver.State{
		Addresses: []resolver.Address{
			{Addr: "push.npcs:54321", ServerName: "push.npcs"},
		},
	})
	grpcClientOpts := []grpc.DialOption{
		grpc.WithResolvers(grpcDiscovery),
		grpc.WithContextDialer(net.Dial),
		grpc.WithTransportCredentials(credentials.NewTLS(ca.GenerateTLSConfig())),
	}

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
		optsCopy := make([]grpc.DialOption, len(grpcClientOpts))
		copy(optsCopy, grpcClientOpts)
		creds := oauth.TokenSource{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: player1Token})}
		player1Opts := append(optsCopy, grpc.WithPerRPCCredentials(creds))
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
		optsCopy := make([]grpc.DialOption, len(grpcClientOpts))
		copy(optsCopy, grpcClientOpts)
		creds := oauth.TokenSource{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: player2Token})}
		player2Opts := append(optsCopy, grpc.WithPerRPCCredentials(creds))
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
