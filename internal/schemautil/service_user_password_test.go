package schemautil

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/go-cty/cty"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/aiven/terraform-provider-aiven/mocks"
)

func TestShouldResetPassword(t *testing.T) {
	t.Parallel()

	t.Run("new resources always require reset", func(t *testing.T) {
		t.Run("with write-only password", func(t *testing.T) {
			d := mocks.NewMockResourceData(t)
			d.EXPECT().Id().Return("") // empty ID means new resource

			d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.StringVal("WriteOnly123!"),
				"password_wo_version": cty.NumberIntVal(1),
			}))

			password, shouldReset := shouldResetPassword(d)

			assert.Equal(t, "WriteOnly123!", password)
			assert.True(t, shouldReset, "new resource should require reset")
		})

		t.Run("with optional password field", func(t *testing.T) {
			d := mocks.NewMockResourceData(t)
			d.EXPECT().Id().Return("")

			d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.NullVal(cty.String),
				"password_wo_version": cty.NullVal(cty.Number),
			}))
			d.EXPECT().GetOk("password").Return("Optional123!", true)

			password, shouldReset := shouldResetPassword(d)

			assert.Equal(t, "Optional123!", password)
			assert.True(t, shouldReset, "new resource should require reset")
		})

		t.Run("with auto-generated password", func(t *testing.T) {
			d := mocks.NewMockResourceData(t)
			d.EXPECT().Id().Return("")

			d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.NullVal(cty.String),
				"password_wo_version": cty.NullVal(cty.Number),
			}))
			d.EXPECT().GetOk("password").Return(nil, false)

			password, shouldReset := shouldResetPassword(d)

			assert.Empty(t, password, "empty password triggers auto-generation")
			assert.True(t, shouldReset, "new resource should require reset")
		})
	})

	t.Run("existing resource with no changes", func(t *testing.T) {
		d := mocks.NewMockResourceData(t)
		d.EXPECT().Id().Return("proj/svc/user") // non-empty ID means existing resource

		d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
			"password_wo":         cty.NullVal(cty.String),
			"password_wo_version": cty.NullVal(cty.Number),
		}))
		d.EXPECT().GetOk("password").Return("Existing123!", true)
		d.EXPECT().HasChange("password").Return(false)
		d.EXPECT().HasChange("password_wo_version").Return(false)

		password, shouldReset := shouldResetPassword(d)

		assert.Equal(t, "Existing123!", password)
		assert.False(t, shouldReset, "no changes means no reset needed")
	})

	t.Run("existing resource password updates", func(t *testing.T) {
		t.Parallel()

		t.Run("optional password field changed", func(t *testing.T) {
			d := mocks.NewMockResourceData(t)
			d.EXPECT().Id().Return("proj/svc/user")

			// user changed password from old value to "NewPassword123!"
			d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.NullVal(cty.String),
				"password_wo_version": cty.NullVal(cty.Number),
			}))
			d.EXPECT().GetOk("password").Return("NewPassword123!", true)
			d.EXPECT().HasChange("password").Return(true)

			password, shouldReset := shouldResetPassword(d)

			assert.Equal(t, "NewPassword123!", password)
			assert.True(t, shouldReset, "password change requires reset")
		})

		t.Run("write-only password version incremented", func(t *testing.T) {
			d := mocks.NewMockResourceData(t)
			d.EXPECT().Id().Return("proj/svc/user")

			// user rotated password by incrementing password_wo_version to 2
			d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.StringVal("Rotated456!"),
				"password_wo_version": cty.NumberIntVal(2),
			}))
			d.EXPECT().HasChange("password").Return(false)
			d.EXPECT().HasChange("password_wo_version").Return(true)

			password, shouldReset := shouldResetPassword(d)

			assert.Equal(t, "Rotated456!", password)
			assert.True(t, shouldReset, "password_wo_version change requires reset")
		})

		t.Run("switched from optional to write-only password", func(t *testing.T) {
			d := mocks.NewMockResourceData(t)
			d.EXPECT().Id().Return("proj/svc/user")

			// User migrated from password to password_wo
			d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.StringVal("NewWriteOnly789!"),
				"password_wo_version": cty.NumberIntVal(1),
			}))
			d.EXPECT().HasChange("password").Return(false)
			d.EXPECT().HasChange("password_wo_version").Return(true)

			password, shouldReset := shouldResetPassword(d)

			assert.Equal(t, "NewWriteOnly789!", password)
			assert.True(t, shouldReset, "switching password mode requires reset")
		})

		t.Run("switched from write-only to auto-generated password", func(t *testing.T) {
			d := mocks.NewMockResourceData(t)
			d.EXPECT().Id().Return("proj/svc/user")

			// user removed password_wo to let service auto-generate
			d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.NullVal(cty.String),
				"password_wo_version": cty.NullVal(cty.Number),
			}))
			d.EXPECT().GetOk("password").Return(nil, false)
			d.EXPECT().HasChange("password").Return(false)
			d.EXPECT().HasChange("password_wo_version").Return(true)

			password, shouldReset := shouldResetPassword(d)

			assert.Empty(t, password, "empty triggers auto-generation")
			assert.True(t, shouldReset, "removing explicit password requires reset")
		})
	})
}

func TestUpsertPassword(t *testing.T) {
	t.Parallel()

	t.Run("creating new resources", func(t *testing.T) {
		t.Parallel()

		t.Run("with auto-generated password calls reset API", func(t *testing.T) {
			d := mocks.NewMockResourceData(t)
			client := avngen.NewMockClient(t)

			// new resource with no password specified
			d.EXPECT().Id().Return("")
			d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.NullVal(cty.String),
				"password_wo_version": cty.NullVal(cty.Number),
			}))
			d.EXPECT().GetOk("password").Return(nil, false)

			// expect: calls reset API (which auto-generates password)
			d.EXPECT().Get("project").Return("test-project")
			d.EXPECT().Get("service_name").Return("test-service")
			d.EXPECT().Get("username").Return("test-user")
			client.EXPECT().ServiceUserCredentialsReset(
				context.Background(), "test-project", "test-service", "test-user",
			).Return(&service.ServiceUserCredentialsResetOut{}, nil)

			err := UpsertPassword(context.Background(), d, client)
			assert.NoError(t, err)
		})

		t.Run("with optional password field calls modify API", func(t *testing.T) {
			d := mocks.NewMockResourceData(t)
			client := avngen.NewMockClient(t)

			// new resource with password = "Custom123!"
			d.EXPECT().Id().Return("")
			d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.NullVal(cty.String),
				"password_wo_version": cty.NullVal(cty.Number),
			}))
			d.EXPECT().GetOk("password").Return("Custom123!", true)

			// expect: calls modify API with specific password
			d.EXPECT().Get("project").Return("test-project")
			d.EXPECT().Get("service_name").Return("test-service")
			d.EXPECT().Get("username").Return("test-user")
			client.EXPECT().ServiceUserCredentialsModify(
				context.Background(), "test-project", "test-service", "test-user",
				&service.ServiceUserCredentialsModifyIn{
					NewPassword: lo.ToPtr("Custom123!"),
					Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
				},
			).Return(&service.ServiceUserCredentialsModifyOut{}, nil)

			err := UpsertPassword(context.Background(), d, client)
			assert.NoError(t, err)
		})

		t.Run("with write-only password calls modify API", func(t *testing.T) {
			d := mocks.NewMockResourceData(t)
			client := avngen.NewMockClient(t)

			// new resource with password_wo = "WriteOnly456!"
			d.EXPECT().Id().Return("")
			d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.StringVal("WriteOnly456!"),
				"password_wo_version": cty.NumberIntVal(1),
			}))

			// calls modify API with write-only password
			d.EXPECT().Get("project").Return("test-project")
			d.EXPECT().Get("service_name").Return("test-service")
			d.EXPECT().Get("username").Return("test-user")
			client.EXPECT().ServiceUserCredentialsModify(
				context.Background(), "test-project", "test-service", "test-user",
				&service.ServiceUserCredentialsModifyIn{
					NewPassword: lo.ToPtr("WriteOnly456!"),
					Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
				},
			).Return(&service.ServiceUserCredentialsModifyOut{}, nil)

			err := UpsertPassword(context.Background(), d, client)
			assert.NoError(t, err)
		})
	})

	t.Run("existing resource with no password changes", func(t *testing.T) {
		t.Parallel()
		d := mocks.NewMockResourceData(t)
		client := avngen.NewMockClient(t)

		// existing resource, password unchanged
		d.EXPECT().Id().Return("proj/svc/user")
		d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
			"password_wo":         cty.NullVal(cty.String),
			"password_wo_version": cty.NullVal(cty.Number),
		}))
		d.EXPECT().GetOk("password").Return("Existing789!", true)
		d.EXPECT().HasChange("password").Return(false)
		d.EXPECT().HasChange("password_wo_version").Return(false)

		// expect: no API calls
		err := UpsertPassword(context.Background(), d, client)
		assert.NoError(t, err)
	})

	t.Run("updating existing resource passwords", func(t *testing.T) {
		t.Parallel()

		t.Run("when optional password field changes", func(t *testing.T) {
			d := mocks.NewMockResourceData(t)
			client := avngen.NewMockClient(t)

			// user changed password field
			d.EXPECT().Id().Return("proj/svc/user")
			d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.NullVal(cty.String),
				"password_wo_version": cty.NullVal(cty.Number),
			}))
			d.EXPECT().GetOk("password").Return("Updated123!", true)
			d.EXPECT().HasChange("password").Return(true)

			// expect: calls modify API with new password
			d.EXPECT().Get("project").Return("test-project")
			d.EXPECT().Get("service_name").Return("test-service")
			d.EXPECT().Get("username").Return("test-user")
			client.EXPECT().ServiceUserCredentialsModify(
				context.Background(), "test-project", "test-service", "test-user",
				&service.ServiceUserCredentialsModifyIn{
					NewPassword: lo.ToPtr("Updated123!"),
					Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
				},
			).Return(&service.ServiceUserCredentialsModifyOut{}, nil)

			err := UpsertPassword(context.Background(), d, client)
			assert.NoError(t, err)
		})

		t.Run("when write-only password version increments", func(t *testing.T) {
			d := mocks.NewMockResourceData(t)
			client := avngen.NewMockClient(t)

			// user rotated password_wo by incrementing version
			d.EXPECT().Id().Return("proj/svc/user")
			d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.StringVal("Rotated789!"),
				"password_wo_version": cty.NumberIntVal(2),
			}))
			d.EXPECT().HasChange("password").Return(false)
			d.EXPECT().HasChange("password_wo_version").Return(true)

			// calls modify API with rotated password
			d.EXPECT().Get("project").Return("test-project")
			d.EXPECT().Get("service_name").Return("test-service")
			d.EXPECT().Get("username").Return("test-user")
			client.EXPECT().ServiceUserCredentialsModify(
				context.Background(), "test-project", "test-service", "test-user",
				&service.ServiceUserCredentialsModifyIn{
					NewPassword: lo.ToPtr("Rotated789!"),
					Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
				},
			).Return(&service.ServiceUserCredentialsModifyOut{}, nil)

			err := UpsertPassword(context.Background(), d, client)
			assert.NoError(t, err)
		})

		t.Run("when switching from write-only to auto-generated", func(t *testing.T) {
			d := mocks.NewMockResourceData(t)
			client := avngen.NewMockClient(t)

			// user removed password_wo to auto-generate
			d.EXPECT().Id().Return("proj/svc/user")
			d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
				"password_wo":         cty.NullVal(cty.String),
				"password_wo_version": cty.NullVal(cty.Number),
			}))
			d.EXPECT().GetOk("password").Return(nil, false)
			d.EXPECT().HasChange("password").Return(false)
			d.EXPECT().HasChange("password_wo_version").Return(true)

			// expect: calls reset API (which auto-generates new password)
			d.EXPECT().Get("project").Return("test-project")
			d.EXPECT().Get("service_name").Return("test-service")
			d.EXPECT().Get("username").Return("test-user")
			client.EXPECT().ServiceUserCredentialsReset(
				context.Background(), "test-project", "test-service", "test-user",
			).Return(&service.ServiceUserCredentialsResetOut{}, nil)

			err := UpsertPassword(context.Background(), d, client)
			assert.NoError(t, err)
		})
	})
}
