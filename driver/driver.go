package driver

import (
	"context"
	"errors"
	"fmt"
	"github.com/activatedio/go-healthchecks/checks"
	"github.com/activatedio/go-healthchecks/config"
	"time"
)

type driver struct {
	registry map[string]checks.Checker
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
	s checks.Status
	e error
}

func (d *driver) Run(ctx context.Context, c *config.Config, opts ...RunOption) (checks.Status, error) {

	o := defaultOptions()

	for _, _o := range opts {
		_o(o)
	}

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

func (d *driver) doRun(ctx context.Context, c *config.Config, ch chan result) {

	// Build and configure drivers
	chks := map[string]checks.Checker{}

	for k, v := range c.Checks {
		chk, ok := d.registry[v.Type]
		if !ok {
			ch <- result{
				s: checks.StatusUnknown,
				e: errors.New("unrecognized checker type: " + v.Type),
			}
			return
		}
		chks[k] = chk
	}

}

type DriverParams struct {
	Registry map[string]checks.Checker
}

func NewDriver(params DriverParams) Driver {
	return &driver{
		registry: params.Registry,
	}
}
