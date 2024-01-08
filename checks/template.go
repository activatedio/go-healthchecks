package checks

import (
	"context"
)

type CheckerFunc[C any] func(ctx context.Context, config *C) (*Status, error)

type template[C any] struct {
	checker CheckerFunc[C]
}

func (c *template[C]) ZeroConfig() any {
	return new(C)
}

func (c *template[C]) Check(ctx context.Context, config any) (*Status, error) {
	return c.checker(ctx, config.(*C))
}

func NewFromTemplate[C any](checkerFunc CheckerFunc[C]) Checker {
	return &template[C]{
		checker: checkerFunc,
	}
}
