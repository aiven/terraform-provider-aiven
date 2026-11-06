package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncludeSecrets(t *testing.T) {
	cases := map[string]string{
		"":        "true",
		"true":    "true",
		"invalid": "true",
		"false":   "false",
		"0":       "false",
	}
	for env, want := range cases {
		t.Run(env, func(t *testing.T) {
			t.Setenv(IncludeSecretsEnv, env)
			assert.Equal(t, [2]string{"include_secrets", want}, IncludeSecrets())
		})
	}
}
