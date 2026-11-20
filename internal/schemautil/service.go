package schemautil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/common"
)

// defaultTimeout is the default timeout for service operations. This is not a const because it can be changed during
// compile time with -ldflags "-X github.com/aiven/terraform-provider-aiven/internal/schemautil.defaultTimeout=30".
var defaultTimeout time.Duration = 20

func DefaultResourceTimeouts() *schema.ResourceTimeout {
	return &schema.ResourceTimeout{
		Create: schema.DefaultTimeout(defaultTimeout * time.Minute),
		Update: schema.DefaultTimeout(defaultTimeout * time.Minute),
		Delete: schema.DefaultTimeout(defaultTimeout * time.Minute),
		// DEPRECATED: Default timeout is deprecated.
		// The Plugin Framework does not support the Default timeout field.
		// This field will be removed in a future major version.
		// See: https://developer.hashicorp.com/terraform/plugin/framework/resources/timeouts
		Default: schema.DefaultTimeout(defaultTimeout * time.Minute),
		Read:    schema.DefaultTimeout(defaultTimeout * time.Minute),
	}
}

// GetDefaultTimeout returns the default timeout for service operations.
func GetDefaultTimeout() time.Duration {
	return defaultTimeout * time.Minute
}

// TODO: promote these to a `type ServiceType string` enum once the remaining
// SDKv2 service resources are migrated to the Plugin Framework. We attempted
// the retype but SDKv2 resource schemas and user-config glue pass the raw
// string through many APIs (`map[string]*schema.Schema` keys, `string + "_user_config"`
// concatenation, `*schema.ResourceData.Set/Get` keys), so the cascade of
// `string(...)` conversions made the diff too large to justify at this stage.
// Revisit after the plugin-framework migration shrinks that surface.
const (
	ServiceTypePG               = "pg"
	ServiceTypeCassandra        = "cassandra"
	ServiceTypeOpenSearch       = "opensearch"
	ServiceTypeGrafana          = "grafana"
	ServiceTypeMySQL            = "mysql"
	ServiceTypeKafka            = "kafka"
	ServiceTypeKafkaConnect     = "kafka_connect"
	ServiceTypeKafkaMirrormaker = "kafka_mirrormaker"
	ServiceTypeFlink            = "flink"
	ServiceTypeClickhouse       = "clickhouse"
	ServiceTypeThanos           = "thanos"
	ServiceTypeValkey           = "valkey"
)

// validServiceTypes returns the set of known service type codes.
// Mirrors the ServiceType* constants above and the `requireServiceReady`
// enum in `definitions/.schema.yml`; keep them in sync when adding a
// new service type.
func validServiceTypes() map[string]bool {
	return map[string]bool{
		ServiceTypePG:               true,
		ServiceTypeCassandra:        true,
		ServiceTypeOpenSearch:       true,
		ServiceTypeGrafana:          true,
		ServiceTypeMySQL:            true,
		ServiceTypeKafka:            true,
		ServiceTypeKafkaConnect:     true,
		ServiceTypeKafkaMirrormaker: true,
		ServiceTypeFlink:            true,
		ServiceTypeClickhouse:       true,
		ServiceTypeThanos:           true,
		ServiceTypeValkey:           true,
	}
}

// DefaultServiceUsername returns the default admin username managed by Aiven for a given service type.
// Different services use different default usernames (e.g., avnadmin, default, etc.).
func DefaultServiceUsername(serviceType string) string {
	switch serviceType {
	case ServiceTypeValkey:
		return "default"
	default:
		return "avnadmin" // most services use avnadmin
	}
}

func ResourceReadHandleNotFound(err error, d ResourceData) error {
	if err != nil && common.IsUnknownResource(err) && !d.IsNewResource() {
		d.SetId("")
		return nil
	}
	return err
}

const redactedSubstr = `\u003credacted\u003e`

var errContainsRedactedCreds = fmt.Errorf("unexpected redacted credentials")

// ContainsRedactedCreds looks for redactedSubstr in the given config
func ContainsRedactedCreds(config map[string]any) error {
	b, err := json.Marshal(&config)
	if err != nil {
		return err
	}

	if bytes.Contains(b, []byte(redactedSubstr)) {
		return errContainsRedactedCreds
	}
	return nil
}

var (
	// serviceCheck caches ServiceGet per (project, serviceName). It is
	// seeded on service creation and update, and dropped on deletion. The value is the
	// lookup result, not a verdict: a (project,serviceName) is one service
	// with one type, so readiness is evaluated per caller against the
	// shared cached shape.
	serviceCheck           DoOnce[*service.ServiceGetOut]
	ErrServicePoweredOff   = fmt.Errorf("the service is powered off")
	ErrServiceTypeMismatch = fmt.Errorf("service type mismatch")
	ErrInvalidServiceType  = fmt.Errorf("invalid service type")
)

// ServiceReadyForget drops the cached ServiceGet for (project, serviceName).
func ServiceReadyForget(project, serviceName string) {
	serviceCheck.Forget(project, serviceName)
}

// ServiceReadyStore caches the ServiceGet result for (project, serviceName).
func ServiceReadyStore(s *service.ServiceGetOut, project, serviceName string) {
	serviceCheck.Store(s, project, serviceName)
}

// CheckServiceReady verifies the target service exists, is powered on,
// and has the expected service type before performing operations that
// require it. Some operations like database management need the service
// to be running. For example, `ServiceDatabaseList` returns a generic
// 503 error when the service is powered off:
// "503: An error occurred. Please try again later."
// This function returns a clearer error by explicitly prechecking.
// serviceTypes lists the accepted service type codes (see
// validServiceTypes). An empty list or an unknown code returns
// ErrInvalidServiceType; a ServiceGetOut.ServiceType outside the list
// returns ErrServiceTypeMismatch; a service in POWEROFF returns
// ErrServicePoweredOff.
func CheckServiceReady(ctx context.Context, client avngen.Client, project, serviceName string, serviceTypes ...string) error {
	if len(serviceTypes) == 0 {
		return fmt.Errorf("%w: no service type given", ErrInvalidServiceType)
	}
	for _, t := range serviceTypes {
		if !validServiceTypes()[t] {
			return fmt.Errorf("%w: %q", ErrInvalidServiceType, t)
		}
	}

	s, err := serviceCheck.Do(func() (*service.ServiceGetOut, error) {
		return client.ServiceGet(ctx, project, serviceName)
	}, project, serviceName)
	if err != nil {
		if avngen.IsNotFound(err) {
			// A 404 may be eventual consistency for a service created
			// elsewhere, so let the next caller retry.
			serviceCheck.Forget(project, serviceName)
		}
		return err
	}
	if !slices.Contains(serviceTypes, s.ServiceType) {
		expected := strings.Join(lo.Map(serviceTypes, func(t string, _ int) string { return strconv.Quote(t) }), ", ")
		if len(serviceTypes) > 1 {
			expected = "one of " + expected
		}
		return fmt.Errorf(
			"%w: service %s/%s is of type %q, expected %s",
			ErrServiceTypeMismatch, project, serviceName, s.ServiceType, expected,
		)
	}
	if s.State == service.ServiceStateTypePoweroff {
		return ErrServicePoweredOff
	}
	return nil
}
