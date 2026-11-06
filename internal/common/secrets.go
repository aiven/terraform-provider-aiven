package common

import (
	"os"
	"strconv"
)

// IncludeSecretsEnv set to false makes the reads return redacted secrets.
const IncludeSecretsEnv = "AIVEN_INCLUDE_SECRETS"

// IncludeSecrets is the include_secrets query parameter: true unless IncludeSecretsEnv is false.
func IncludeSecrets() [2]string {
	v, err := strconv.ParseBool(os.Getenv(IncludeSecretsEnv))
	return [2]string{"include_secrets", strconv.FormatBool(err != nil || v)}
}
