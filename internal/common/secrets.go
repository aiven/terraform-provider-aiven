package common

import (
	"os"
	"strconv"
)

// IncludeSecretsEnv set to false makes the reads return redacted secrets.
const IncludeSecretsEnv = "AIVEN_INCLUDE_SECRETS"

// ReadSecrets is true unless IncludeSecretsEnv is false.
func ReadSecrets() bool {
	v, err := strconv.ParseBool(os.Getenv(IncludeSecretsEnv))
	return err != nil || v
}

// IncludeSecrets is the include_secrets query parameter.
func IncludeSecrets() [2]string {
	return [2]string{"include_secrets", strconv.FormatBool(ReadSecrets())}
}
