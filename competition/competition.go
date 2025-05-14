package competition

import (
	"context"
	"crypto/tls"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/thejerf/suture/v4"
	"google.golang.org/grpc"
)

type SystemEventKind int

const (
	SystemEventKindReady SystemEventKind = iota
)

type System struct {
	network        GRPCNetwork
	serviceAddress string
	auth           Auth
	tls            *tls.Config
	Events         chan SystemEventKind
	core           *matcher
}

func (s *System) Serve(ctx context.Context) error {
	v1 := &clientCompetitionService{
		auth: s.auth,
		core: s.core,
	}
	registry := newGameRegistryService(s.core)
	enginePlaneOrchestration := newEngines(s.core)

	listener, err := s.network.Listener(ctx, s.serviceAddress, func(ctx context.Context, server *grpc.Server) error {
		wire.RegisterCompetitionV1Server(server, v1)
		wire.RegisterGameRegistryServer(server, registry)
		wire.RegisterGameEngineOrchestrationServer(server, enginePlaneOrchestration)
		return nil
	})
	if err != nil {
		return err
	}
	if err := listener.Start(ctx); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return listener.Stop(ctx)
	}
}

// todo remove this structure
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

func NewCompetitionSystem(auth Auth, listenAt string, onNetwork GRPCNetwork, tlsConfig *tls.Config) *System {
	return &System{
		network:        onNetwork,
		serviceAddress: listenAt,
		auth:           auth,
		tls:            tlsConfig,
		Events:         make(chan SystemEventKind, 4),
		core:           newMatcher(),
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
