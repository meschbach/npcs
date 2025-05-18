package competition

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/fx"
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
	//todo: consider replacing with len(players)
	playerCount int
	instanceURL string
	instanceID  string
	players     []string
	winner      string
	gameID      string
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

func (m *matcher) registerInstance(ctx context.Context, game, id, instanceURL string) (string, error) {
	slog.InfoContext(ctx, "matcher.registerInstance", "Game", game, "instanceURL", instanceURL)
	lobby, err := m.findGameLobby(ctx, game)
	if err != nil {
		return "", err
	}

	lobby.state.Lock()
	defer lobby.state.Unlock()

	lobby.available = append(lobby.available, &gameInstanceLobby{
		phase:        waitingForPlayers,
		desiredCount: 2,
		playerCount:  0,
		instanceURL:  instanceURL,
		instanceID:   id,
		gameID:       game,
	})
	return id, nil
}

func (m *matcher) findMatchInstance(ctx context.Context, game, playerID string) (*matchedGame, error) {
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
	instance.players = append(instance.players, playerID)
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
	lobby.waiting.Broadcast()
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

func (m *matcher) gameCompleted(ctx context.Context, game, instanceID, winner string) error {
	lobby, err := m.findGameLobby(ctx, game)
	if err != nil {
		return err
	}

	lobby.state.Lock()
	defer lobby.state.Unlock()
	matched, stillPlaying := fx.Split(lobby.playing, func(e *gameInstanceLobby) bool {
		return e.instanceID == instanceID
	})
	slog.InfoContext(ctx, "matcher.gameCompleted__found playing", "Game", game, "instanceID", instanceID, "stillPlaying", stillPlaying, "matched", matched)
	if len(matched) != 1 {
		return errors.New("no such match")
	}
	lobby.playing = stillPlaying
	lobby.completed = append(lobby.completed, matched...)

	match := matched[0]
	match.phase = completed
	//todo: check winner is in the set of players
	match.winner = winner
	lobby.waiting.Broadcast()
	return nil
}

func (m *matcher) findAllGamesForPlayer(ctx context.Context, playerID string) []*gameInstanceLobby {
	m.state.Lock()
	defer m.state.Unlock()
	var out []*gameInstanceLobby

	for _, game := range m.gameMatches {
		game.state.Lock()
		finished := game.completed
		for _, match := range finished {
			if match.phase == completed {
				matching := fx.Filter(match.players, func(e string) bool {
					return e == playerID
				})
				if len(matching) >= 1 {
					out = append(out, match)
				}
			}
		}
		game.state.Unlock()
	}
	return out
}
