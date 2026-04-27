package projectvpc

import (
	"context"
	"fmt"
	"regexp"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/vpc"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var projectNameRegexp = regexp.MustCompile("^[a-zA-Z0-9_-]*$")

func NewDataSource() datasource.DataSource {
	return adapter.NewDataSource(adapter.DataSourceOptions{
		TypeName:         typeName,
		IDFields:         idFields(),
		Schema:           dataSourceSchema,
		SchemaInternal:   dataSourceSchemaInternal(),
		Read:             dataSourceRead,
		ValidateConfig:   validateDataSourceConfig,
		ConfigValidators: dataSourceConfigValidators,
	})
}

func dataSourceSchema(ctx context.Context) datasourceschema.Schema {
	return datasourceschema.Schema{
		Attributes: map[string]datasourceschema.Attribute{
			"id": datasourceschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource ID composed as: `project/project_vpc_id`.",
			},
			"project": datasourceschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifies the project this resource belongs to.",
				Optional:            true,
			},
			"cloud_name": datasourceschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The cloud provider and region where the service is hosted in the format `CLOUD_PROVIDER-REGION_NAME`. For example, `google-europe-west1` or `aws-us-east-2`.",
				Optional:            true,
			},
			"vpc_id": datasourceschema.StringAttribute{
				MarkdownDescription: "The ID of the VPC. This can be used to filter out the other VPCs if there are more than one for the project and cloud.",
				Optional:            true,
			},
			"network_cidr": datasourceschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Network address range used by the VPC. For example, `192.168.0.0/24`.",
			},
			"state": datasourceschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: userconfig.Desc("State of the VPC.").PossibleValuesString(vpc.VpcStateTypeChoices()...).Build(),
			},
		},
		Blocks:              map[string]datasourceschema.Block{"timeouts": timeouts.Block(ctx)},
		MarkdownDescription: "Gets information about the VPC for an Aiven project.",
	}
}

func dataSourceSchemaInternal() *adapter.Schema {
	return &adapter.Schema{
		Properties: map[string]*adapter.Schema{
			"id": {
				Computed: true,
				Type:     adapter.SchemaTypeString,
			},
			"project": {
				Computed: true,
				Type:     adapter.SchemaTypeString,
			},
			"cloud_name": {
				Computed: true,
				Type:     adapter.SchemaTypeString,
			},
			"vpc_id": {
				Type: adapter.SchemaTypeString,
			},
			"network_cidr": {
				Computed: true,
				Type:     adapter.SchemaTypeString,
			},
			"state": {
				Computed: true,
				Type:     adapter.SchemaTypeString,
			},
			"timeouts": {
				Properties: map[string]*adapter.Schema{
					"read": {Type: adapter.SchemaTypeString},
				},
				Type: adapter.SchemaTypeObject,
			},
		},
		Type: adapter.SchemaTypeObject,
	}
}

func dataSourceConfigValidators(_ context.Context, _ avngen.Client) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("cloud_name"),
			path.MatchRoot("vpc_id"),
		),
		datasourcevalidator.Conflicting(
			path.MatchRoot("project"),
			path.MatchRoot("vpc_id"),
		),
		datasourcevalidator.RequiredTogether(
			path.MatchRoot("project"),
			path.MatchRoot("cloud_name"),
		),
	}
}

func validateDataSourceConfig(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
	if v, ok := d.GetOk("project"); ok && v.(string) != "" && !projectNameRegexp.MatchString(v.(string)) {
		return fmt.Errorf("project name should be alphanumeric")
	}
	if v, ok := d.GetOk("vpc_id"); ok && v.(string) != "" {
		_, err := schemautil.SplitResourceID(v.(string), 2)
		if err != nil {
			return fmt.Errorf("invalid vpc_id, should have the following format {project_name}/{project_vpc_id}: %w", err)
		}
	}

	return nil
}

func dataSourceRead(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	var projectName, projectVPCID, cloudName string
	if v, ok := d.GetOk("vpc_id"); ok && v.(string) != "" {
		chunks, err := schemautil.SplitResourceID(v.(string), 2)
		if err != nil {
			return fmt.Errorf("error splitting vpc_id: %w", err)
		}
		projectName = chunks[0]
		projectVPCID = chunks[1]
	} else {
		projectName = d.Get("project").(string)
		cloudName = d.Get("cloud_name").(string)
	}

	vpcList, err := client.VpcList(ctx, projectName)
	if err != nil {
		return fmt.Errorf("error getting a list of project %q VPCs: %w", projectName, err)
	}

	projectVPC, err := getVPC(vpcList, projectVPCID, cloudName)
	if err != nil {
		return err
	}

	if err := d.Set("project", projectName); err != nil {
		return err
	}
	if err := d.Set("cloud_name", projectVPC.CloudName); err != nil {
		return err
	}
	if err := d.Set("network_cidr", projectVPC.NetworkCidr); err != nil {
		return err
	}
	if err := d.Set("state", projectVPC.State); err != nil {
		return err
	}
	return d.SetID(projectName, projectVPC.ProjectVpcId)
}

func getVPC(vpcList []vpc.VpcOut, vpcID, cloudName string) (projectVPC *vpc.VpcOut, err error) {
	if (vpcID == "") == (cloudName == "") {
		return nil, fmt.Errorf("provide exactly one: vpc_id or cloud_name")
	}

	for _, v := range vpcList {
		if v.ProjectVpcId == vpcID {
			return &v, nil
		}
		if v.CloudName != cloudName {
			continue
		}
		if projectVPC != nil {
			return nil, fmt.Errorf("multiple project VPC with cloud_name %q, use vpc_id instead", cloudName)
		}
		projectVPC = &v
	}

	if projectVPC == nil {
		err = fmt.Errorf("not found project VPC with cloud_name %q", cloudName)
	}

	return projectVPC, err
}
