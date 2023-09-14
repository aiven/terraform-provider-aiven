package schemautil

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	aivenTargetState           = "RUNNING"
	aivenPendingState          = "REBUILDING"
	aivenRebalancingState      = "REBALANCING"
	aivenServicesStartingState = "WAITING_FOR_SERVICES"
)

// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func WaitForServiceCreation(ctx context.Context, d *schema.ResourceData, m interface{}) (*aiven.Service, error) {
	client := m.(*aiven.Client)

	projectName, serviceName := d.Get("project").(string), d.Get("service_name").(string)

	timeout := d.Timeout(schema.TimeoutCreate)
	log.Printf("[DEBUG] Service creation waiter timeout %.0f minutes", timeout.Minutes())

	conf := &resource.StateChangeConf{
		Pending:                   []string{aivenPendingState, aivenRebalancingState, aivenServicesStartingState},
		Target:                    []string{aivenTargetState},
		Delay:                     10 * time.Second,
		Timeout:                   timeout,
		MinTimeout:                2 * time.Second,
		ContinuousTargetOccurence: 5,
		Refresh: func() (interface{}, string, error) {
			service, err := client.Services.Get(ctx, projectName, serviceName)
			if err != nil {
				return nil, "", fmt.Errorf("unable to fetch service from api: %w", err)
			}

			state := service.State

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

			if rdy, err := staticIpsReady(ctx, d, m); err != nil {
				return nil, "", fmt.Errorf("unable to check if static ips are ready: %w", err)
			} else if !rdy {
				log.Printf("[DEBUG] service reports as %s, still waiting for static ips", state)
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

// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func WaitForServiceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) (*aiven.Service, error) {
	client := m.(*aiven.Client)

	projectName, serviceName := d.Get("project").(string), d.Get("service_name").(string)

	timeout := d.Timeout(schema.TimeoutCreate)
	log.Printf("[DEBUG] Service update waiter timeout %.0f minutes", timeout.Minutes())

	conf := &resource.StateChangeConf{
		Pending:                   []string{"updating"},
		Target:                    []string{"updated"},
		Delay:                     10 * time.Second,
		Timeout:                   timeout,
		MinTimeout:                2 * time.Second,
		ContinuousTargetOccurence: 5,
		Refresh: func() (interface{}, string, error) {
			service, err := client.Services.Get(ctx, projectName, serviceName)
			if err != nil {
				return nil, "", fmt.Errorf("unable to fetch service from api: %w", err)
			}

			state := service.State

			if rdy := backupsReady(service); !rdy {
				log.Printf("[DEBUG] service reports as %s, still waiting for service backups", state)
				return service, "updating", nil
			}

			if rdy := grafanaReady(service); !rdy {
				log.Printf("[DEBUG] service reports as %s, still waiting for grafana", state)
				return service, "updating", nil
			}

			if rdy, err := staticIpsReady(ctx, d, m); err != nil {
				return nil, "", fmt.Errorf("unable to check if static ips are ready: %w", err)
			} else if !rdy {
				log.Printf("[DEBUG] service reports as %s, still waiting for static ips", state)
				return service, "updating", nil
			}

			return service, "updated", nil
		},
	}

	aux, err := conf.WaitForStateContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to wait for service state change: %w", err)
	}
	return aux.(*aiven.Service), nil
}

// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func WaitStaticIpsDissassociation(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	timeout := d.Timeout(schema.TimeoutDelete)
	log.Printf("[DEBUG] Static Ip dissassociation timeout %.0f minutes", timeout.Minutes())

	conf := &resource.StateChangeConf{
		Pending:                   []string{"doing"},
		Target:                    []string{"done"},
		Delay:                     10 * time.Second,
		Timeout:                   timeout,
		MinTimeout:                2 * time.Second,
		ContinuousTargetOccurence: 5,
		Refresh: func() (interface{}, string, error) {
			if dis, err := staticIpsAreDisassociated(ctx, d, m); err != nil {
				return nil, "", fmt.Errorf("unable to check if static ips are disassociated: %w", err)
			} else if !dis {
				log.Printf("[DEBUG] still waiting for static ips to be disassociated")
				return struct{}{}, "doing", nil
			}
			return struct{}{}, "done", nil
		},
	}

	_, err := conf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to wait for for static ips to be dissassociated: %w", err)
	}
	return nil
}

// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func WaitForDeletion(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName := d.Get("project").(string), d.Get("service_name").(string)

	timeout := d.Timeout(schema.TimeoutDelete)
	log.Printf("[DEBUG] Service deletion waiter timeout %.0f minutes", timeout.Minutes())

	conf := &resource.StateChangeConf{
		Pending:                   []string{"deleting"},
		Target:                    []string{"deleted"},
		Delay:                     10 * time.Second,
		Timeout:                   timeout,
		MinTimeout:                20 * time.Second,
		ContinuousTargetOccurence: 5,
		Refresh: func() (interface{}, string, error) {
			_, err := client.Services.Get(ctx, projectName, serviceName)
			if err != nil && !aiven.IsNotFound(err) {
				return nil, "", fmt.Errorf("unable to check if service is gone: %w", err)
			}

			log.Printf("[DEBUG] service gone, still waiting for static ips to be disassociated")

			if dis, err := staticIpsDisassociatedAfterServiceDeletion(ctx, d, m); err != nil {
				return nil, "", fmt.Errorf("unable to check if static ips are disassociated: %w", err)
			} else if !dis {
				return struct{}{}, "deleting", nil
			}

			return struct{}{}, "deleted", nil
		},
	}

	if _, err := conf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("unable to wait for service deletion: %w", err)
	}
	return nil
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

	if service.Type == "redis" && service.UserConfig["redis_persistence"] == "off" {
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

// staticIpsReady checks that the static ips that are associated with the service are either
// in state 'assigned' or 'available'
func staticIpsReady(ctx context.Context, d *schema.ResourceData, m interface{}) (bool, error) {
	expectedStaticIps := staticIpsForServiceFromSchema(d)
	if len(expectedStaticIps) == 0 {
		return true, nil
	}

	client := m.(*aiven.Client)
	projectName, serviceName := d.Get("project").(string), d.Get("service_name").(string)

	staticIpsList, err := client.StaticIPs.List(ctx, projectName)
	if err != nil {
		return false, fmt.Errorf("unable to fetch static ips for project '%s': '%w", projectName, err)
	}

L:
	for _, eip := range expectedStaticIps {
		for _, sip := range staticIpsList.StaticIPs {
			assignedOrAvailable := sip.State == StaticIPAssigned || sip.State == StaticIPAvailable
			belongsToService := sip.ServiceName == serviceName
			isExpectedIP := sip.StaticIPAddressID == eip

			if isExpectedIP && belongsToService && assignedOrAvailable {
				continue L
			}
		}
		return false, nil
	}

	return true, nil
}

// staticIpsDisassociatedAfterServiceDeletion checks that after service deletion
// all static ips that were associated to the service are available again
func staticIpsDisassociatedAfterServiceDeletion(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) (bool, error) {
	expectedStaticIps := staticIpsForServiceFromSchema(d)
	if len(expectedStaticIps) == 0 {
		return true, nil
	}

	client := m.(*aiven.Client)
	projectName := d.Get("project").(string)

	staticIpsList, err := client.StaticIPs.List(ctx, projectName)
	if err != nil {
		return false, fmt.Errorf("unable to fetch static ips for project '%s': '%w", projectName, err)
	}

	for _, eip := range expectedStaticIps {
		for _, sip := range staticIpsList.StaticIPs {
			// no check for service name since after deletion the field is gone, but the
			// static ip lingers in the assigned state for a while until it gets usable again
			ipIsAssigned := sip.State == StaticIPAssigned
			isExpectedIP := sip.StaticIPAddressID == eip

			if isExpectedIP && ipIsAssigned {
				return false, nil
			}
		}
	}
	return true, nil
}

// staticIpsAreDisassociated checks that after service update
// all static ips that are not used by the service anymore are available again
func staticIpsAreDisassociated(ctx context.Context, d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	staticIpsList, err := client.StaticIPs.List(ctx, projectName)
	if err != nil {
		return false, fmt.Errorf("unable to fetch static ips for project '%s': '%w", projectName, err)
	}
	currentStaticIps := staticIpsForServiceFromSchema(d)
L:
	for _, sip := range staticIpsList.StaticIPs {
		ipBelongsToService := sip.ServiceName == serviceName
		if !ipBelongsToService {
			continue L
		}
		for _, csip := range currentStaticIps {
			if sip.StaticIPAddressID == csip {
				continue L
			}
		}
		return false, nil
	}
	return true, nil
}

func staticIpsForServiceFromSchema(d *schema.ResourceData) []string {
	return FlattenToString(d.Get("static_ips").(*schema.Set).List())
}
