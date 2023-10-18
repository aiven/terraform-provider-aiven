package kafka

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/service"
)

var (
	_ datasource.DataSource              = &kafkaDataSource{}
	_ datasource.DataSourceWithConfigure = &kafkaDataSource{}
)

func NewKafkaDataSource() datasource.DataSource {
	return &kafkaDataSource{}
}

type kafkaDataSource struct {
	client *aiven.Client
	crud   service.Crud
}

func (s *kafkaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	s.client = req.ProviderData.(*aiven.Client)
	s.crud = service.NewCrud(s.client, service.TypeKafka)
}

func (s *kafkaDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "aiven_kafka"
}

func (s *kafkaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Service Integration data source provides information about the existing Aiven Service Integration.",
		Attributes:  map[string]schema.Attribute{},
		Blocks:      map[string]schema.Block{},
	}
}

// Read reads datasource
// All functions adapted for kafkaModel, so we use it as donor
// Copies state from datasource to resource, then back, when things are done
func (s *kafkaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// fixme: replace with datasource
	//s.crud.Read(ctx, req, resp, new(kafkaModel))
}
