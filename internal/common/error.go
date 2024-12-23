package common

import (
	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
)

// IsCritical returns true if the given error is critical
func IsCritical(err error) bool {
	return !(err == nil || aiven.IsNotFound(err) || avngen.IsNotFound(err))
}
