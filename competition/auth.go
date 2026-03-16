// Package competition implements a multiplayer game competition system.
package competition

import "context"

type Auth interface {
	GetUserID(ctx context.Context) (string, error)
}
