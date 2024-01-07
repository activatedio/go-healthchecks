package config_test

import (
	"github.com/activatedio/go-healthchecks/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewConfig(t *testing.T) {

	type s struct {
		arrange func() string
		assert  func(got *config.Config, err error)
	}

	cases := map[string]s{
		"invalid path": {
			arrange: func() string {
				return "invalid"
			},
			assert: func(got *config.Config, err error) {
				assert.Nil(t, got)
				assert.EqualError(t, err, "open invalid: no such file or directory")
			},
		},
		"valid": {
			arrange: func() string {
				return "./testdata/config.yaml"
			},
			assert: func(got *config.Config, err error) {
				assert.Nil(t, err)
				assert.Equal(t, &config.Config{
					Checks: map[string]*config.Check{
						"a": {
							Type: "typeA",
							Config: map[string]any{
								"a1": "a1value",
								"a2": "a2value",
							},
						},
						"b": {
							Type: "typeB",
							Config: map[string]any{
								"b1": 1234,
								"b2": []any{
									"a", "b", "c",
								},
							},
						},
					},
				}, got)
			},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {

			config.ConfigFilePath = v.arrange()
			v.assert(config.NewConfig())
		})
	}
}
