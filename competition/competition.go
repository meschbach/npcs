package competition

import (
	"context"
	"crypto/tls"
	"github.com/meschbach/npcs/competition/wire"
	"github.com/meschbach/npcs/junk/realnet"
	"google.golang.org/grpc"
	"log/slog"
	"sync"
)

type SystemPhase int

const (
	SystemStarting SystemPhase = iota
	SystemReady
	SystemStopped
)

type System struct {
	network        realnet.GRPCNetwork
	serviceAddress string
	auth           Auth
	tls            *tls.Config
	core           *matcher

	state       *sync.Mutex
	stateChange *sync.Cond
	phase       SystemPhase
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
	s.transitionToReady()
	slog.InfoContext(ctx, "Competition system ready")
	select {
	case <-ctx.Done():
		return listener.Stop(ctx)
	}
}

func (s *System) transitionToReady() {
	s.state.Lock()
	defer s.state.Unlock()

	s.phase = SystemReady
	s.stateChange.Broadcast()
}

func (s *System) WaitForReady() {
	s.state.Lock()
	defer s.state.Unlock()

	for s.phase == SystemStarting {
		s.stateChange.Wait()
	}
}

func NewCompetitionSystem(auth Auth, listenAt string, onNetwork realnet.GRPCNetwork, tlsConfig *tls.Config) *System {
	lock := &sync.Mutex{}
	return &System{
		network:        onNetwork,
		serviceAddress: listenAt,
		auth:           auth,
		tls:            tlsConfig,
		core:           newMatcher(),
		state:          lock,
		stateChange:    sync.NewCond(lock),
		phase:          SystemStarting,
	}
}
