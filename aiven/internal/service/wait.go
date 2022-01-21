// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package service

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/aiven/aiven-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	aivenTargetState           = "RUNNING"
	aivenPendingState          = "REBUILDING"
	aivenRebalancingState      = "REBALANCING"
	aivenServicesStartingState = "WAITING_FOR_SERVICES"
)

func WaitForCreation(ctx context.Context, d *schema.ResourceData, m interface{}) (*aiven.Service, error) {
	return waitForCreateOrUpdate(ctx, d, m, false)
}

func WaitForUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) (*aiven.Service, error) {
	return waitForCreateOrUpdate(ctx, d, m, true)
}

func waitForCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}, isUpdate bool) (*aiven.Service, error) {
	client := m.(*aiven.Client)

	projectName, serviceName := d.Get("project").(string), d.Get("service_name").(string)

	timeout := d.Timeout(schema.TimeoutCreate)
	if isUpdate {
		log.Printf("[DEBUG] Service update waiter timeout %.0f minutes", timeout.Minutes())
		timeout = d.Timeout(schema.TimeoutUpdate)
	} else {
		log.Printf("[DEBUG] Service create waiter timeout %.0f minutes", timeout.Minutes())
	}

	conf := &resource.StateChangeConf{
		Pending:                   []string{aivenPendingState, aivenRebalancingState, aivenServicesStartingState},
		Target:                    []string{aivenTargetState},
		Delay:                     10 * time.Second,
		Timeout:                   timeout,
		MinTimeout:                2 * time.Second,
		ContinuousTargetOccurence: 5,
		Refresh: func() (interface{}, string, error) {
			service, err := client.Services.Get(projectName, serviceName)
			if err != nil {
				return nil, "", fmt.Errorf("unable to fetch service from api: %w", err)
			}

			state := service.State
			if isUpdate {
				// When updating service don't wait for it to enter RUNNING state because that can take
				// very long time if for example service plan or cloud it runs in is changed and the
				// service has a lot of data. If the service was already previously in RUNNING state we
				// can manage the associated resources even if the service is rebuilding.
				state = aivenTargetState
			}

			if state != aivenTargetState {
				log.Printf("[DEBUG] service reports as %s, still for it to be in state %s", state, aivenTargetState)
				return service, state, nil
			}

			if rdy := backupsReady(service); !rdy {
				log.Printf("[DEBUG] service reports as %s, still waiting for service backups", state)
				return service, aivenServicesStartingState, nil
			}

			if rdy := grafanaReady(service); !rdy {
				log.Printf("[DEBUG] service reports as %s, still waiting for grafana", state)
				return service, aivenServicesStartingState, nil
			}
			return service, state, nil
		},
	}

	aux, err := conf.WaitForStateContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to wait for service state change: %w", err)
	}
	return aux.(*aiven.Service), nil
}

func grafanaReady(service *aiven.Service) bool {
	if service.Type != "grafana" {
		return true
	}

	// if IP filter is anything but 0.0.0.0/0 skip Grafana service availability checks
	ipFilters, ok := service.UserConfig["ip_filter"]
	if ok {
		f := ipFilters.([]interface{})
		if len(f) > 1 {
			log.Printf("[DEBUG] grafana serivce has `%+v` ip filters, and availability checks will be skipped", ipFilters)

			return true
		}

		if len(f) == 1 {
			if f[0] != "0.0.0.0/0" {
				log.Printf("[DEBUG] grafana serivce has `%+v` ip filters, and availability checks will be skipped", ipFilters)

				return true
			}
		}
	}

	var publicGrafana string

	// constructing grafana public address if available
	for _, component := range service.Components {
		if component.Route == "public" && component.Usage == "primary" {
			publicGrafana = component.Host + ":" + strconv.Itoa(component.Port)
			continue
		}
	}

	// checking if public grafana is reachable
	if publicGrafana != "" {
		_, err := net.DialTimeout("tcp", publicGrafana, 1*time.Second)
		if err != nil {
			log.Printf("[DEBUG] public grafana is not yet reachable")
			return false
		}

		log.Printf("[DEBUG] public grafana is reachable")
		return true
	}

	return true
}

func backupsReady(service *aiven.Service) bool {
	if service.Type != "pg" && service.Type != "elasticsearch" &&
		service.Type != "redis" && service.Type != "influxdb" {
		return true
	}

	// no backups for read replicas type of service
	for _, i := range service.Integrations {
		if i.IntegrationType == "read_replica" && *i.DestinationService == service.Name {
			return true
		}
	}

	return len(service.Backups) > 0
}
