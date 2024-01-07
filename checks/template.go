package checks

import (
	"context"
)

type CheckerFunc[C any] func(ctx context.Context, config any) (*Status, error)

type checkerTemplate[C any] struct {
	checker CheckerFunc[C]
}

func (c *checkerTemplate[C]) ZeroConfig() any {
	return new(C)
}

func (c *checkerTemplate[C]) Check(ctx context.Context, config any) (*Status, error) {
	return c.checker(ctx, config)
}

func NewCheckerTemplate[C any](checkerFunc CheckerFunc[C]) Checker {
	return &checkerTemplate[C]{
		checker: checkerFunc,
	}
}
