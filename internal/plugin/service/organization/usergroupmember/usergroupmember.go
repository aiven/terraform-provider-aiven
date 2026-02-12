package usergroupmember

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/usergroup"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return upsert(ctx, client, d, usergroup.OperationTypeAddMembers)
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return upsert(ctx, client, d, usergroup.OperationTypeRemoveMembers)
}

func upsert(ctx context.Context, client avngen.Client, d adapter.ResourceData, operation usergroup.OperationType) error {
	err := client.UserGroupMembersUpdate(
		ctx,
		d.Get("organization_id").(string),
		d.Get("group_id").(string),
		&usergroup.UserGroupMembersUpdateIn{
			Operation: operation,
			MemberIds: []string{d.Get("user_id").(string)},
		},
	)
	if err != nil {
		return fmt.Errorf("operation %s failed: %w", operation, err)
	}

	return nil
}
