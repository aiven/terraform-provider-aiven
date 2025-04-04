package servicecomponent

import (
	"context"
	"strconv"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func DatasourceServiceComponent() *schema.Resource {
	return &schema.Resource{
		Description: "The Service Component data source provides information about the existing Aiven service Component.",
		ReadContext: datasourceServiceComponentRead,
		Schema: map[string]*schema.Schema{
			"project": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project name",
			},
			"service_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Service name",
			},
			"component": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service component name",
				ValidateFunc: validation.StringInSlice([]string{ // fixme: choices are missing from the schema
					"cassandra",
					"elasticsearch",
					"grafana",
					"influxdb",
					"kafka",
					"kafka_connect",
					"mysql",
					"pg",
					"redis",
					"kibana",
					"kafka_connect",
					"kafka_rest",
					"schema_registry",
					"pgbouncer",
					"prometheus",
					"opensearch",
					"opensearch_dashboards",
				}, false),
			},
			"route": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(service.RouteTypeChoices(), false),
				Description:  userconfig.Desc("Network access route").PossibleValuesString(service.RouteTypeChoices()...).Build(),
			},
			"kafka_authentication_method": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(service.KafkaAuthenticationMethodTypeChoices(), false),
				Description:  userconfig.Desc("Kafka authentication method. This is a value specific to the 'kafka' service component").PossibleValuesString(service.KafkaAuthenticationMethodTypeChoices()...).Build(),
			},
			"kafka_ssl_ca": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: userconfig.Desc("Kafka certificate used").PossibleValuesString(service.KafkaSslCaTypeChoices()...).Build(),
			},
			"ssl": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Whether the endpoint is encrypted or accepts plaintext. By default endpoints are " +
					"always encrypted and this property is only included for service components that may " +
					"disable encryption",
			},
			"usage": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "primary",
				ValidateFunc: validation.StringInSlice(service.UsageTypeChoices(), false),
				Description:  userconfig.Desc("DNS usage name").PossibleValuesString(service.UsageTypeChoices()...).Build(),
			},
			"host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS name for connecting to the service component",
			},
			"port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Port number for connecting to the service component",
			},
		},
	}
}

func datasourceServiceComponentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	componentName := d.Get("component").(string)
	route := d.Get("route").(string)
	usage := d.Get("usage").(string)

	service, err := client.Services.Get(ctx, projectName, serviceName)
	if err != nil {
		return diag.Errorf("common %s/%s not found: %s", projectName, serviceName, err)
	}

	if len(service.Components) == 0 {
		return diag.Errorf("cannot find component %s/%s for service %s",
			componentName, route, serviceName)
	}

	filteredResult := make([]*aiven.ServiceComponents, 0)

	for _, c := range service.Components {
		if c.Component == componentName && c.Route == route && c.Usage == usage {
			if ssl, ok := d.GetOk("ssl"); ok && *c.Ssl != ssl {
				continue
			}

			if m, ok := d.GetOk("kafka_authentication_method"); ok && c.KafkaAuthenticationMethod != m {
				continue
			}

			filteredResult = append(filteredResult, c)
		}
	}

	if len(filteredResult) == 0 {
		return diag.Errorf("cannot find component %s/%s for service %s",
			componentName, route, serviceName)
	}

	// There should NOT be more than one result returned.
	c := filteredResult[0]
	d.SetId(schemautil.BuildResourceID(c.Host, strconv.Itoa(c.Port)))

	if err := d.Set("project", projectName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("component", componentName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("route", route); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("host", c.Host); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("port", c.Port); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("usage", c.Usage); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("kafka_authentication_method", c.KafkaAuthenticationMethod); err != nil {
		return diag.FromErr(err)
	}

	if c.Ssl != nil {
		if err := d.Set("ssl", *c.Ssl); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil

}
