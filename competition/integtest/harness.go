package integtest

import (
	"context"
	"crypto/tls"
	"github.com/madflojo/testcerts"
	"github.com/meschbach/npcs/competition"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/meschbach/npcs/junk/inProc"
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

var serviceAddress = "competition.npcs:11234"

func WithCompetitionSystem(t *testing.T, fn func(ctx context.Context, t *testing.T, h *Harness)) {
	ctx, done := context.WithTimeout(context.Background(), 1*time.Second)
	t.Cleanup(done)
	h := &Harness{
		t:         t,
		ctx:       ctx,
		inProcNet: inProc.NewNetwork(),
		Auth:      NewFakeAuth(),
		CA:        testcerts.NewCA(),
	}
	t.Cleanup(func() {
		h.cleanUp()
	})
	servicePair, err := h.CA.NewKeyPair("competition.npcs")
	require.NoError(t, err)
	serviceCfg, err := servicePair.ConfigureTLSConfig(&tls.Config{})
	require.NoError(t, err)

	//spawn the base system
	surface, system := competition.NewIntegHarness(h.Auth, serviceAddress, h.inProcNet, serviceCfg)
	h.Surface = surface
	h.System = system

	surface.Run(ctx)
	select {
	case <-h.ctx.Done():
		require.NoError(t, h.ctx.Err())
	case <-system.Events:
	}
	fn(ctx, t, h)
}

type Harness struct {
	t         *testing.T
	ctx       context.Context
	inProcNet *inProc.Network
	Surface   *competition.TestSurface
	System    *competition.System
	Auth      *FakeAuth
	CA        *testcerts.CertificateAuthority

	changes   sync.Mutex
	cleanedUp bool
	cleanup   []func(ctx context.Context) error
}

func (h *Harness) cleanUp() {
	h.changes.Lock()
	defer h.changes.Unlock()

	h.ensureActiveLocked("already cleaned up")
}

func (h *Harness) onCleanup(f func(ctx context.Context) error) {
	h.changes.Lock()
	defer h.changes.Unlock()

	h.ensureActiveLocked("not active")
	h.cleanup = append(h.cleanup, f)
}

func (h *Harness) ensureActive(what string) {
	h.changes.Lock()
	defer h.changes.Unlock()
	h.ensureActiveLocked(what)
}

func (h *Harness) ensureActiveLocked(what string) {
	if h.cleanedUp {
		panic(what)
	}
}

func (h *Harness) NewClient(userToken string) wire.CompetitionV1Client {
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})

	p := manual.NewBuilderWithScheme("in-proc")
	p.InitialState(resolver.State{
		Addresses: []resolver.Address{
			{Addr: serviceAddress, ServerName: "competition.npcs"},
		},
	})

	perRPC := oauth.TokenSource{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: userToken})}
	conn, err := grpc.NewClient("in-proc://"+serviceAddress, grpc.WithContextDialer(h.inProcNet.Dial), grpc.WithPerRPCCredentials(perRPC), grpc.WithTransportCredentials(creds), grpc.WithResolvers(p))
	require.NoError(h.t, err)
	return wire.NewCompetitionV1Client(conn)
}
