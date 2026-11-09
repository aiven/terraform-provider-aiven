package adapter

import (
	"errors"
	"fmt"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
)

// ErrorRule maps an operation error. Status and Contains match an avngen.Error, without them the rule matches any error.
type ErrorRule struct {
	Status   int
	Contains string
	// As wraps the matched error, nil drops it.
	As error
}

func (r ErrorRule) matches(err error) bool {
	if r.Status == 0 && r.Contains == "" {
		return true
	}
	var e avngen.Error
	return errors.As(err, &e) && (r.Status == 0 || e.Status == r.Status) && strings.Contains(e.Message, r.Contains)
}

// MapError applies the first rule that matches err.
func MapError(err error, rules ...ErrorRule) error {
	if err == nil {
		return nil
	}
	for _, r := range rules {
		if !r.matches(err) {
			continue
		}
		if r.As == nil {
			return nil
		}
		return fmt.Errorf("%w: %w", r.As, err)
	}
	return err
}
