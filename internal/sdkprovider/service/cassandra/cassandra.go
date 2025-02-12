package cassandra

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

const deprecationMessage = `
!> **End of life notice**
Aiven for Apache Cassandra® is entering its [end-of-life cycle](https://aiven.io/docs/platform/reference/end-of-life).
From **November 30, 2025**, it will not be possible to start a new Cassandra service, but existing services will continue to operate until end of life.
From **December 31, 2025**, all active Aiven for Apache Cassandra services are powered off and deleted, making data from these services inaccessible.
To ensure uninterrupted service, complete your migration out of Aiven for Apache Cassandra
before December 31, 2025. For further assistance, contact your account team.
`

func cassandraSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeCassandra)
	s[schemautil.ServiceTypeCassandra] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Values provided by the Cassandra server.",
		MaxItems:    1,
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Cassandra server URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
			},
		},
	}
	return s
}

func ResourceCassandra() *schema.Resource {
	return &schema.Resource{
		Description:        "Creates and manages an [Aiven for Apache Cassandra®](https://aiven.io/docs/products/cassandra) service.",
		DeprecationMessage: deprecationMessage,
		CreateContext:      schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeCassandra),
		ReadContext:        schemautil.ResourceServiceRead,
		UpdateContext:      schemautil.ResourceServiceUpdate,
		DeleteContext:      schemautil.ResourceServiceDelete,
		CustomizeDiff:      schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeCassandra),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:         cassandraSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.Cassandra(),
	}
}
