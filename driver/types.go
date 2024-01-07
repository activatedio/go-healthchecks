package driver

import (
	"context"
	"github.com/activatedio/go-healthchecks/checks"
	"github.com/activatedio/go-healthchecks/config"
)

type Logger func(msg string)

type Driver interface {
	Run(ctx context.Context, c *config.Config, opts ...RunOption) (checks.Status, error)
}
