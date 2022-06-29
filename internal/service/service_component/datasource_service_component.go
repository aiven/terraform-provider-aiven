package service_component

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/aiven/terraform-provider-aiven/internal/meta"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				ValidateFunc: validation.StringInSlice([]string{
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
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Network access route",
				ValidateFunc: validation.StringInSlice([]string{
					"dynamic",
					"public",
					"private",
					"privatelink",
				}, false),
			},
			"kafka_authentication_method": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Kafka authentication method. This is a value specific to the 'kafka' service component",
				ValidateFunc: validation.StringInSlice([]string{
					"certificate",
					"sasl",
				}, false),
			},
			"ssl": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Whether the endpoint is encrypted or accepts plaintext. By default endpoints are " +
					"always encrypted and this property is only included for service components that may " +
					"disable encryption",
			},
			"usage": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DNS usage name",
				Default:     "primary",
				ValidateFunc: validation.StringInSlice([]string{
					"primary",
					"replica",
					"syncing",
				}, false),
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

func datasourceServiceComponentRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	componentName := d.Get("component").(string)
	route := d.Get("route").(string)
	usage := d.Get("usage").(string)

	service, err := client.Services.Get(projectName, serviceName)
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
			filteredBySsl, err := filterDatasourceServiceComponents(d, "ssl", service.Components)
			if err != nil {
				return diag.FromErr(err)
			}
			filteredByKafkaAuthMethod, err := filterDatasourceServiceComponents(d, "kafka_authentication_method", filteredBySsl)
			if err != nil {
				return diag.FromErr(err)
			}

			filteredResult = filteredByKafkaAuthMethod
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

func filterDatasourceServiceComponents(d *schema.ResourceData, filter string, components []*aiven.ServiceComponents) ([]*aiven.ServiceComponents, error) {
	filteredResult := make([]*aiven.ServiceComponents, 0)

	switch filter {
	case "ssl":
		// check optional ssl search criteria, if not set by a user match entries
		// without ssl or ssl=true
		ssl, ok := d.GetOk("ssl")
		for _, c := range components {
			if ok {
				if c.Ssl == nil {
					continue
				}

				if *c.Ssl != ssl.(bool) {
					continue
				}
			} else {
				if !(c.Ssl == nil || *c.Ssl) {
					continue
				}
			}

			filteredResult = append(filteredResult, c)
		}

		if len(filteredResult) == 0 {
			var errorMessage string
			if ok {
				errorMessage = "cannot match the components with the given SSL criteria"
			} else {
				errorMessage = "please try specifying 'ssl' to filter the results"
			}
			return nil, errors.New(errorMessage)
		}

		return filteredResult, nil

	case "kafka_authentication_method":
		// check optional kafka_authentication_method search criteria, if not set by a
		// user match entries without kafka_authentication_method
		method, ok := d.GetOk("kafka_authentication_method")
		for _, c := range components {
			if ok {
				if c.KafkaAuthenticationMethod != method {
					continue
				}
			} else {
				if c.KafkaAuthenticationMethod != "" {
					continue
				}
			}

			filteredResult = append(filteredResult, c)
		}

		if len(filteredResult) == 0 {
			var errorMessage string
			if ok {
				errorMessage = fmt.Sprintf("no result matches (kafka_authentication_method=%s)", method.(string))
			} else {
				errorMessage = "please try specifying 'kafka_authentication_method' to filter the results"
			}
			return nil, errors.New(errorMessage)
		}
		return filteredResult, nil
	default:
		return nil, errors.New("no filtering criteria provided")
	}
}
