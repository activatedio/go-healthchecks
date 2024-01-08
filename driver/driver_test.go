package driver_test

import (
	"context"
	"errors"
	"github.com/activatedio/go-healthchecks/checks"
	"github.com/activatedio/go-healthchecks/config"
	"github.com/activatedio/go-healthchecks/driver"
	"github.com/stretchr/testify/assert"
	"strings"
	"sync"
	"testing"
	"time"
)

type result struct {
	status *checks.Status
	err    error
}

type StubConfig struct {
	Pattern string
}

type stubChecker struct {
	index   int
	results []*result
	lock    sync.Mutex
}

func (s *stubChecker) ZeroConfig() any {
	return &StubConfig{}
}

func (s *stubChecker) Check(ctx context.Context, config any) (*checks.Status, error) {

	_config := config.(*StubConfig)

	if _config.Pattern == "u*" {
		return checks.StatusUnhealthy, nil
	}
	if _config.Pattern == "un*" {
		return checks.StatusUnknown, nil
	}

	s.lock.Lock()
	defer func() {
		s.index++
		s.lock.Unlock()
	}()

	if s.results == nil {
		for _, p := range strings.Split(_config.Pattern, ",") {
			switch p {
			case "u":
				s.results = append(s.results, &result{
					status: checks.StatusUnhealthy,
				})
			case "h":
				s.results = append(s.results, &result{
					status: checks.StatusHealthy,
				})
			case "e":
				s.results = append(s.results, &result{
					status: checks.StatusUnknown,
					err:    errors.New("test-error"),
				})
			}
		}
	}

	if len(s.results) <= s.index {
		panic("index out of bounds")
	}
	r := s.results[s.index]
	return r.status, r.err
}

type delayChecker struct {
}

type DelayConfig struct {
	Seconds int
}

func (d *delayChecker) ZeroConfig() any {
	return &DelayConfig{}
}

func (d *delayChecker) Check(ctx context.Context, config any) (*checks.Status, error) {
	_config := config.(*DelayConfig)
	time.Sleep(time.Duration(_config.Seconds) * time.Second)
	return checks.StatusHealthy, nil
}

func TestDriver_Run(t *testing.T) {

	type s struct {
		arrange func(ctx context.Context) (context.Context, map[string]func() checks.Checker, *config.Config, []driver.RunOption)
		assert  func(got *checks.Status, err error)
	}

	registry := map[string]func() checks.Checker{
		"stub": func() checks.Checker {
			return &stubChecker{}
		},
		"delay": func() checks.Checker {
			return &delayChecker{}
		},
	}

	cases := map[string]s{
		"timeout": {
			arrange: func(ctx context.Context) (context.Context, map[string]func() checks.Checker, *config.Config, []driver.RunOption) {
				return ctx, registry, &config.Config{
					Checks: map[string]*config.Check{
						"a": {
							Type: "delay",
							Config: map[string]any{
								"seconds": 2,
							},
						},
						"b": {
							Type: "delay",
							Config: map[string]any{
								"seconds": 2,
							},
						},
					},
				}, []driver.RunOption{driver.WithTimeout(2)}
			},
			assert: func(got *checks.Status, err error) {
				assert.Equal(t, checks.StatusUnknown, got)
				assert.EqualError(t, err, "timeout after 2 seconds")
			},
		},
		"healthy": {
			arrange: func(ctx context.Context) (context.Context, map[string]func() checks.Checker, *config.Config, []driver.RunOption) {
				return ctx, registry, &config.Config{
					Checks: map[string]*config.Check{
						"a": {
							Type: "stub",
							Config: map[string]any{
								"pattern": "u,u,h,u",
							},
						},
						"b": {
							Type: "stub",
							Config: map[string]any{
								"pattern": "u,u,h,u",
							},
						},
					},
				}, []driver.RunOption{driver.WithTimeout(120)}
			},
			assert: func(got *checks.Status, err error) {
				assert.Nil(t, err)
				assert.Equal(t, checks.StatusHealthy, got)
			},
		},
		"unhealthy": {
			arrange: func(ctx context.Context) (context.Context, map[string]func() checks.Checker, *config.Config, []driver.RunOption) {
				return ctx, registry, &config.Config{
					Checks: map[string]*config.Check{
						"a": {
							Type: "stub",
							Config: map[string]any{
								"pattern": "u,u,h",
							},
						},
						"b": {
							Type: "stub",
							Config: map[string]any{
								"pattern": "u*",
							},
						},
					},
				}, []driver.RunOption{driver.WithTimeout(5)}
			},
			assert: func(got *checks.Status, err error) {
				assert.EqualError(t, err, "timeout after 5 seconds")
				assert.Equal(t, checks.StatusUnknown, got)
			},
		},
		"unkonwn": {
			arrange: func(ctx context.Context) (context.Context, map[string]func() checks.Checker, *config.Config, []driver.RunOption) {
				return ctx, registry, &config.Config{
					Checks: map[string]*config.Check{
						"a": {
							Type: "stub",
							Config: map[string]any{
								"pattern": "u,u,h",
							},
						},
						"b": {
							Type: "stub",
							Config: map[string]any{
								"pattern": "un*",
							},
						},
					},
				}, []driver.RunOption{driver.WithTimeout(5)}
			},
			assert: func(got *checks.Status, err error) {
				assert.EqualError(t, err, "timeout after 5 seconds")
				assert.Equal(t, checks.StatusUnknown, got)
			},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {

			ctx := context.Background()

			ctx, r, c, o := v.arrange(ctx)

			unit := driver.NewDriver(driver.DriverParams{
				Registry: r,
			})

			v.assert(unit.Run(ctx, c, o...))
		})
	}
}
