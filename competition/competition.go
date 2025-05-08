package competition

import (
	"context"
	"crypto/tls"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/thejerf/suture/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"sync"
)

type SystemEventKind int

const (
	SystemEventKindReady SystemEventKind = iota
)

type System struct {
	network        Network
	serviceAddress string
	auth           Auth
	tls            *tls.Config
	Events         chan SystemEventKind
}

func (s *System) Serve(ctx context.Context) error {
	v1 := &v1Service{
		auth:         s.auth,
		lock:         sync.Mutex{},
		persistent:   make(map[string]*persistentPlayer),
		t3MatchesURL: s.serviceAddress,
	}
	transportCreds := credentials.NewTLS(s.tls)
	grpcEvents := make(chan GRPCEventKind, 5)
	socket := &GRPCService{
		network: s.network,
		address: s.serviceAddress,
		register: func(ctx context.Context, srv *grpc.Server) error {
			wire.RegisterCompetitionV1Server(srv, v1)
			return nil
		},
		opts: []grpc.ServerOption{
			grpc.Creds(transportCreds),
		},
		events: grpcEvents,
	}
	subsystems := suture.NewSimple("competition-system")
	subsystems.Add(socket)
	subsystems.Add(&grpcEventAdapter{
		source: grpcEvents,
		target: s.Events,
	})
	return subsystems.Serve(ctx)
}

type grpcEventAdapter struct {
	source chan GRPCEventKind
	target chan SystemEventKind
}

func (g *grpcEventAdapter) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-g.source:
			switch e {
			case GRPCEventReady:
				g.target <- SystemEventKindReady
			default:
			}
		}
	}
}

func NewCompetitionSystem(auth Auth, listenAt string, onNetwork Network, tlsConfig *tls.Config) *System {
	return &System{
		network:        onNetwork,
		serviceAddress: listenAt,
		auth:           auth,
		tls:            tlsConfig,
		Events:         make(chan SystemEventKind, 4),
	}
}

type TestSurface struct {
	system *System
}

func (t *TestSurface) Run(ctx context.Context) {
	s := suture.NewSimple("competition-system")
	s.Add(t.system)
	s.ServeBackground(ctx)
}

func NewIntegHarness(auth Auth, listenAt string, onNetwork Network, tlsConfig *tls.Config) (*TestSurface, *System) {
	system := NewCompetitionSystem(auth, listenAt, onNetwork, tlsConfig)
	return &TestSurface{
		system: system,
	}, system
}
