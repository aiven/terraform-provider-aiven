package kafka

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"golang.org/x/sync/errgroup"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/service"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/userconfig/service/kafka"
)

var (
	_ resource.Resource                = &kafkaResource{}
	_ resource.ResourceWithConfigure   = &kafkaResource{}
	_ resource.ResourceWithImportState = &kafkaResource{}
)

func NewKafkaResource() resource.Resource {
	return &kafkaResource{}
}

type kafkaResource struct {
	client *aiven.Client
	crud   service.Crud
}

func (s *kafkaResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	s.client = req.ProviderData.(*aiven.Client)
	s.crud = service.NewCrud(s.client, service.TypeKafka)
}

func (s *kafkaResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aiven_kafka"
}

func (s *kafkaResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = service.WithServiceCommon(ctx, schema.Schema{
		Description: "The Kafka resource allows the creation and management of Aiven Kafka services.",
		Attributes: map[string]schema.Attribute{
			"default_acl": &schema.BoolAttribute{
				Description: "Create default wildcard Kafka ACL",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
				Default: booldefault.StaticBool(true),
				//DiffSuppressFunc: schemautil.CreateOnlyDiffSuppressFunc,
			},
		},
		Blocks: map[string]schema.Block{
			"kafka": schema.SetNestedBlock{
				Description: "Kafka server provided values",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"access_cert": schema.StringAttribute{
							Description: "The Kafka client certificate",
							Computed:    true,
							Sensitive:   true,
						},
						"access_key": schema.StringAttribute{
							Description: "The Kafka client certificate key",
							Computed:    true,
							Sensitive:   true,
						},
						"connect_uri": schema.StringAttribute{
							Description: "The Kafka Connect URI, if any",
							Computed:    true,
							Sensitive:   true,
						},
						"rest_uri": schema.StringAttribute{
							Description: "The Kafka REST URI, if any",
							Computed:    true,
							Sensitive:   true,
						},
						"schema_registry_uri": schema.StringAttribute{
							Description: "The Schema Registry URI, if any",
							Computed:    true,
							Sensitive:   true,
						},
					},
				},
			},
			"kafka_user_config": kafka.NewResourceSchema(),
		},
	})
}

func (s *kafkaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	o := new(kafkaModel)
	s.crud.Create(ctx, req, resp, o)
	if resp.Diagnostics.HasError() {
		return
	}

	if !o.DefaultACL.ValueBool() {
		err := s.deleteDefaultACL(ctx, o.Project.ValueString(), o.ServiceName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("plan copy error", err.Error())
			return
		}
	}
}

func (s *kafkaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	o := new(kafkaModel)
	s.crud.Read(ctx, req, resp, o)
}

// Update todo
func (s *kafkaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	s.crud.Update(ctx, req, resp, new(kafkaModel))
}

func (s *kafkaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	s.crud.Delete(ctx, req, resp, new(kafkaModel))
}

func (s *kafkaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s.crud.ImportState(ctx, req, resp, new(kafkaModel))
}

func (s *kafkaResource) deleteDefaultACL(ctx context.Context, project, serviceName string) error {
	toDelete := map[string]func(ctx context.Context, project, serviceName, toDelete string) error{
		"default":                  s.client.KafkaACLs.Delete,
		"default-sr-admin-config":  s.client.KafkaSchemaRegistryACLs.Delete,
		"default-sr-admin-subject": s.client.KafkaSchemaRegistryACLs.Delete,
	}

	g, gCtx := errgroup.WithContext(ctx)
	for key := range toDelete {
		k := key
		f := toDelete[k]
		g.Go(func() error {
			err := f(gCtx, project, serviceName, k)
			if aiven.IsNotFound(err) {
				return nil
			}
			return err
		})
	}
	return g.Wait()
}
