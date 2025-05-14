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
	match, err := v.core.findMatchInstance(ctx, in.Game)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "ClientCompetition#QuickMatch done", "Game", in.Game, "match", match)
	return &wire.QuickMatchOut{
		MatchURL: match.instanceURL,
		UUID:     match.instanceID,
	}, nil
}
