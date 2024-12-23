package alloydbomni

import (
	"context"
	"encoding/json"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/alloydbomni"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const serviceAccountCredentials = "service_account_credentials"

func aivenAlloyDBOmniSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeAlloyDBOmni)
	s[serviceAccountCredentials] = &schema.Schema{
		Description:      "Your [Google service account key](https://cloud.google.com/iam/docs/service-account-creds#key-types) in JSON format.",
		Optional:         true,
		Sensitive:        true,
		Type:             schema.TypeString,
		ValidateDiagFunc: validateServiceAccountCredentials,
	}
	s[schemautil.ServiceTypeAlloyDBOmni] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Values provided by the AlloyDB Omni server.",
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// TODO: Remove `uri` in the next major version.
				"uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "AlloyDB Omni primary connection URI.",
					Optional:    true,
					Sensitive:   true,
				},
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "AlloyDB Omni primary connection URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
				"bouncer": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "PgBouncer connection details for [connection pooling](https://aiven.io/docs/products/postgresql/concepts/pg-connection-pooling).",
					Deprecated:  "This field was added by mistake and has never worked. It will be removed in future versions.",
				},
				// TODO: Remove `host` in the next major version.
				"host": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "AlloyDB Omni primary node host IP or name.",
				},
				// TODO: Remove `port` in the next major version.
				"port": {
					Type:        schema.TypeInt,
					Computed:    true,
					Sensitive:   true,
					Description: "AlloyDB Omni port.",
				},
				// TODO: Remove `sslmode` in the next major version.
				"sslmode": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "AlloyDB Omni SSL mode setting.",
				},
				// TODO: Remove `user` in the next major version.
				"user": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "AlloyDB Omni admin user name.",
				},
				// TODO: Remove `password` in the next major version.
				"password": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "AlloyDB Omni admin user password.",
					Sensitive:   true,
				},
				// TODO: Remove `dbname` in the next major version.
				"dbname": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "Primary AlloyDB Omni database name.",
				},
				"params": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "AlloyDB Omni connection parameters.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"host": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "AlloyDB Omni host IP or name.",
							},
							"port": {
								Type:        schema.TypeInt,
								Computed:    true,
								Sensitive:   true,
								Description: "AlloyDB Omni port.",
							},
							"sslmode": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "AlloyDB Omni SSL mode setting.",
							},
							"user": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "AlloyDB Omni admin user name.",
							},
							"password": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "AlloyDB Omni admin user password.",
							},
							"database_name": {
								Type:        schema.TypeString,
								Computed:    true,
								Sensitive:   true,
								Description: "Primary AlloyDB Omni database name.",
							},
						},
					},
				},
				"replica_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "AlloyDB Omni replica URI for services with a replica.",
					Sensitive:   true,
				},
				"standby_uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "AlloyDB Omni standby connection URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
				"syncing_uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "AlloyDB Omni syncing connection URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
				// TODO: This isn't in the connection info, but it's in the metadata.
				//  We should move this to the other part of the schema in the next major version.
				"max_connections": {
					Type:        schema.TypeInt,
					Computed:    true,
					Sensitive:   true,
					Description: "The [number of allowed connections](https://aiven.io/docs/products/postgresql/reference/pg-connection-limits). Varies based on the service plan.",
				},
			},
		},
	}
	return s
}

func ResourceAlloyDBOmni() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages an Aiven for AlloyDB Omni service.",
		CreateContext: schemautil.ComposeContexts(
			schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeAlloyDBOmni),
			common.WithGenClient(serviceAccountCredentialsUpsert),
		),
		ReadContext: schemautil.ComposeContexts(
			schemautil.ResourceServiceRead,
			common.WithGenClient(serviceAccountCredentialsRead),
		),
		UpdateContext: schemautil.ComposeContexts(
			schemautil.ResourceServiceUpdate,
			common.WithGenClient(serviceAccountCredentialsUpsert),
		),
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeAlloyDBOmni),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts:      schemautil.DefaultResourceTimeouts(),
		Schema:        aivenAlloyDBOmniSchema(),
		SchemaVersion: 1,
	}
}

// serviceAccountCredentialsUpsert sets, updates and removes service account credentials
func serviceAccountCredentialsUpsert(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	privateKey, ok := d.GetOk(serviceAccountCredentials)
	if ok {
		req := &alloydbomni.AlloyDbOmniGoogleCloudPrivateKeySetIn{PrivateKey: privateKey.(string)}
		_, err = client.AlloyDbOmniGoogleCloudPrivateKeySet(ctx, projectName, serviceName, req)
	} else {
		_, err = client.AlloyDbOmniGoogleCloudPrivateKeyRemove(ctx, projectName, serviceName)
	}
	return err
}

// serviceAccountCredentialsRead reads remote service account credentials and compares with the local value
// This function is unnecessary for the most cases:
// terraform will handle all the plan and apply logic for add/change/delete.
// It covers just one case: when the remote value is different from the local value.
// For instance, when it was modified or removed in the Console.
func serviceAccountCredentialsRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	// Compares remote and local private keys
	// When they differ, the local value is set to an empty string to enforce plan to show the change
	remote, err := client.AlloyDbOmniGoogleCloudPrivateKeyIdentify(ctx, projectName, serviceName)
	if err != nil {
		// Something terrible happened, we should return the error
		// This endpoint doesn't return 404, if service exists, it will return the DTO with null values
		return err
	}

	// Unmarshals local private key into DTO to compare with the remote value
	local := new(alloydbomni.AlloyDbOmniGoogleCloudPrivateKeyIdentifyOut)
	if v, ok := d.GetOk(serviceAccountCredentials); ok {
		err = json.Unmarshal([]byte(v.(string)), local)
		if err != nil {
			return fmt.Errorf("failed to unmarshal config service_account_credentials: %w", err)
		}
	}

	if !cmp.Equal(remote, local) {
		// 1. Remote key does not exist, or
		// 2. Remote key is different from the local value
		// This will enforce the plan to show the change
		return d.Set(serviceAccountCredentials, "")
	}

	return nil
}
