package driver_test

import (
	"context"
	"github.com/activatedio/go-healthchecks/checks"
	"github.com/activatedio/go-healthchecks/config"
	"github.com/activatedio/go-healthchecks/driver"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDriver_Run(t *testing.T) {

	type s struct {
		arrange func() (map[string]checks.Checker, *config.Config, []driver.RunOption)
		assert  func(got checks.Status, err error)
	}

	cases := map[string]s{
		"timeout": {
			arrange: func() (map[string]checks.Checker, *config.Config, []driver.RunOption) {
				return map[string]checks.Checker{}, &config.Config{}, []driver.RunOption{driver.WithTimeout(1)}
			},
			assert: func(got checks.Status, err error) {
				assert.Equal(t, checks.StatusUnknown, got)
				assert.EqualError(t, err, "timeout after 1 seconds")
			},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {

			r, c, o := v.arrange()

			ctx := context.Background()

			unit := driver.NewDriver(driver.DriverParams{
				Registry: r,
			})

			v.assert(unit.Run(ctx, c, o...))
		})
	}
}
