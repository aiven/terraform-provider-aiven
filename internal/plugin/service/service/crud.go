package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/docker/go-units"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

type Crud interface {
	Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, service any)
	Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse, service any)
	Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse, service any)
	Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse, service any)
	ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse, service any)
}

var _ Crud = &crud{}

func NewCrud(client *aiven.Client, serviceType string) Crud {
	return &crud{client: client, serviceType: serviceType}
}

type crud struct {
	client      *aiven.Client
	serviceType string
}

func (c *crud) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse, service any) {
	parts, ok := common.SplitN(req.ID, 2)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: project/service_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_name"), parts[1])...)
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (c *crud) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, service any) {
	o := new(Resource)
	commit := StateTransaction(ctx, req, resp, service, o)
	if resp.Diagnostics.HasError() {
		return
	}
	defer commit()

	config := expandUserConfig(ctx, resp.Diagnostics, o, true)
	if resp.Diagnostics.HasError() {
		return
	}

	staticIPs := schemautil.ExpandSet[string](ctx, resp.Diagnostics, o.StaticIPs)
	if resp.Diagnostics.HasError() {
		return
	}

	integrations := schemautil.ExpandSetNested(ctx, resp.Diagnostics, expandServiceIntegration, o.ServiceIntegrations)
	if resp.Diagnostics.HasError() {
		return
	}

	disk := c.diskSpace(ctx, resp.Diagnostics, o)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := c.client.Services.Create(
		ctx,
		o.Project.ValueString(),
		aiven.CreateServiceRequest{
			ServiceType:           c.serviceType,
			Cloud:                 o.CloudName.ValueString(),
			Plan:                  o.Plan.ValueString(),
			ProjectVPCID:          common.PathIndexPointer(o.ProjectVPCId.ValueString(), common.ProjectIndex),
			ServiceIntegrations:   fromPointers(integrations),
			MaintenanceWindow:     getMaintenanceWindow(o),
			ServiceName:           o.ServiceName.ValueString(),
			TerminationProtection: o.TerminationProtection.ValueBool(),
			DiskSpaceMB:           disk.CalcDiskSpace(),
			UserConfig:            config,
			StaticIPs:             staticIPs,
		},
	)

	if err == nil {
		err = common.WaitActive(ctx, func() error {
			s, err := c.client.Services.Get(ctx, o.Project.ValueString(), o.ServiceName.ValueString())
			if err != nil {
				return err
			}

			if s.State != "RUNNING" {
				return fmt.Errorf("service state is not running: %q", s.State)
			}
			return nil
		})
	}

	if err != nil {
		resp.Diagnostics.AddError(errmsg.SummaryErrorCreatingResource, err.Error())
		return
	}
	c.read(ctx, resp.Diagnostics, o)
}

func (c *crud) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse, service any) {
	o := new(Resource)
	commit := StateTransaction(ctx, req, resp, service, o)
	if resp.Diagnostics.HasError() {
		return
	}
	defer commit()
	c.read(ctx, resp.Diagnostics, o)
}

func (c *crud) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse, service any) {
	//TODO implement me
	panic("implement me")
}

func (c *crud) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse, service any) {
	o := new(Resource)
	commit := StateTransaction(ctx, req, resp, service, o)
	if resp.Diagnostics.HasError() {
		return
	}
	defer commit()

	err := c.client.Services.Delete(ctx, o.Project.ValueString(), o.ServiceName.ValueString())
	if err != nil && !aiven.IsNotFound(err) {
		resp.Diagnostics.AddError(errmsg.SummaryErrorDeletingResource, err.Error())
		return
	}
}

func (c *crud) read(ctx context.Context, diags diag.Diagnostics, o *Resource) {
	project := o.Project.ValueString()
	serviceName := o.ServiceName.ValueString()
	s, err := c.client.Services.Get(ctx, project, serviceName)
	if err != nil {
		diags.AddError(errmsg.SummaryErrorReadingResource, err.Error())
		return
	}

	o.ServiceType = types.StringValue(c.serviceType)
	flattenUserConfig(ctx, diags, o, s)
	if diags.HasError() {
		return
	}

	integrations := schemautil.FlattenSetNested(ctx, diags, flattenServiceIntegration, s.Integrations, serviceIntegrationAttrs)
	if diags.HasError() {
		return
	}

	tagsRsp, err := c.client.ServiceTags.Get(ctx, project, serviceName)
	if err != nil {
		diags.AddError("Unable to get service tags", err.Error())
		return
	}

	tags := flattenTags(ctx, diags, tagsRsp.Tags)
	if diags.HasError() {
		return
	}

	staticIPsRsp, err := c.client.StaticIPs.List(ctx, project)
	if err != nil {
		diags.AddError(fmt.Sprintf("Unable to list static ips for project '%s'", project), err.Error())
		return
	}

	staticIPs, d := types.SetValueFrom(ctx, types.StringType, staticIPsRsp.StaticIPs)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	o.ID = types.StringValue(filepath.Join(project, serviceName))
	o.CloudName = types.StringValue(s.CloudName)
	o.MaintenanceWindowDow = types.StringValue(s.MaintenanceWindow.DayOfWeek)
	o.MaintenanceWindowTime = types.StringValue(s.MaintenanceWindow.TimeOfDay)
	o.Plan = types.StringValue(s.Plan)
	o.ProjectVPCId = types.StringPointerValue(s.ProjectVPCID)
	o.Project = types.StringValue(project)
	o.ServiceHost = types.StringValue(s.URIParams["host"])
	if p := s.URIParams["password"]; p != "" {
		o.ServicePassword = types.StringValue(p)
	}
	if u := s.URIParams["user"]; u != "" {
		o.ServiceUsername = types.StringValue(u)
	}
	port, err := strconv.ParseInt(s.URIParams["port"], 10, 32)
	if err != nil {
		diags.AddError("Can't parse port", err.Error())
		return
	}
	o.ServicePort = types.Int64Value(port)
	o.ServiceName = types.StringValue(serviceName)
	o.ServiceURI = types.StringValue(s.URI)
	o.ServiceIntegrations = integrations
	o.State = types.StringValue(s.State)
	o.StaticIPs = staticIPs
	o.Tag = tags
	o.TerminationProtection = types.BoolValue(s.TerminationProtection)

	disk := c.diskSpace(ctx, diags, o)
	if diags.HasError() {
		return
	}
	o.DiskSpaceUsed = disk.String(s.DiskSpaceMB) // Actual service disk size
	o.AdditionalDiskSpace = disk.String(s.DiskSpaceMB - disk.AdditionalDiskSpace())
	o.DiskSpaceCap = disk.String(disk.plan.DiskSpaceCapMB)
	o.DiskSpaceDefault = disk.String(disk.plan.DiskSpaceMB)
	o.DiskSpaceStep = disk.String(disk.plan.DiskSpaceStepMB)
}

func (c *crud) diskSpace(ctx context.Context, diags diag.Diagnostics, o *Resource) *diskSpace {
	plan, err := c.client.ServiceTypes.GetPlan(ctx, o.Project.ValueString(), c.serviceType, o.Plan.ValueString())
	if err != nil {
		diags.AddError("Error getting service default plan parameters", err.Error())
		return nil
	}
	return &diskSpace{plan: plan, resource: o}
}

type diskSpace struct {
	plan     *aiven.GetServicePlanResponse
	resource *Resource
}

func (d *diskSpace) String(size int) types.String {
	return types.StringValue(schemautil.HumanReadableByteSize(size * units.MiB))
}

func (d *diskSpace) AdditionalDiskSpace() int {
	return inMiB(d.resource.AdditionalDiskSpace.ValueString())
}

func (d *diskSpace) CalcDiskSpace() int {
	return d.plan.DiskSpaceMB + d.AdditionalDiskSpace()
}

type fakeState struct{}

func (f *fakeState) Get(ctx context.Context, target any) diag.Diagnostics {
	return diag.Diagnostics{}
}
func (f *fakeState) Set(ctx context.Context, target any) diag.Diagnostics {
	return diag.Diagnostics{}
}

type getType interface {
	Get(ctx context.Context, target any) diag.Diagnostics
}

type setType interface {
	Set(ctx context.Context, target any) diag.Diagnostics
}

func StateTransaction(ctx context.Context, req, rsp, service, generic any) func() {
	var diags diag.Diagnostics
	var getter getType = &fakeState{}
	var setter setType = &fakeState{}

	switch req.(type) {
	case resource.CreateRequest:
		rqst := req.(resource.CreateRequest)
		resp := rsp.(*resource.CreateResponse)
		getter = rqst.Plan
		setter = &resp.State
		diags = resp.Diagnostics
	case resource.ReadRequest:
		rqst := req.(resource.ReadRequest)
		resp := rsp.(*resource.ReadResponse)
		getter = rqst.State
		setter = &resp.State
		diags = resp.Diagnostics
	case resource.UpdateRequest:
		rqst := req.(resource.UpdateRequest)
		resp := rsp.(*resource.UpdateResponse)
		getter = rqst.Plan
		setter = &resp.State
		diags = resp.Diagnostics
	case resource.DeleteRequest:
		rqst := req.(resource.DeleteRequest)
		resp := rsp.(*resource.DeleteResponse)
		getter = rqst.State
		setter = &resp.State
		diags = resp.Diagnostics
	case resource.ImportStateRequest:
		resp := rsp.(*resource.ImportStateResponse)
		diags = resp.Diagnostics
	}

	diags.Append(getter.Get(ctx, service)...)
	if diags.HasError() {
		return nil
	}

	copyBack := common.Copy(service, generic)
	return func() {
		if diags.HasError() {
			return
		}
		copyBack()
		diags.Append(setter.Set(ctx, service)...)
	}
}
