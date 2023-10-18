package service

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/common"
)

var reProjectServicePath = regexp.MustCompile(`^[a-zA-Z0-9_-]*/[a-zA-Z0-9_-]*$`)

func WithServiceCommon(ctx context.Context, s schema.Schema) schema.Schema {
	base := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						reProjectServicePath,
						"id should have the following format: project/service_name",
					),
				},
			},
			"project": common.ProjectString(),
			"cloud_name": schema.StringAttribute{
				Description: "Defines where the cloud provider and region where the service is hosted in. This can be changed freely after service is created. Changing the value will trigger a potentially lengthy migration process for the service. Format is cloud provider name (`aws`, `azure`, `do` `google`, `upcloud`, etc.), dash, and the cloud provider specific region name. These are documented on each Cloud provider's own support articles, like [here for Google](https://cloud.google.com/compute/docs/regions-zones/) and [here for AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html).",
				Optional:    true,
			},
			"plan": schema.StringAttribute{
				Description: "Defines what kind of computing resources are allocated for the service. It can be changed after creation, though there are some restrictions when going to a smaller plan such as the new plan must have sufficient amount of disk space to store all current data and switching to a plan with fewer nodes might not be supported. The basic plan names are `hobbyist`, `startup-x`, `business-x` and `premium-x` where `x` is (roughly) the amount of memory on each node (also other attributes like number of CPUs and amount of disk space varies but naming is based on memory). The available options can be seem from the [Aiven pricing page](https://aiven.io/pricing).",
				Required:    true,
			},
			"service_name": schema.StringAttribute{
				Description: "Specifies the actual name of the service. The name cannot be changed later without destroying and re-creating the service so name should be picked based on intended service usage rather than current attributes.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_type": schema.StringAttribute{
				Description: "Aiven internal service type code",
				Computed:    true,
			},
			"project_vpc_id": schema.StringAttribute{
				Description: "Specifies the VPC the service should run in. If the value is not set the service is not run inside a VPC. When set, the value should be given as a reference to set up dependencies correctly and the VPC must be in the same cloud and region as the service itself. Project can be freely moved to and from VPC after creation but doing so triggers migration to new servers so the operation can take significant amount of time to complete if the service has a lot of data.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						reProjectServicePath,
						"invalid project_vpc_id, should have the following format {project_name}/{project_vpc_id}",
					),
				},
			},
			"maintenance_window_dow": schema.StringAttribute{
				Description: "Day of week when maintenance operations should be performed. One monday, tuesday, wednesday, etc.",
				Optional:    true,
				//DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				//	return new == ""
				//},
				//// There is also `never` value, which can't be set, but can be received from the backend.
				//// Sending `never` is suppressed in GetMaintenanceWindow function,
				//// but then we need to not let to set `never` manually
				//ValidateFunc: validation.StringInSlice([]string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}, false),
			},
			"maintenance_window_time": schema.StringAttribute{
				Description: "Time of day when maintenance operations should be performed. UTC time in HH:mm:ss format.",
				Optional:    true,
			},
			"termination_protection": schema.BoolAttribute{
				Description: "Prevents the service from being deleted. It is recommended to set this to `true` for all production services to prevent unintentional service deletion. This does not shield against deleting databases or topics but for services with backups much of the content can at least be restored from backup in case accidental deletion is done.",
				Optional:    true,
			},
			"disk_space_used": schema.StringAttribute{
				Computed:    true,
				Description: "Disk space that service is currently using",
			},
			"disk_space_default": schema.StringAttribute{
				Description: "The default disk space of the service, possible values depend on the service type, the cloud provider and the project. Its also the minimum value for `disk_space`",
				Computed:    true,
			},
			"additional_disk_space": schema.StringAttribute{
				Description: "Additional disk space. Possible values depend on the service type, the cloud provider and the project. Therefore, reducing will result in the service rebalancing.",
				Optional:    true,
				//ValidateFunc:  ValidateHumanByteSizeString,
			},
			"disk_space_step": schema.StringAttribute{
				Description: "The default disk space step of the service, possible values depend on the service type, the cloud provider and the project. `disk_space` needs to increment from `disk_space_default` by increments of this size.",
				Computed:    true,
			},
			"disk_space_cap": schema.StringAttribute{
				Description: "The maximum disk space of the service, possible values depend on the service type, the cloud provider and the project.",
				Computed:    true,
			},
			"service_uri": schema.StringAttribute{
				Description: "URI for connecting to the service. Service specific info is under \"kafka\", \"pg\", etc.",
				Computed:    true,
				Sensitive:   true,
			},
			"service_host": schema.StringAttribute{
				Description: "The hostname of the service.",
				Computed:    true,
			},
			"service_port": schema.Int64Attribute{
				Description: "The port of the service",
				Computed:    true,
			},
			"service_password": schema.StringAttribute{
				Description: "Password used for connecting to the service, if applicable",
				Computed:    true,
				Sensitive:   true,
			},
			"service_username": schema.StringAttribute{
				Description: "Username used for connecting to the service, if applicable",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "Service state. One of `POWEROFF`, `REBALANCING`, `REBUILDING` or `RUNNING`",
				Computed:    true,
			},
			"static_ips": schema.SetAttribute{
				Description: "Static IPs that are going to be associated with this service. Please assign a value using the 'toset' function. Once a static ip resource is in the 'assigned' state it cannot be unbound from the node again",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"components": schema.SetNestedBlock{
				Description: "Service component information objects",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"component": schema.StringAttribute{
							Description: "Service component name",
							Computed:    true,
						},
						"host": schema.StringAttribute{
							Description: "DNS name for connecting to the service component",
							Computed:    true,
						},
						"kafka_authentication_method": schema.StringAttribute{
							Description: "Kafka authentication method. This is a value specific to the 'kafka' service component",
							Computed:    true,
						},
						"port": schema.Int64Attribute{
							Description: "Port number for connecting to the service component",
							Computed:    true,
						},
						"route": schema.StringAttribute{
							Description: "Network access route",
							Computed:    true,
						},
						"ssl": schema.BoolAttribute{
							Description: "Whether the endpoint is encrypted or accepts plaintext. By default endpoints are " +
								"always encrypted and this property is only included for service components they may " +
								"disable encryption",
							Computed: true,
						},
						"usage": schema.StringAttribute{
							Description: "DNS usage name",
							Computed:    true,
						},
					},
				},
			},
			"service_integrations": schema.SetNestedBlock{
				Description: "Service integrations to specify when creating a service. Not applied after initial service creation",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"source_service_name": schema.StringAttribute{
							Description: "Name of the source service",
							Required:    true,
						},
						"integration_type": schema.StringAttribute{
							Description: "Type of the service integration. The only supported value at the moment is `read_replica`",
							Required:    true,
						},
					},
				},
			},
			"tag": schema.SetNestedBlock{
				Description: "Tags are key-value pairs that allow you to categorize services.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "Service tag key",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Service tag value",
							Required:    true,
						},
					},
				},
			},
		},
	}
	return common.MergeSchemas(s, common.WithDefaultTimeouts(ctx, base))
}
