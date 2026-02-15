package plugin

import (
	"context"
	"testing"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
)

func TestBetaResourcesRegistered(t *testing.T) {
	t.Setenv(util.AivenEnableBeta, "1")

	p := New("test").(*AivenProvider)
	resources := p.Resources(context.Background())

	names := make(map[string]bool)
	for _, fn := range resources {
		r := fn()
		resp := resource.MetadataResponse{}
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "aiven"}, &resp)
		names[resp.TypeName] = true
	}

	assert.True(t, names["aiven_byoc_aws_entity"], "aiven_byoc_aws_entity should be registered")
	assert.True(t, names["aiven_byoc_aws_provision"], "aiven_byoc_aws_provision should be registered")
	assert.True(t, names["aiven_byoc_permissions"], "aiven_byoc_permissions should be registered")
}
