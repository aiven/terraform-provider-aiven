package kafka

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util/serviceutil"
)

var svc = serviceutil.Service{Type: "kafka"}

var (
	expandModifier = svc.ExpandModifier
	modifyPlan     = svc.ModifyPlan
)

func init() {
	svc.Wrap(&ResourceOptions, &DataSourceOptions)
	create := ResourceOptions.Create
	ResourceOptions.Create = func(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
		if err := create(ctx, client, d); err != nil || d.Get("default_acl").(bool) {
			return err
		}

		// default_acl = false: deletes the ACLs a new service has, i.e., a wildcard Kafka ACL and Schema Registry ones
		project, name := d.Get("project").(string), d.Get("service_name").(string)
		if _, err := client.ServiceKafkaAclDelete(ctx, project, name, "default"); common.IsCritical(err) {
			return fmt.Errorf("cannot delete default wildcard kafka acl: %w", err)
		}

		for _, id := range []string{"default-sr-admin-config", "default-sr-admin-subject"} {
			if _, err := client.ServiceSchemaRegistryAclDelete(ctx, project, name, id); common.IsCritical(err) {
				return fmt.Errorf("cannot delete `%s` kafka ACL for Schema Registry: %w", id, err)
			}
		}
		return nil
	}
}

// flattenModifier sets default_acl, which isn't in the API, for an imported service.
func flattenModifier(ctx context.Context, client avngen.Client) adapter.MapModifier {
	return adapter.ComposeMapModifiers(svc.FlattenModifier(ctx, client), func(d adapter.ResourceData, dto map[string]any) error {
		if _, ok := d.GetOk("default_acl"); !ok {
			dto["default_acl"] = true
		}
		return nil
	})
}
