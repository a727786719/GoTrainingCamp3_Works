package app

import (
	"context"
)

type App interface {
	Name() string
	Version() string
	Run(ctx context.Context) error
	Stop() error
}
