package user

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestResetPassword(t *testing.T) {
	t.Parallel()

	const project, serviceName, username = "prj", "svc", "usr"
	ctx := context.Background()

	tests := []struct {
		name               string
		plan               map[string]any
		state              map[string]any
		config             map[string]any
		wantModifyCall     bool
		wantAuthentication string
		wantSetPassword    *string
	}{
		{
			name: "new resource without password and authentication does not call Modify",
			plan: map[string]any{
				"project": project, "service_name": serviceName, "username": username,
			},
			state:              nil,
			config:             nil,
			wantModifyCall:     false,
			wantAuthentication: "",
			wantSetPassword:    nil,
		},
		{
			name: "new resource with authentication but no password calls Modify with nil password",
			plan: map[string]any{
				"project": project, "service_name": serviceName, "username": username,
				"authentication": "caching_sha2_password",
			},
			state:              nil,
			config:             nil,
			wantModifyCall:     true,
			wantAuthentication: "caching_sha2_password",
			wantSetPassword:    nil,
		},
		{
			name: "new resource with password calls Modify",
			plan: map[string]any{
				"project": project, "service_name": serviceName, "username": username,
				"password": "Custom$Pass1",
			},
			state:              nil,
			config:             nil,
			wantModifyCall:     true,
			wantAuthentication: "",
			wantSetPassword:    lo.ToPtr("Custom$Pass1"),
		},
		{
			name: "new resource with password and authentication calls Modify",
			plan: map[string]any{
				"project": project, "service_name": serviceName, "username": username,
				"password": "Custom$Pass1", "authentication": "mysql_native_password",
			},
			state:              nil,
			config:             nil,
			wantModifyCall:     true,
			wantAuthentication: "mysql_native_password",
			wantSetPassword:    lo.ToPtr("Custom$Pass1"),
		},
		{
			name: "new resource with password_wo calls Modify",
			plan: map[string]any{
				"project": project, "service_name": serviceName, "username": username,
				"password_wo": "WriteOnlyPass$1", "password_wo_version": 1,
			},
			state:              nil,
			config:             nil,
			wantModifyCall:     true,
			wantAuthentication: "",
			wantSetPassword:    lo.ToPtr("WriteOnlyPass$1"),
		},
		{
			name: "existing resource no password or authentication change does not call Modify",
			plan: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password": "SamePass",
			},
			state: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password": "SamePass",
			},
			config:             nil,
			wantModifyCall:     false,
			wantAuthentication: "",
			wantSetPassword:    nil,
		},
		{
			name: "existing resource password changed calls Modify",
			plan: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password": "NewPass456",
			},
			state: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password": "OldPass",
			},
			config:             nil,
			wantModifyCall:     true,
			wantAuthentication: "",
			wantSetPassword:    lo.ToPtr("NewPass456"),
		},
		{
			name: "existing resource authentication changed calls Modify",
			plan: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password": "SamePass", "authentication": "mysql_native_password",
			},
			state: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password": "SamePass", "authentication": "caching_sha2_password",
			},
			config:             nil,
			wantModifyCall:     true,
			wantAuthentication: "mysql_native_password",
			wantSetPassword:    lo.ToPtr("SamePass"),
		},
		{
			name: "existing resource password_wo_version changed calls Modify",
			plan: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password_wo": "Rotated$2", "password_wo_version": 2,
			},
			state: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password_wo_version": 1,
			},
			config:             nil,
			wantModifyCall:     true,
			wantAuthentication: "",
			wantSetPassword:    lo.ToPtr("Rotated$2"),
		},
		{
			name: "existing resource password removed calls Modify with nil so backend generates",
			plan: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
			},
			state: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password": "OldPass",
			},
			config:             nil,
			wantModifyCall:     true,
			wantAuthentication: "",
			wantSetPassword:    nil,
		},
		{
			name: "existing resource switch from password to password_wo calls Modify",
			plan: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password_wo": "WriteOnlyPass$1", "password_wo_version": 1,
			},
			state: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password": "OldPass",
			},
			config:             nil,
			wantModifyCall:     true,
			wantAuthentication: "",
			wantSetPassword:    lo.ToPtr("WriteOnlyPass$1"),
		},
		{
			name: "existing resource switch from password_wo back to password calls Modify",
			plan: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password": "BackToCustom$99",
			},
			state: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password_wo_version": 1,
			},
			config:             nil,
			wantModifyCall:     true,
			wantAuthentication: "",
			wantSetPassword:    lo.ToPtr("BackToCustom$99"),
		},
		{
			name: "existing resource password_wo removed calls Modify with nil so backend generates",
			plan: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
			},
			state: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password_wo_version": 1,
			},
			config:             nil,
			wantModifyCall:     true,
			wantAuthentication: "",
			wantSetPassword:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			config := tt.config
			if config == nil {
				config = tt.plan
			}
			d, err := adapter.NewResourceDataFromMaps(resourceSchemaInternal(), idFields(), tt.plan, tt.state, config)
			require.NoError(t, err)

			if tt.wantModifyCall {
				client.EXPECT().
					ServiceUserCredentialsModify(ctx, project, serviceName, username, &service.ServiceUserCredentialsModifyIn{
						Operation:      service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
						Authentication: service.AuthenticationType(tt.wantAuthentication),
						NewPassword:    tt.wantSetPassword,
					}).
					Return(&service.ServiceUserCredentialsModifyOut{}, nil)
			}

			err = resetPassword(ctx, client, d)
			require.NoError(t, err)
		})
	}
}
