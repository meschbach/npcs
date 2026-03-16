package example

import (
	"context"
	"testing"
	"time"

	"github.com/meschbach/npcs/junk/inproc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type greeterService struct {
	UnimplementedSimpleServer
	prefix string
}

func (g *greeterService) SayHello(ctx context.Context, in *HelloIn) (*HelloOut, error) {
	return &HelloOut{
		Greeting: g.prefix + " " + in.Name,
	}, nil
}

func TestGRPCNetwork(t *testing.T) {
	t.Parallel()
	ctx, done := context.WithTimeout(t.Context(), 1*time.Second)
	defer done()
	prefix := "Greetings"

	net := inproc.NewGRPCNetwork(t)
	l, err := net.Listener(ctx, "localhost:19432", func(ctx context.Context, server *grpc.Server) error {
		RegisterSimpleServer(server, &greeterService{prefix: prefix})
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, l.Start(ctx))
	defer func() {
		require.NoError(t, l.Stop(ctx))
	}()

	conn, err := net.Client(ctx, "localhost:19432")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, conn.Close())
	}()
	reply, err := NewSimpleClient(conn).SayHello(ctx, &HelloIn{Name: "inProc"})
	require.NoError(t, err)
	assert.Equal(t, prefix+" inProc", reply.Greeting)
}
