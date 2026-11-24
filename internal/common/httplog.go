package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// EnvLogHTTP sets the level of the API request and response logs, e.g., DEBUG.
// The logs are off without it, so TF_LOG alone doesn't log the API bodies.
const EnvLogHTTP = "TF_LOG_PROVIDER_AIVEN_HTTP"

const logSubsystem = "http"

// secretKey matches the JSON keys whose values are masked, e.g., "password",
// "ssl_client_key", "connection.password" or "secretKey". It also matches
// a few harmless keys, e.g., the tag "key".
var secretKey = regexp.MustCompile(
	`(?:^|[_.-])(?i:password|passphrase|secrets?|token|key|credentials|connection_string)$` +
		`|[a-z0-9](?:Password|Passphrase|Secrets?|Token|Key|Credentials)$`,
)

// secretValue matches a PEM private key or a JWT, whatever key holds it.
var secretValue = regexp.MustCompile(
	`-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----(?s:.*?)(?:-----END [A-Z0-9 ]*PRIVATE KEY-----|$)` +
		`|\beyJ[A-Za-z0-9_-]{2,}\.eyJ[A-Za-z0-9_-]{2,}\.[A-Za-z0-9_-]*`,
)

// loggingTransport logs each API request and response, with the secrets masked.
type loggingTransport struct{ next http.RoundTripper }

func (t loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := tflog.NewSubsystem(req.Context(), logSubsystem, tflog.WithLevelFromEnv(EnvLogHTTP))
	ctx = tflog.SubsystemSetField(ctx, logSubsystem, "operation_id", ctx.Value(avngen.OperationIDKey{}))

	var reqBody []byte
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		reqBody = b
		req = req.Clone(req.Context())
		req.Body = io.NopCloser(bytes.NewReader(b))
	}

	tflog.SubsystemDebug(ctx, logSubsystem, "Sending API request", map[string]any{
		"method": req.Method,
		"url":    req.URL.String(),
		"body":   maskBody(reqBody),
	})

	start := time.Now()
	rsp, err := t.next.RoundTrip(req)
	if err != nil {
		tflog.SubsystemDebug(ctx, logSubsystem, "API request failed", map[string]any{
			"duration": time.Since(start).String(),
			"error":    err.Error(),
		})
		return nil, err
	}

	rspBody, err := io.ReadAll(rsp.Body)
	if err = errors.Join(err, rsp.Body.Close()); err != nil {
		return nil, err
	}
	rsp.Body = io.NopCloser(bytes.NewReader(rspBody))

	tflog.SubsystemDebug(ctx, logSubsystem, "Received API response", map[string]any{
		"duration":    time.Since(start).String(),
		"status_code": rsp.StatusCode,
		"body":        maskBody(rspBody),
	})
	return rsp, nil
}

// maskBody returns the JSON body with the values of secret keys, secretValue
// and the passwords in URIs masked. A body that isn't JSON, e.g., a proxy error
// page, is returned with only secretValue masked.
func maskBody(b []byte) string {
	if !json.Valid(b) {
		return secretValue.ReplaceAllString(string(b), "***")
	}

	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var v any
	if err := d.Decode(&v); err != nil {
		return fmt.Sprintf("<%d bytes, can't mask: %s>", len(b), err)
	}

	masked, err := json.Marshal(maskValue(v))
	if err != nil {
		return fmt.Sprintf("<%d bytes, can't mask: %s>", len(b), err)
	}
	return string(masked)
}

func maskValue(v any) any {
	switch v := v.(type) {
	case map[string]any:
		for k, x := range v {
			if x != nil && secretKey.MatchString(k) {
				v[k] = "***"
			} else {
				v[k] = maskValue(x)
			}
		}
	case []any:
		for i, x := range v {
			v[i] = maskValue(x)
		}
	case string:
		if secretValue.MatchString(v) {
			return "***"
		}
		if u, err := url.Parse(v); err == nil && u.User != nil {
			return u.Redacted()
		}
	}
	return v
}
