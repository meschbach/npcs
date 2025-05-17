package competition

import (
	"context"
	"github.com/meschbach/npcs/competition/wire"
	"log/slog"
)

// clientCompetitionService is a gRPC server implementation for managing player competitions and game results.
// It embeds wire.UnimplementedCompetitionV1Server to ensure forward compatibility with the CompetitionV1 interface.
type clientCompetitionService struct {
	wire.UnimplementedCompetitionV1Server
	// auth is used to authenticate and retrieve the user information for managing persistent players.
	auth Auth
	core *matcher
}

func (v *clientCompetitionService) QuickMatch(ctx context.Context, in *wire.QuickMatchIn) (*wire.QuickMatchOut, error) {
	slog.InfoContext(ctx, "ClientCompetition#QuickMatch", "Game", in.Game)
	match, err := v.core.findMatchInstance(ctx, in.Game, in.PlayerName)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "ClientCompetition#QuickMatch done", "Game", in.Game, "match", match)
	return &wire.QuickMatchOut{
		MatchURL: match.instanceURL,
		UUID:     match.instanceID,
	}, nil
}

func (v *clientCompetitionService) GetHistory(ctx context.Context, in *wire.RecordIn) (*wire.RecordOut, error) {
	out := &wire.RecordOut{}
	for _, game := range v.core.findAllGamesForPlayer(ctx, in.ForPlayer) {
		out.Games = append(out.Games, &wire.CompletedGame{
			Game:       game.gameID,
			InstanceID: game.instanceID,
			Players:    game.players,
			Winner:     game.winner,
			Start:      nil,
			End:        nil,
		})
	}
	return out, nil
}
