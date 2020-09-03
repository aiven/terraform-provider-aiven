// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"strconv"
)

func datasourceServiceComponent() *schema.Resource {
	return &schema.Resource{
		Read: datasourceServiceComponentRead,
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

func datasourceServiceComponentRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	componentName := d.Get("component").(string)
	route := d.Get("route").(string)
	usage := d.Get("usage").(string)

	service, err := client.Services.Get(projectName, serviceName)
	if err != nil {
		return fmt.Errorf("service %s/%s not found: %w", projectName, serviceName, err)
	}

	for _, c := range service.Components {
		if c.Component == componentName && c.Route == route && c.Usage == usage {
			// check optional ssl search criteria, if not set by a user match entries
			// without ssl or ssl=true
			if ssl, ok := d.GetOk("ssl"); ok {
				if c.Ssl == nil {
					continue
				}

				if *c.Ssl != ssl.(bool) {
					continue
				}
			} else {
				if !(c.Ssl == nil || *c.Ssl == true) {
					continue
				}
			}

			// check optional kafka_authentication_method search criteria, if not set by a
			// user match entries without kafka_authentication_method
			if method, ok := d.GetOk("kafka_authentication_method"); ok {
				if c.KafkaAuthenticationMethod != method {
					continue
				}
			} else {
				if c.KafkaAuthenticationMethod != "" {
					continue
				}
			}

			d.SetId(buildResourceID(c.Host, strconv.Itoa(c.Port)))

			if err := d.Set("project", projectName); err != nil {
				return err
			}
			if err := d.Set("service_name", serviceName); err != nil {
				return err
			}
			if err := d.Set("component", componentName); err != nil {
				return err
			}
			if err := d.Set("route", route); err != nil {
				return err
			}
			if err := d.Set("host", c.Host); err != nil {
				return err
			}
			if err := d.Set("port", c.Port); err != nil {
				return err
			}
			if err := d.Set("usage", c.Usage); err != nil {
				return err
			}
			if err := d.Set("kafka_authentication_method", c.KafkaAuthenticationMethod); err != nil {
				return err
			}

			if c.Ssl != nil {
				if err := d.Set("ssl", *c.Ssl); err != nil {
					return err
				}
			}

			return nil
		}
	}

	return fmt.Errorf("cannot find component %s/%s for service %s",
		componentName, route, serviceName)
}
