package competition

import (
	"context"
	"crypto/tls"
	"github.com/meschbach/npcs/competition/wire"
	"google.golang.org/grpc"
)

type System struct {
	network        GRPCNetwork
	serviceAddress string
	auth           Auth
	tls            *tls.Config
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

func NewCompetitionSystem(auth Auth, listenAt string, onNetwork GRPCNetwork, tlsConfig *tls.Config) *System {
	return &System{
		network:        onNetwork,
		serviceAddress: listenAt,
		auth:           auth,
		tls:            tlsConfig,
		core:           newMatcher(),
	}
}
