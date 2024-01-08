package driver

import (
	"context"
	"errors"
	"fmt"
	"github.com/activatedio/go-healthchecks/checks"
	"github.com/activatedio/go-healthchecks/config"
	"github.com/mitchellh/mapstructure"
	"sync"
	"time"
)

type driver struct {
	registry map[string]func() checks.Checker
	logger   Logger
}

type Options struct {
	timeout int
	logger  Logger
}

func defaultOptions() *Options {
	return &Options{
		timeout: 120,
		logger: func(msg string) {
			fmt.Println(msg)
		},
	}
}

type RunOption func(o *Options)

func WithLogger(l Logger) RunOption {
	return func(o *Options) {
		o.logger = l
	}
}

func WithTimeout(timeout int) RunOption {
	return func(o *Options) {
		o.timeout = timeout
	}
}

type result struct {
	s *checks.Status
	e error
}

func (d *driver) Run(ctx context.Context, c *config.Config, opts ...RunOption) (*checks.Status, error) {

	o := defaultOptions()

	for _, _o := range opts {
		_o(o)
	}

	d.logger = o.logger

	ch := make(chan result, 1)

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, time.Second*time.Duration(o.timeout))
	defer cancel()

	go d.doRun(ctx, c, ch)

	select {
	case <-ctx.Done():
		return checks.StatusUnknown, errors.New(fmt.Sprintf("timeout after %d seconds", o.timeout))
	case r := <-ch:
		return r.s, r.e
	}

}

type run struct {
	results map[string]*result
	lock    sync.Mutex
}

func (r *run) Set(key string, value *result) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.results[key] = value
}

func (d *driver) doRun(ctx context.Context, c *config.Config, ch chan result) {

	type e struct {
		c   checks.Checker
		cfg any
	}

	// Build and configure drivers
	chks := map[string]e{}

	for k, v := range c.Checks {
		chkf, ok := d.registry[v.Type]
		if !ok {
			ch <- result{
				s: checks.StatusUnknown,
				e: errors.New("unrecognized checker type: " + v.Type),
			}
			return
		}
		chk := chkf()
		cfg := chk.ZeroConfig()
		err := mapstructure.Decode(v.Config, cfg)
		if err != nil {
			ch <- result{
				e: err,
			}
		}
		chks[k] = e{
			c:   chk,
			cfg: cfg,
		}
	}

	var wg sync.WaitGroup

	_run := &run{
		results: map[string]*result{},
	}

	for k, v := range chks {

		name := k
		chk := v

		_run.Set(k, &result{
			s: checks.StatusUnknown,
		})

		wg.Add(1)

		go func() {
			defer func() {
				wg.Done()
			}()
			for {
				d.logger(fmt.Sprintf("checking %s", name))
				status, err := chk.c.Check(ctx, chk.cfg)
				d.logger(fmt.Sprintf("checked %s, result: %s, error: %s", name, status, err))
				_run.Set(name, &result{
					s: status,
					e: err,
				})
				if err != nil || status == checks.StatusHealthy {
					break
				}
				time.Sleep(time.Second)
			}
		}()

	}

	wg.Wait()

	var hasUnhealthy bool
	var hasUnknown bool
	var hasHealthy bool

	if len(_run.results) == 0 {
		ch <- result{
			s: checks.StatusUnknown,
		}
		return
	}

	for _, v := range _run.results {
		if v.e != nil {
			ch <- result{
				s: checks.StatusUnknown,
				e: v.e,
			}
			return
		}
		switch v.s {
		case checks.StatusUnknown:
			hasUnknown = true
		case checks.StatusHealthy:
			hasHealthy = true
		case checks.StatusUnhealthy:
			hasUnhealthy = true
		default:
			panic("unrecognized status")
		}
	}

	if hasUnknown {
		ch <- result{
			s: checks.StatusUnknown,
		}
		return
	}
	if hasUnhealthy {
		ch <- result{
			s: checks.StatusUnhealthy,
		}
		return
	}
	if hasHealthy {
		ch <- result{
			s: checks.StatusHealthy,
		}
		return
	}

	panic("unexpected condition")

}

type DriverParams struct {
	Registry map[string]func() checks.Checker
}

func NewDriver(params DriverParams) Driver {
	return &driver{
		registry: params.Registry,
	}
}
