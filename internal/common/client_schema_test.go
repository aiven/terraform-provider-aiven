package common_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/provider"
)

// uriField holds a connection URI: its password is masked, not the whole value.
var uriField = regexp.MustCompile(`(uri|uris|url|bouncer)$`)

// unmaskedSensitive are sensitive fields that aren't secrets in the API bodies.
var unmaskedSensitive = map[string]string{
	"access_cert":      "a certificate",
	"ca_cert":          "a certificate",
	"database_name":    "connection parameter",
	"dbname":           "connection parameter",
	"host":             "connection parameter",
	"host_ports":       "connection parameter",
	"max_connections":  "connection parameter",
	"password_wo":      `sent as "password"`,
	"port":             "connection parameter",
	"service_password": `returned as "password"`,
	"sslmode":          "connection parameter",
	"user":             "connection parameter",
	"username":         "connection parameter",
}

// sensitiveField is an example path of a sensitive field and its value in JSON.
type sensitiveField struct {
	path, value string
}

const (
	jsonString = `"secret-value"`
	jsonList   = `["secret-value"]`
	jsonMap    = `{"name": "secret-value"}`
	jsonURI    = `"postgres://avnadmin:secret-value@host:1234/defaultdb"`
	jsonURIs   = `["postgres://avnadmin:secret-value@host:1234/defaultdb"]`
)

// TestSecretValuesMaskSensitiveFields fails when a new sensitive field would be logged in clear text:
// mask it in maskBody or, if it isn't a secret in the API, add it to unmaskedSensitive.
func TestSecretValuesMaskSensitiveFields(t *testing.T) {
	t.Setenv(util.AivenEnableBeta, "true")
	fields := make(map[string]sensitiveField)

	p, err := provider.Provider("dev")
	require.NoError(t, err)
	for name, r := range p.ResourcesMap {
		sensitiveSDKFields(name, r.Schema, fields)
	}

	ctx := context.Background()
	for _, newResource := range plugin.New("dev").Resources(ctx) {
		r := newResource()
		var meta resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "aiven"}, &meta)
		var rsp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &rsp)
		sensitiveFrameworkFields(meta.TypeName, rsp.Schema.Attributes, rsp.Schema.Blocks, fields)
	}

	require.NotEmpty(t, fields)
	for name, f := range fields {
		if _, ok := unmaskedSensitive[name]; ok {
			continue
		}
		value := f.value
		if uriField.MatchString(name) {
			value = map[string]string{jsonString: jsonURI, jsonList: jsonURIs}[value]
		}
		masked := common.MaskBody(fmt.Appendf(nil, `{"name": "app", %q: %s}`, name, value))
		assert.NotContains(t, masked, "secret-value", "%s isn't masked in the API logs", f.path)
		assert.Contains(t, masked, `"name":"app"`)
	}
}

// sensitiveSDKFields maps the names of sensitive leaf fields to an example.
func sensitiveSDKFields(path string, m map[string]*sdkschema.Schema, fields map[string]sensitiveField) {
	for name, s := range m {
		if r, ok := s.Elem.(*sdkschema.Resource); ok {
			sensitiveSDKFields(path+"."+name, r.Schema, fields)
			continue
		}
		if !s.Sensitive {
			continue
		}
		value := jsonString
		switch s.Type {
		case sdkschema.TypeList, sdkschema.TypeSet:
			value = jsonList
		case sdkschema.TypeMap:
			value = jsonMap
		}
		fields[name] = sensitiveField{path: path + "." + name, value: value}
	}
}

// sensitiveFrameworkFields maps the names of sensitive leaf fields to an example.
func sensitiveFrameworkFields(
	path string,
	attrs map[string]schema.Attribute,
	blocks map[string]schema.Block,
	fields map[string]sensitiveField,
) {
	for name, a := range attrs {
		p := path + "." + name
		value := jsonString
		switch n := a.(type) {
		case schema.ListNestedAttribute:
			sensitiveFrameworkFields(p, n.NestedObject.Attributes, nil, fields)
			continue
		case schema.SetNestedAttribute:
			sensitiveFrameworkFields(p, n.NestedObject.Attributes, nil, fields)
			continue
		case schema.MapNestedAttribute:
			sensitiveFrameworkFields(p, n.NestedObject.Attributes, nil, fields)
			continue
		case schema.SingleNestedAttribute:
			sensitiveFrameworkFields(p, n.Attributes, nil, fields)
			continue
		case schema.ListAttribute, schema.SetAttribute:
			value = jsonList
		case schema.MapAttribute:
			value = jsonMap
		}
		if a.IsSensitive() {
			fields[name] = sensitiveField{path: p, value: value}
		}
	}
	for name, b := range blocks {
		p := path + "." + name
		switch n := b.(type) {
		case schema.ListNestedBlock:
			sensitiveFrameworkFields(p, n.NestedObject.Attributes, n.NestedObject.Blocks, fields)
		case schema.SetNestedBlock:
			sensitiveFrameworkFields(p, n.NestedObject.Attributes, n.NestedObject.Blocks, fields)
		case schema.SingleNestedBlock:
			sensitiveFrameworkFields(p, n.Attributes, n.Blocks, fields)
		}
	}
}
