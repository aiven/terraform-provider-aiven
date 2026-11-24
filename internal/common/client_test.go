package common

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenClientLogsMaskedRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(`{"service": {
			"service_name": "my-pg",
			"service_uri": "postgres://avnadmin:uri-secret@my-pg.aivencloud.com:1234/defaultdb",
			"users": [{"username": "app", "password": "user-secret"}]
		}}`))
		assert.NoError(t, err)
	}))
	defer server.Close()
	t.Setenv("AIVEN_WEB_URL", server.URL)

	modify := func(t *testing.T) string {
		client, err := NewAivenGenClient(TokenOpt("token-secret"))
		require.NoError(t, err)

		var output bytes.Buffer
		ctx := tflogtest.RootLogger(t.Context(), &output)
		out, err := client.ServiceUserCredentialsModify(
			ctx,
			"my-project",
			"my-pg",
			"app",
			&service.ServiceUserCredentialsModifyIn{
				NewPassword: new("request-secret"),
				Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "my-pg", out.ServiceName)
		return output.String()
	}

	t.Run("off by default", func(t *testing.T) {
		assert.NotContains(t, modify(t), "API re")
	})

	t.Run("masked", func(t *testing.T) {
		t.Setenv(EnvLogHTTP, "DEBUG")
		logs := modify(t)
		assert.Contains(t, logs, `"@message":"Sending API request"`)
		assert.Contains(t, logs, `"@message":"Received API response"`)
		assert.Contains(t, logs, `"operation_id":"ServiceUserCredentialsModify"`)
		assert.Contains(t, logs, `\"new_password\":\"***\"`)
		assert.Contains(t, logs, `\"password\":\"***\"`)
		assert.Contains(t, logs, `postgres://avnadmin:xxxxx@my-pg.aivencloud.com:1234/defaultdb`)
		assert.Contains(t, logs, `\"username\":\"app\"`)
		for _, secret := range []string{"token-secret", "request-secret", "uri-secret", "user-secret"} {
			assert.NotContains(t, logs, secret)
		}
	})
}

func TestMaskBody(t *testing.T) {
	cases := []struct {
		name, body, expect string
	}{
		{name: "empty"},
		{
			name:   "secret keys",
			body:   `{"password": "a", "ssl_client_key": "b", "connection.password": "c", "secretKey": "d", "API_KEY": "e"}`,
			expect: `{"API_KEY":"***","connection.password":"***","password":"***","secretKey":"***","ssl_client_key":"***"}`,
		},
		{
			name:   "nested and non-string secrets",
			body:   `{"users": [{"username": "app", "access_key": "a"}], "secrets": {"A": "b"}, "credentials": ["c"]}`,
			expect: `{"credentials":"***","secrets":"***","users":[{"access_key":"***","username":"app"}]}`,
		},
		{
			name:   "uri password",
			body:   `{"service_uri": "postgres://avnadmin:secret@host:1234/db?sslmode=require", "uri": "https://host/path"}`,
			expect: `{"service_uri":"postgres://avnadmin:xxxxx@host:1234/db?sslmode=require","uri":"https://host/path"}`,
		},
		{
			name:   "keeps nulls, numbers and harmless keys",
			body:   `{"password": null, "retention_ms": 9223372036854775807, "valkey": {"token_prefix": "abc"}}`,
			expect: `{"password":null,"retention_ms":9223372036854775807,"valkey":{"token_prefix":"abc"}}`,
		},
		{
			name: "secret values under harmless keys",
			body: `{"cert": "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----",` +
				` "pem": "-----BEGIN RSA PRIVATE KEY-----\nMIIE\n-----END RSA PRIVATE KEY-----",` +
				` "id": "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhcHAifQ.c2lnbmF0dXJl", "name": "eyJ"}`,
			expect: `{"cert":"-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----","id":"***","name":"eyJ","pem":"***"}`,
		},
		{
			name:   "not json",
			body:   `<html>502 Bad Gateway</html>`,
			expect: `<html>502 Bad Gateway</html>`,
		},
		{
			name:   "not json with secret values",
			body:   "error: Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhcHAifQ.c2ln for -----BEGIN PRIVATE KEY-----\nMIIE\n-----END PRIVATE KEY----- done",
			expect: "error: Bearer *** for *** done",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, maskBody([]byte(tt.body)))
		})
	}
}

func TestNewClientOpts(t *testing.T) {
	cases := []struct {
		name      string
		opts      []ClientOpt
		envToken  string
		expectErr error
		expect    *clientOpts
	}{
		{
			name:      "empty options",
			expectErr: errTokenRequired,
		},
		{
			name:     "env token",
			envToken: "foo",
			expect: &clientOpts{
				token:        "foo",
				tfVersion:    "0.11+compatible",
				buildVersion: "dev",
				userAgent:    "terraform-provider-aiven/0.11+compatible/dev",
			},
		},
		{
			name: "custom token",
			opts: []ClientOpt{
				TokenOpt("opt token"),
			},
			expect: &clientOpts{
				token:        "opt token",
				tfVersion:    "0.11+compatible",
				buildVersion: "dev",
				userAgent:    "terraform-provider-aiven/0.11+compatible/dev",
			},
		},
		{
			name:     "custom version number",
			envToken: "foo",
			opts: []ClientOpt{
				TFVersionOpt("bar"),
			},
			expect: &clientOpts{
				token:        "foo",
				tfVersion:    "bar",
				buildVersion: "dev",
				userAgent:    "terraform-provider-aiven/bar/dev",
			},
		},
		{
			name:     "custom build number",
			envToken: "bar",
			opts: []ClientOpt{
				BuildVersionOpt("baz"),
			},
			expect: &clientOpts{
				token:        "bar",
				tfVersion:    "0.11+compatible",
				buildVersion: "baz",
				userAgent:    "terraform-provider-aiven/0.11+compatible/baz",
			},
		},
		{
			name: "custom all",
			opts: []ClientOpt{
				BuildVersionOpt("baz"),
				TFVersionOpt("bar"),
				TokenOpt("foo"),
			},
			expect: &clientOpts{
				token:        "foo",
				tfVersion:    "bar",
				buildVersion: "baz",
				userAgent:    "terraform-provider-aiven/bar/baz",
			},
		},
	}

	for _, o := range cases {
		t.Run(o.name, func(t *testing.T) {
			t.Setenv("AIVEN_TOKEN", o.envToken) // must not expose a real token in logs
			actual, err := newClientOpts(o.opts...)
			assert.Equal(t, o.expectErr, err)
			assert.Equal(t, o.expect, actual)
		})
	}
}
