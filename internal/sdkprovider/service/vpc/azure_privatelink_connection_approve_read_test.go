package vpc

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestAzurePrivatelinkConnectionApprovalRead(t *testing.T) {
	const resourceID = "test-project/test-service"
	connection := aiven.AzurePrivatelinkConnectionResponse{
		PrivatelinkConnectionID: "plc-one", State: "active", UserIPAddress: "10.0.2.4",
	}
	other := aiven.AzurePrivatelinkConnectionResponse{
		PrivatelinkConnectionID: "plc-other", State: "pending-user-approval",
	}
	newData := func(t *testing.T, connectionID string) *schema.ResourceData {
		t.Helper()
		d := schema.TestResourceDataRaw(t, aivenAzurePrivatelinkConnectionApprovalSchema, map[string]any{
			"privatelink_connection_id": connectionID,
		})
		d.SetId(resourceID)
		return d
	}

	t.Run("import discovers the only connection", func(t *testing.T) {
		client := azureApprovalTestClient(t, http.StatusOK, []aiven.AzurePrivatelinkConnectionResponse{connection})
		d := newData(t, "")
		r := ResourceAzurePrivatelinkConnectionApproval()
		imported, err := r.Importer.StateContext(t.Context(), d, client)
		require.NoError(t, err)
		require.Len(t, imported, 1)

		diags := r.ReadContext(t.Context(), imported[0], client)
		require.False(t, diags.HasError(), "%v", diags)
		require.Equal(t, resourceID, imported[0].Id())
		require.Equal(t, "test-project", imported[0].Get("project"))
		require.Equal(t, "test-service", imported[0].Get("service_name"))
		require.Equal(t, connection.PrivatelinkConnectionID, imported[0].Get("privatelink_connection_id"))
		require.Equal(t, connection.State, imported[0].Get("state"))
		require.Equal(t, connection.UserIPAddress, imported[0].Get("endpoint_ip_address"))
	})

	t.Run("read matches the stored connection ID", func(t *testing.T) {
		client := azureApprovalTestClient(t, http.StatusOK, []aiven.AzurePrivatelinkConnectionResponse{other, connection})
		d := newData(t, connection.PrivatelinkConnectionID)

		diags := resourceAzurePrivatelinkConnectionApprovalRead(t.Context(), d, client)
		require.False(t, diags.HasError(), "%v", diags)
		require.Equal(t, resourceID, d.Id())
		require.Equal(t, connection.PrivatelinkConnectionID, d.Get("privatelink_connection_id"))
		require.Equal(t, connection.State, d.Get("state"))
	})

	t.Run("import rejects ambiguous connections", func(t *testing.T) {
		client := azureApprovalTestClient(t, http.StatusOK, []aiven.AzurePrivatelinkConnectionResponse{other, connection})
		d := newData(t, "")

		diags := resourceAzurePrivatelinkConnectionApprovalRead(t.Context(), d, client)
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary, "multiple Azure Private Link connections")
		require.Equal(t, resourceID, d.Id())
	})

	t.Run("empty list removes the resource", func(t *testing.T) {
		client := azureApprovalTestClient(t, http.StatusOK, nil)
		d := newData(t, "")

		diags := resourceAzurePrivatelinkConnectionApprovalRead(t.Context(), d, client)
		require.False(t, diags.HasError(), "%v", diags)
		require.Empty(t, d.Id())
	})

	t.Run("missing stored ID does not adopt another connection", func(t *testing.T) {
		client := azureApprovalTestClient(t, http.StatusOK, []aiven.AzurePrivatelinkConnectionResponse{other})
		d := newData(t, connection.PrivatelinkConnectionID)

		diags := resourceAzurePrivatelinkConnectionApprovalRead(t.Context(), d, client)
		require.False(t, diags.HasError(), "%v", diags)
		require.Empty(t, d.Id())
	})

	t.Run("API errors preserve the resource", func(t *testing.T) {
		client := azureApprovalTestClient(t, http.StatusForbidden, nil)
		d := newData(t, "")

		diags := resourceAzurePrivatelinkConnectionApprovalRead(t.Context(), d, client)
		require.True(t, diags.HasError())
		require.Equal(t, resourceID, d.Id())
	})
}

func azureApprovalTestClient(t *testing.T, status int, connections []aiven.AzurePrivatelinkConnectionResponse) *aiven.Client {
	t.Helper()
	client, err := aiven.NewTokenClient("test-token", "test")
	require.NoError(t, err)
	body, err := json.Marshal(map[string]any{"connections": connections})
	require.NoError(t, err)
	client.Client = &http.Client{Transport: azureApprovalTestTransport(func(request *http.Request) (*http.Response, error) {
		require.Equal(t, http.MethodGet, request.Method)
		require.Equal(t, "/v1/project/test-project/service/test-service/privatelink/azure/connections", request.URL.Path)
		return &http.Response{StatusCode: status, Body: io.NopCloser(bytes.NewReader(body)), Header: make(http.Header)}, nil
	})}
	return client
}

type azureApprovalTestTransport func(*http.Request) (*http.Response, error)

func (f azureApprovalTestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
