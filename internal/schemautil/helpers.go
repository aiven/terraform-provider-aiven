package schemautil

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/docker/go-units"
)

// ResourceStateOrResourceDiff either *schema.ResourceState or *schema.ResourceDiff
type ResourceStateOrResourceDiff interface {
	GetOk(key string) (interface{}, bool)
	Get(key string) interface{}
}

// PlanParameters service plan aparameters
type PlanParameters struct {
	DiskSizeMBDefault int
	DiskSizeMBStep    int
	DiskSizeMBMax     int
}

func GetAPIServiceIntegrations(d ResourceStateOrResourceDiff) []aiven.NewServiceIntegration {
	var apiServiceIntegrations []aiven.NewServiceIntegration
	tfServiceIntegrations := d.Get("service_integrations")
	if tfServiceIntegrations != nil {
		tfServiceIntegrationList := tfServiceIntegrations.([]interface{})
		for _, definition := range tfServiceIntegrationList {
			definitionMap := definition.(map[string]interface{})
			sourceService := definitionMap["source_service_name"].(string)
			apiIntegration := aiven.NewServiceIntegration{
				IntegrationType: definitionMap["integration_type"].(string),
				SourceService:   &sourceService,
				UserConfig:      make(map[string]interface{}),
			}
			apiServiceIntegrations = append(apiServiceIntegrations, apiIntegration)
		}
	}
	return apiServiceIntegrations
}

func GetProjectVPCIdPointer(d ResourceStateOrResourceDiff) (*string, error) {
	vpcID := d.Get("project_vpc_id").(string)
	if len(vpcID) == 0 {
		return nil, nil
	}

	var vpcIDPointer *string

	parts := strings.SplitN(vpcID, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid project_vpc_id, should have the following format {project_name}/{project_vpc_id}")
	}

	p1 := parts[1]
	vpcIDPointer = &p1
	return vpcIDPointer, nil
}

func GetMaintenanceWindow(d ResourceStateOrResourceDiff) *aiven.MaintenanceWindow {
	dow := d.Get("maintenance_window_dow").(string)
	if dow == "never" {
		// `never` is not available in the API, but can be set on the backend
		// Sending this back to the backend will fail the validation
		return nil
	}
	t := d.Get("maintenance_window_time").(string)
	if len(dow) > 0 && len(t) > 0 {
		return &aiven.MaintenanceWindow{DayOfWeek: dow, TimeOfDay: t}
	}
	return nil
}

func ConvertToDiskSpaceMB(b string) int {
	diskSizeMB, _ := units.RAMInBytes(b)
	return int(diskSizeMB / units.MiB)
}

func GetServicePlanParametersFromServiceResponse(ctx context.Context, client *aiven.Client, project string, service *aiven.Service) (PlanParameters, error) {
	return getServicePlanParametersInternal(ctx, client, project, service.Type, service.Plan)
}

func GetServicePlanParametersFromSchema(ctx context.Context, client *aiven.Client, d ResourceStateOrResourceDiff) (PlanParameters, error) {
	project := d.Get("project").(string)
	serviceType := d.Get("service_type").(string)
	servicePlan := d.Get("plan").(string)

	return getServicePlanParametersInternal(ctx, client, project, serviceType, servicePlan)
}

func getServicePlanParametersInternal(_ context.Context, client *aiven.Client, project, serviceType, servicePlan string) (PlanParameters, error) {
	servicePlanResponse, err := client.ServiceTypes.GetPlan(project, serviceType, servicePlan)
	if err != nil {
		return PlanParameters{}, err
	}
	return PlanParameters{
		DiskSizeMBDefault: servicePlanResponse.DiskSpaceMB,
		DiskSizeMBMax:     servicePlanResponse.DiskSpaceCapMB,
		DiskSizeMBStep:    servicePlanResponse.DiskSpaceStepMB,
	}, nil
}

func dynamicDiskSpaceIsAllowedByPricing(_ context.Context, client *aiven.Client, d ResourceStateOrResourceDiff) (bool, error) {
	// to check if dynamic disk space is allowed, we currently have to check
	// the pricing api to see if the `extra_disk_price_per_gb_usd` field is set

	project := d.Get("project").(string)
	serviceType := d.Get("service_type").(string)
	servicePlan := d.Get("plan").(string)
	cloudName := d.Get("cloud_name").(string)

	servicePlanPricingResponse, err := client.ServiceTypes.GetPlanPricing(project, serviceType, servicePlan, cloudName)
	if err != nil {
		return false, fmt.Errorf("unable to get service plan pricing from api: %w", err)
	}
	return len(servicePlanPricingResponse.ExtraDiskPricePerGBUSD) > 0, nil
}

func HumanReadableByteSize(s int) string {
	// we only allow GiB as unit and show decimal places to MiB precision, this should fix rounding issues
	// when converting from mb to human readable and back
	var suffixes = []string{"B", "KiB", "MiB", "GiB"}

	return units.CustomSize("%.12g%s", float64(s), 1024.0, suffixes)
}
