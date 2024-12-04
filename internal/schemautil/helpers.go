package schemautil

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/docker/go-units"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ResourceStateOrResourceDiff either *schema.ResourceState or *schema.ResourceDiff
type ResourceStateOrResourceDiff interface {
	GetRawConfig() cty.Value
	GetOk(key string) (interface{}, bool)
	Get(key string) interface{}
}

func HasConfigValue(d ResourceStateOrResourceDiff, key string) bool {
	c := d.GetRawConfig()
	return !(c.IsNull() || c.AsValueMap()[key].IsNull())
}

// PlanParameters service plan aparameters
type PlanParameters struct {
	DiskSizeMBDefault int
	DiskSizeMBStep    int
	DiskSizeMBMax     int
}

func GetAPIServiceIntegrations(d ResourceStateOrResourceDiff) []service.ServiceIntegrationIn {
	tfServiceIntegrations := d.Get("service_integrations").(*schema.Set).List()
	apiServiceIntegrations := make([]service.ServiceIntegrationIn, 0, len(tfServiceIntegrations))
	for _, definition := range tfServiceIntegrations {
		definitionMap := definition.(map[string]interface{})
		sourceService := definitionMap["source_service_name"].(string)
		userConfig := make(map[string]any)
		integrationType := definitionMap["integration_type"].(string)
		apiIntegration := service.ServiceIntegrationIn{
			IntegrationType: service.IntegrationType(integrationType),
			SourceService:   &sourceService,
			UserConfig:      &userConfig,
		}
		apiServiceIntegrations = append(apiServiceIntegrations, apiIntegration)
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

func GetMaintenanceWindow(d ResourceStateOrResourceDiff) *service.MaintenanceIn {
	dow := d.Get("maintenance_window_dow").(string)
	if dow == "never" {
		// `never` is not available in the API, but can be set on the backend
		// Sending this back to the backend will fail the validation
		return nil
	}
	t := d.Get("maintenance_window_time").(string)
	if len(dow) > 0 && len(t) > 0 {
		return &service.MaintenanceIn{Dow: service.DowType(dow), Time: &t}
	}
	return nil
}

func ConvertToDiskSpaceMB(b string) int {
	diskSizeMB, _ := units.RAMInBytes(b)
	return int(diskSizeMB / units.MiB)
}

func GetServicePlanParametersFromServiceResponse(ctx context.Context, client *aiven.Client, project string, s *service.ServiceGetOut) (PlanParameters, error) {
	return getServicePlanParametersInternal(ctx, client, project, s.ServiceType, s.Plan)
}

func GetServicePlanParametersFromSchema(ctx context.Context, client *aiven.Client, d ResourceStateOrResourceDiff) (PlanParameters, error) {
	project := d.Get("project").(string)
	serviceType := d.Get("service_type").(string)
	servicePlan := d.Get("plan").(string)

	return getServicePlanParametersInternal(ctx, client, project, serviceType, servicePlan)
}

// fixme: GetPlane is not available in the generated client
func getServicePlanParametersInternal(ctx context.Context, client *aiven.Client, project, serviceType, servicePlan string) (PlanParameters, error) {
	servicePlanResponse, err := client.ServiceTypes.GetPlan(ctx, project, serviceType, servicePlan)
	if err != nil {
		return PlanParameters{}, err
	}
	return PlanParameters{
		DiskSizeMBDefault: servicePlanResponse.DiskSpaceMB,
		DiskSizeMBMax:     servicePlanResponse.DiskSpaceCapMB,
		DiskSizeMBStep:    servicePlanResponse.DiskSpaceStepMB,
	}, nil
}

// fixme: GetPlanPricing is not available in the generated client
func dynamicDiskSpaceIsAllowedByPricing(ctx context.Context, client *aiven.Client, d ResourceStateOrResourceDiff) (bool, error) {
	// to check if dynamic disk space is allowed, we currently have to check
	// the pricing api to see if the `extra_disk_price_per_gb_usd` field is set

	project := d.Get("project").(string)
	serviceType := d.Get("service_type").(string)
	servicePlan := d.Get("plan").(string)
	cloudName := d.Get("cloud_name").(string)

	servicePlanPricingResponse, err := client.ServiceTypes.GetPlanPricing(ctx, project, serviceType, servicePlan, cloudName)
	if err != nil {
		return false, fmt.Errorf("unable to get service plan pricing from api: %w", err)
	}
	return len(servicePlanPricingResponse.ExtraDiskPricePerGBUSD) > 0, nil
}

func HumanReadableByteSize(s int) string {
	// we only allow GiB as unit and show decimal places to MiB precision, this should fix rounding issues
	// when converting from mb to human readable and back
	suffixes := []string{"B", "KiB", "MiB", "GiB"}

	return units.CustomSize("%.12g%s", float64(s), 1024.0, suffixes)
}

// isStringAnOrganizationID is a helper function that returns true if the string is an organization ID.
func isStringAnOrganizationID(s string) bool {
	return strings.HasPrefix(s, "org")
}

// NormalizeOrganizationID is a helper function that returns the ID to use for the API call.
// If the ID is an organization ID, it will be converted to an account ID via the API.
// If the ID is an account ID, it will be returned as is, without performing any API calls.
func NormalizeOrganizationID(ctx context.Context, client *aiven.Client, id string) (string, error) {
	if isStringAnOrganizationID(id) {
		r, err := client.Organization.Get(ctx, id)
		if err != nil {
			return "", err
		}

		id = r.AccountID
	}

	return id, nil
}

// DetermineMixedOrganizationConstraintIDToStore is a helper function that returns the ID to store in the state.
// We have several fields that can be either an organization ID or an account ID.
// We want to store the one that was already in the state, if it was already there.
// If it was not, we want to prioritize the organization ID, but if it is not available, we want to store the account
// ID.
// If the ID is an account ID, it will be returned as is, without performing any API calls.
// If the ID is an organization ID, it will be refreshed via the provided account ID and returned.
func DetermineMixedOrganizationConstraintIDToStore(
	ctx context.Context,
	client *aiven.Client,
	stateID string,
	accountID string,
) (string, error) {
	if len(accountID) == 0 {
		return "", nil
	}

	if !isStringAnOrganizationID(stateID) {
		return accountID, nil
	}

	r, err := client.Accounts.Get(ctx, accountID)
	if err != nil {
		return "", err
	}

	if len(r.Account.OrganizationId) == 0 {
		return accountID, nil
	}

	return r.Account.OrganizationId, nil
}

// StringToDiagWarning is a function that converts a string to a diag warning.
func StringToDiagWarning(msg string) diag.Diagnostics {
	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  msg,
	}}
}

// ErrorToDiagWarning is a function that converts an error to a diag warning.
func ErrorToDiagWarning(err error) diag.Diagnostics {
	return StringToDiagWarning(err.Error())
}
