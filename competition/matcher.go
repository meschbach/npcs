package competition

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"log/slog"
	"sync"
)

type matchedGame struct {
	instanceURL string
	instanceID  string
}

type gameInstancePhase int

const (
	waitingForPlayers gameInstancePhase = iota
	playing
	completed
)

type gameInstanceLobby struct {
	phase        gameInstancePhase
	desiredCount int
	playerCount  int
	instanceURL  string
	instanceID   string
}

type gameLobby struct {
	state     *sync.Mutex
	waiting   *sync.Cond
	available []*gameInstanceLobby
	playing   []*gameInstanceLobby
	completed []*gameInstanceLobby
}

type matcher struct {
	state       *sync.Mutex
	gameMatches map[string]*gameLobby
}

func newMatcher() *matcher {
	return &matcher{
		state:       &sync.Mutex{},
		gameMatches: make(map[string]*gameLobby),
	}
}

func (m *matcher) ensureGame(ctx context.Context, game string) error {
	m.state.Lock()
	defer m.state.Unlock()
	if _, ok := m.gameMatches[game]; !ok {
		slog.InfoContext(ctx, "matcher.ensureGame -- registering", "Game", game)
		stateLock := &sync.Mutex{}
		m.gameMatches[game] = &gameLobby{
			state:   stateLock,
			waiting: sync.NewCond(stateLock),
		}
	}
	return nil
}

func (m *matcher) registerInstance(ctx context.Context, game, instanceURL string) (string, error) {
	slog.InfoContext(ctx, "matcher.registerInstance", "Game", game, "instanceURL", instanceURL)
	lobby, err := m.findGameLobby(ctx, game)
	if err != nil {
		return "", err
	}

	structuredID, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	id := structuredID.String()

	lobby.state.Lock()
	defer lobby.state.Unlock()

	lobby.available = append(lobby.available, &gameInstanceLobby{
		phase:        waitingForPlayers,
		desiredCount: 2,
		playerCount:  0,
		instanceURL:  instanceURL,
		instanceID:   id,
	})
	return id, nil
}

func (m *matcher) findMatchInstance(ctx context.Context, game string) (*matchedGame, error) {
	slog.InfoContext(ctx, "matcher.findMatchInstance", "Game", game)
	lobby, err := m.findGameLobby(ctx, game)
	if err != nil {
		return nil, err
	}

	lobby.state.Lock()
	defer lobby.state.Unlock()
	if len(lobby.available) == 0 {
		slog.InfoContext(ctx, "matcher.findMatchInstance -- no instances available", "Game", game)
		return nil, errors.New("no game instances available")
	}

	instance := lobby.available[0]
	instance.playerCount++
	for instance.playerCount < instance.desiredCount {
		slog.InfoContext(ctx, "matcher.findMatchInstance -- waiting on additional players", "Game", game, "instance", instance.instanceID, "playerCount", instance.playerCount, "desiredCount", instance.desiredCount)
		lobby.waiting.Wait()
	}
	if instance.phase == waitingForPlayers {
		slog.InfoContext(ctx, "matcher.findMatchInstance -- first awoke", "Game", game)
		instance.phase = playing
		lobby.playing = append(lobby.playing, instance)
		lobby.available = lobby.available[1:]
	}
	lobby.waiting.Signal()
	return &matchedGame{instanceURL: instance.instanceURL, instanceID: instance.instanceID}, nil
}

func (m *matcher) findGameLobby(ctx context.Context, game string) (*gameLobby, error) {
	m.state.Lock()
	defer m.state.Unlock()

	lobby, ok := m.gameMatches[game]
	if !ok {
		return nil, errors.New("game not found")
	}
	return lobby, nil
}
