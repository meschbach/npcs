package network

import (
	"context"
	"errors"
	"github.com/meschbach/npcs/t3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"sync"
)

type PushService struct {
	UnimplementedT3PushServer
	Router *PushServiceRouter
}

func NewPush() *PushService {
	return &PushService{
		Router: &PushServiceRouter{
			changes:       sync.Mutex{},
			playerByToken: make(map[string]*pushTokenSlot),
		},
	}
}

var Unauthenticated = status.New(codes.Unauthenticated, "no authorization").Err()
var Unauthorized = status.New(codes.Unauthenticated, "unauthorized").Err()

func (p *PushService) ConnectToGame(in *JoinGameIn, out grpc.ServerStreamingServer[T3PushEvent]) error {
	md, hasMetaData := metadata.FromIncomingContext(out.Context())
	if !hasMetaData {
		return Unauthenticated
	}
	auth, hasAuth := md["authorization"]
	if !hasAuth || len(auth) != 1 {
		return Unauthenticated
	}

	parts := strings.SplitN(auth[0], " ", 2)
	if len(parts) != 2 {
		return Unauthorized
	}

	game, has := p.Router.hasGame(parts[1], in.GameID)
	if !has {
		return Unauthorized
	}

	if err := out.Send(&T3PushEvent{
		Initial: &JoinGameOut{
			YourPlayer: game.playerID,
		},
	}); err != nil {
		return err
	}

	for {
		ctx := out.Context()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-game.outgoing:
			if !ok { //closed
				return nil
			}
			if err := out.Send(event); err != nil {
				return err
			}
		}
	}
}

func (p *PushService) Move(ctx context.Context, move *PushMoveIn) (*PushMoveOut, error) {
	md, hasMetaData := metadata.FromIncomingContext(ctx)
	if !hasMetaData {
		return nil, Unauthenticated
	}
	auth, hasAuth := md["authorization"]
	if !hasAuth || len(auth) != 1 {
		return nil, Unauthenticated
	}

	parts := strings.SplitN(auth[0], " ", 2)
	if len(parts) != 2 {
		return nil, Unauthorized
	}

	game, has := p.Router.hasGame(parts[1], move.GameID)
	if !has {
		return nil, Unauthorized
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case game.moves <- t3.Move{
		Row:    int(move.Move.Row),
		Column: int(move.Move.Column),
	}:
		return nil, nil
	}
}

type PushServiceRouter struct {
	changes       sync.Mutex
	playerByToken map[string]*pushTokenSlot
}

func newPushServiceRouter() *PushServiceRouter {
	return &PushServiceRouter{
		changes:       sync.Mutex{},
		playerByToken: make(map[string]*pushTokenSlot),
	}
}

func (p *PushServiceRouter) Register(token string, gameID string, side int64) (t3.Player, error) {
	p.changes.Lock()
	defer p.changes.Unlock()
	slot, has := p.playerByToken[token]
	if !has {
		slot = &pushTokenSlot{
			changes: sync.Mutex{},
			games:   make(map[string]*pushGame),
		}
		p.playerByToken[token] = slot
	}

	slot.changes.Lock()
	defer slot.changes.Unlock()
	_, hadGame := slot.games[gameID]
	if hadGame {
		return nil, errors.New("game already exists")
	}
	game := &pushGame{
		playerID: side,
		outgoing: make(chan *T3PushEvent, 8),
		moves:    make(chan t3.Move, 1),
	}
	slot.games[gameID] = game
	return game, nil
}

func (p *PushServiceRouter) GameComplete(ctx context.Context, token, gameID string, winner int) error {
	p.changes.Lock()
	defer p.changes.Unlock()
	slot, has := p.playerByToken[token]
	if !has {
		return nil
	}
	slot.changes.Lock()
	defer slot.changes.Unlock()
	game, ok := slot.games[gameID]
	if !ok {
		return nil
	}
	return game.concluded(ctx, winner)
}

func (p *PushServiceRouter) Remove(token string) {
	p.changes.Lock()
	defer p.changes.Unlock()
	delete(p.playerByToken, token)
}

func (p *PushServiceRouter) hasGame(token, gameID string) (*pushGame, bool) {
	//todo: holds on to lock longer than needed when looking up the game
	p.changes.Lock()
	defer p.changes.Unlock()

	slot, has := p.playerByToken[token]
	if !has {
		return nil, false
	}
	//no longer need p.changes

	slot.changes.Lock()
	defer slot.changes.Unlock()
	game, hasGame := slot.games[gameID]
	if !hasGame {
		return nil, false
	}
	return game, true
}

type pushTokenSlot struct {
	changes sync.Mutex
	//gameID -> game
	games map[string]*pushGame
}

type pushGame struct {
	playerID int64
	outgoing chan *T3PushEvent
	moves    chan t3.Move
}

func (p *pushGame) NextPlay(ctx context.Context) (t3.Move, error) {
	//notify the client we would like a move
	select {
	case <-ctx.Done():
		return t3.Move{}, ctx.Err()
	case p.outgoing <- &T3PushEvent{
		DoTurn: &PlayerTurn{},
	}:
	}

	//Wait for a reply
	select {
	case m := <-p.moves:
		return m, nil
	case <-ctx.Done():
		return t3.Move{}, ctx.Err()
	}
}

func (p *pushGame) PushHistory(ctx context.Context, move t3.Move) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.outgoing <- &T3PushEvent{Move: &MoveMadeIn{
		Player: int64(move.Player),
		Row:    int64(move.Row),
		Column: int64(move.Column),
	}}:
		return nil
	}
}

func (p *pushGame) concluded(ctx context.Context, winner int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.outgoing <- &T3PushEvent{
		Conclusion: &ConclusionIn{
			Stalemate: winner == 0,
			Winner:    int64(winner),
			Withdraw:  false,
		},
	}:
		close(p.outgoing)
		return nil
	}
}
