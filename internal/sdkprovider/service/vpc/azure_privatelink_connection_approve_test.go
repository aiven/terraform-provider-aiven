package vpc_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/vpc"
)

type azureSecrets struct {
	Project        string `envconfig:"AIVEN_PROJECT_NAME" required:"true"`
	AivenAppID     string `envconfig:"AIVEN_AZURE_APP_ID" required:"true"`
	TenantID       string `envconfig:"AZURE_TENANT_ID" required:"true"`
	SubscriptionID string `envconfig:"AZURE_SUBSCRIPTION_ID" required:"true"`
}

func TestAccAivenAzurePrivatelinkConnectionApproval_basic(t *testing.T) {
	var s azureSecrets
	err := envconfig.Process("", &s)
	if err != nil {
		t.Skipf("Not all values has been provided: %s", err)
	}

	prefix := "test-tf-acc-plapproval-" + acctest.RandString(7)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"azurerm": {
				Source:            "hashicorp/azurerm",
				VersionConstraint: "=3.30.0",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccAzurePrivatelinkConnectionApprovalResource(prefix, &s),
				Check: resource.ComposeTestCheckFunc(
					// Aiven resources
					resource.TestCheckResourceAttr("aiven_project_vpc.project_vpc", "state", "ACTIVE"),
					resource.TestCheckResourceAttr("aiven_pg.pg", "state", "RUNNING"),
					resource.TestCheckResourceAttr("aiven_azure_privatelink.private_link", "state", "active"),
					resource.TestCheckResourceAttr("aiven_azure_privatelink_connection_approval.approval", "state", "active"),
					resource.TestCheckResourceAttrSet("aiven_azure_privatelink_connection_approval.approval", "endpoint_ip_address"),
					resource.TestCheckResourceAttrSet("aiven_azure_privatelink_connection_approval.approval", "privatelink_connection_id"),

					// Azure resources
					resource.TestCheckResourceAttrSet("azurerm_resource_group.resource_group", "id"),
					resource.TestCheckResourceAttrSet("azurerm_virtual_network.virtual_network", "id"),
					resource.TestCheckResourceAttrSet("azurerm_subnet.subnet", "id"),
					resource.TestCheckResourceAttrSet("azurerm_private_endpoint.private_endpoint", "id"),
				),
			},
			importStateByName("aiven_project_vpc.project_vpc"),
			importStateByName("aiven_pg.pg"),
			importStateByName("aiven_azure_privatelink.private_link"),
			{
				ResourceName:      "aiven_azure_privatelink_connection_approval.approval",
				ImportState:       true,
				ImportStateVerify: true,
			},
			importStateByName("azurerm_resource_group.resource_group"),
			importStateByName("azurerm_virtual_network.virtual_network"),
			importStateByName("azurerm_subnet.subnet"),
			importStateByName("azurerm_private_endpoint.private_endpoint"),
		},
	})
}

func TestAzurePrivatelinkConnectionApprovalRead(t *testing.T) {
	const resourceID = "test-project/test-service"
	r := vpc.ResourceAzurePrivatelinkConnectionApproval()
	connection := aiven.AzurePrivatelinkConnectionResponse{
		PrivatelinkConnectionID: "plc-one", State: "active", UserIPAddress: "10.0.2.4",
	}
	other := aiven.AzurePrivatelinkConnectionResponse{
		PrivatelinkConnectionID: "plc-other", State: "pending-user-approval",
	}
	newData := func(t *testing.T, connectionID string) *schema.ResourceData {
		t.Helper()
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
			"privatelink_connection_id": connectionID,
		})
		d.SetId(resourceID)
		return d
	}

	t.Run("import discovers the only connection", func(t *testing.T) {
		client := azureApprovalTestClient(t, http.StatusOK, []aiven.AzurePrivatelinkConnectionResponse{connection})
		d := newData(t, "")
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

		diags := r.ReadContext(t.Context(), d, client)
		require.False(t, diags.HasError(), "%v", diags)
		require.Equal(t, resourceID, d.Id())
		require.Equal(t, connection.PrivatelinkConnectionID, d.Get("privatelink_connection_id"))
		require.Equal(t, connection.State, d.Get("state"))
	})

	t.Run("import rejects ambiguous connections", func(t *testing.T) {
		client := azureApprovalTestClient(t, http.StatusOK, []aiven.AzurePrivatelinkConnectionResponse{other, connection})
		d := newData(t, "")

		diags := r.ReadContext(t.Context(), d, client)
		require.True(t, diags.HasError())
		require.Contains(t, diags[0].Summary, "multiple Azure Private Link connections")
		require.Equal(t, resourceID, d.Id())
	})

	t.Run("empty list removes the resource", func(t *testing.T) {
		client := azureApprovalTestClient(t, http.StatusOK, nil)
		d := newData(t, "")

		diags := r.ReadContext(t.Context(), d, client)
		require.False(t, diags.HasError(), "%v", diags)
		require.Empty(t, d.Id())
	})

	t.Run("missing stored ID does not adopt another connection", func(t *testing.T) {
		client := azureApprovalTestClient(t, http.StatusOK, []aiven.AzurePrivatelinkConnectionResponse{other})
		d := newData(t, connection.PrivatelinkConnectionID)

		diags := r.ReadContext(t.Context(), d, client)
		require.False(t, diags.HasError(), "%v", diags)
		require.Empty(t, d.Id())
	})

	t.Run("API errors preserve the resource", func(t *testing.T) {
		client := azureApprovalTestClient(t, http.StatusForbidden, nil)
		d := newData(t, "")

		diags := r.ReadContext(t.Context(), d, client)
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

func testAccAzurePrivatelinkConnectionApprovalResource(prefix string, s *azureSecrets) string {
	return fmt.Sprintf(`
data "aiven_project" "project" {
  project = %[2]q
}

provider "azurerm" {
  features {}
  tenant_id       = %[3]q
  subscription_id = %[4]q
}

resource "aiven_project_vpc" "project_vpc" {
  project      = data.aiven_project.project.project
  cloud_name   = "azure-germany-north"
  network_cidr = "192.168.1.0/24"

  timeouts {
    create = "15m"
  }
}

resource "aiven_static_ip" "static_ips" {
  count      = 2
  project    = aiven_project_vpc.project_vpc.project
  cloud_name = aiven_project_vpc.project_vpc.cloud_name
}

resource "aiven_pg" "pg" {
  service_name   = "%[1]s-pg"
  project        = data.aiven_project.project.project
  project_vpc_id = aiven_project_vpc.project_vpc.id
  cloud_name     = aiven_project_vpc.project_vpc.cloud_name
  plan           = "startup-4"
  static_ips     = [for sip in aiven_static_ip.static_ips : sip.static_ip_address_id]

  pg_user_config {
    static_ips = true

    privatelink_access {
      pg        = true
      pgbouncer = true
    }
  }
}

resource "aiven_azure_privatelink" "private_link" {
  project      = data.aiven_project.project.project
  service_name = aiven_pg.pg.service_name

  user_subscription_ids = [
    %[4]q,
  ]
}

resource "azurerm_resource_group" "resource_group" {
  location = "germanynorth"
  name     = "%[1]s-private-link"
}

resource "azurerm_virtual_network" "virtual_network" {
  name                = "%[1]s-virtual-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
}

resource "azurerm_subnet" "subnet" {
  name                 = "%[1]s-subnet"
  resource_group_name  = azurerm_resource_group.resource_group.name
  virtual_network_name = azurerm_virtual_network.virtual_network.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_private_endpoint" "private_endpoint" {
  name                = "%[1]s-private-endpoint"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  subnet_id           = azurerm_subnet.subnet.id

  private_service_connection {
    name                           = aiven_pg.pg.service_name
    request_message                = aiven_pg.pg.service_name
    private_connection_resource_id = aiven_azure_privatelink.private_link.azure_service_id
    is_manual_connection           = true
  }

  depends_on = [
    aiven_azure_privatelink.private_link,
  ]
}

resource "aiven_azure_privatelink_connection_approval" "approval" {
  project             = data.aiven_project.project.project
  service_name        = aiven_pg.pg.service_name
  endpoint_ip_address = azurerm_private_endpoint.private_endpoint.private_service_connection[0].private_ip_address
}`, prefix, s.Project, s.TenantID, s.SubscriptionID)
}
