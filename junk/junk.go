// Package junk provides experimental utilities and prototypes.
package junk

import "context"

type Component interface {
	Start(ctx context.Context) error
	Stop(shutdownCtx context.Context) error
}
