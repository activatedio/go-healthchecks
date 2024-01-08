package checks

import "context"

type Status struct {
	slug string
}

func (s *Status) String() string {
	return s.slug
}

var (
	StatusUnknown   = &Status{""}
	StatusUnhealthy = &Status{"unhealthy"}
	StatusHealthy   = &Status{"healthy"}
)

type CheckResult struct {
}

// Type C - config type
type Checker interface {
	ZeroConfig() any
	Check(ctx context.Context, config any) (*Status, error)
}
