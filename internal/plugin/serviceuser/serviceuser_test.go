package serviceuser_test

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/serviceuser"
)

func schemaInternal() *adapter.Schema {
	return &adapter.Schema{
		Type: adapter.SchemaTypeObject,
		Properties: map[string]*adapter.Schema{
			"id":                  {Type: adapter.SchemaTypeString, Computed: true},
			"project":             {Type: adapter.SchemaTypeString},
			"service_name":        {Type: adapter.SchemaTypeString},
			"username":            {Type: adapter.SchemaTypeString},
			"password":            {Type: adapter.SchemaTypeString, Computed: true, ZeroNotAllowed: true},
			"password_wo":         {Type: adapter.SchemaTypeString, ZeroNotAllowed: true},
			"password_wo_version": {Type: adapter.SchemaTypeInt, ZeroNotAllowed: true},
		},
	}
}

func idFields() []string {
	return []string{"project", "service_name", "username"}
}

func TestResetPassword(t *testing.T) {
	t.Parallel()

	const project, serviceName, username = "prj", "svc", "usr"
	ctx := context.Background()

	tests := []struct {
		name            string
		plan            map[string]any
		state           map[string]any
		config          map[string]any
		wantModifyCall  bool
		wantSetPassword *string
	}{
		{
			name: "new resource without password does not call Modify",
			plan: map[string]any{
				"project": project, "service_name": serviceName, "username": username,
			},
			state:           nil,
			config:          nil,
			wantModifyCall:  false,
			wantSetPassword: nil,
		},
		{
			name: "new resource with password calls Modify",
			plan: map[string]any{
				"project": project, "service_name": serviceName, "username": username,
				"password": "Custom$Pass1",
			},
			state:           nil,
			config:          nil,
			wantModifyCall:  true,
			wantSetPassword: new("Custom$Pass1"),
		},
		{
			name: "new resource with password_wo calls Modify",
			plan: map[string]any{
				"project": project, "service_name": serviceName, "username": username,
				"password_wo": "WriteOnlyPass$1", "password_wo_version": 1,
			},
			state:           nil,
			config:          nil,
			wantModifyCall:  true,
			wantSetPassword: new("WriteOnlyPass$1"),
		},
		{
			name: "existing resource no password change does not call Modify",
			plan: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password": "SamePass",
			},
			state: map[string]any{
				"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
				"password": "SamePass",
			},
			config:          nil,
			wantModifyCall:  false,
			wantSetPassword: nil,
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
			config:          nil,
			wantModifyCall:  true,
			wantSetPassword: new("NewPass456"),
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
			config:          nil,
			wantModifyCall:  true,
			wantSetPassword: new("Rotated$2"),
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
			config:          nil,
			wantModifyCall:  true,
			wantSetPassword: nil, // nil NewPassword: backend generates password
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
			config:          nil,
			wantModifyCall:  true,
			wantSetPassword: new("WriteOnlyPass$1"),
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
			config:          nil,
			wantModifyCall:  true,
			wantSetPassword: new("BackToCustom$99"),
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
			config:          nil,
			wantModifyCall:  true,
			wantSetPassword: nil, // nil NewPassword: backend generates password
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := avngen.NewMockClient(t)
			config := tt.config
			if config == nil {
				config = tt.plan
			}
			d, err := adapter.NewResourceData(schemaInternal(), idFields(),
				adapter.WithTestPlan(tt.plan),
				adapter.WithTestState(tt.state),
				adapter.WithTestConfig(config),
			)
			require.NoError(t, err)

			if tt.wantModifyCall {
				client.EXPECT().
					ServiceUserCredentialsModify(ctx, project, serviceName, username, &service.ServiceUserCredentialsModifyIn{
						Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
						NewPassword: tt.wantSetPassword,
					}).
					Return(&service.ServiceUserCredentialsModifyOut{}, nil)
			}

			err = serviceuser.ResetPassword(ctx, client, d)
			require.NoError(t, err)
		})
	}
}
