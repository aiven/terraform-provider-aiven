package serviceuser_test

import (
	"context"
	"maps"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/stretchr/testify/mock"
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
		wantErr         error
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
			wantErr:         adapter.ErrNoChanges,
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
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	const project, serviceName, username = "prj", "svc", "usr"
	ctx := context.Background()
	schema := func() *adapter.Schema {
		s := schemaInternal()
		s.Properties["authentication"] = &adapter.Schema{Type: adapter.SchemaTypeString}
		s.Properties["pg_allow_replication"] = &adapter.Schema{Type: adapter.SchemaTypeBool, HoistedFrom: serviceuser.AccessControlField}
		return s
	}
	state := func() map[string]any {
		return map[string]any{
			"id": "prj/svc/usr", "project": project, "service_name": serviceName, "username": username,
			"password": "OldPass", "authentication": "caching_sha2_password", "pg_allow_replication": false,
		}
	}
	setAccessControl := &service.ServiceUserCredentialsModifyIn{
		Operation:     service.ServiceUserCredentialsModifyOperationTypeSetAccessControl,
		AccessControl: &service.AccessControlIn{PGAllowReplication: new(true)},
	}
	resetCredentials := &service.ServiceUserCredentialsModifyIn{
		Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
		NewPassword: new("NewPass456"),
	}

	tests := []struct {
		name    string
		plan    map[string]any
		want    []*service.ServiceUserCredentialsModifyIn
		wantErr error
	}{
		{
			name:    "no changes",
			wantErr: adapter.ErrNoChanges,
		},
		{
			name: "access control change sends only the access control",
			plan: map[string]any{"pg_allow_replication": true},
			want: []*service.ServiceUserCredentialsModifyIn{setAccessControl},
		},
		{
			name: "access control and password change",
			plan: map[string]any{"pg_allow_replication": true, "password": "NewPass456"},
			want: []*service.ServiceUserCredentialsModifyIn{setAccessControl, resetCredentials},
		},
		{
			name: "password change",
			plan: map[string]any{"password": "NewPass456"},
			want: []*service.ServiceUserCredentialsModifyIn{resetCredentials},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plan := state()
			maps.Copy(plan, tt.plan)
			d, err := adapter.NewResourceData(schema(), idFields(),
				adapter.WithTestPlan(plan),
				adapter.WithTestState(state()),
				adapter.WithTestConfig(plan),
			)
			require.NoError(t, err)

			client := avngen.NewMockClient(t)
			var prev *mock.Call
			for _, req := range tt.want {
				call := client.EXPECT().
					ServiceUserCredentialsModify(ctx, project, serviceName, username, req).
					Return(&service.ServiceUserCredentialsModifyOut{}, nil).
					Once()
				if prev != nil {
					call.NotBefore(prev)
				}
				prev = call
			}

			require.ErrorIs(t, serviceuser.Update(ctx, client, d), tt.wantErr)
		})
	}
}

func TestPasswordIsReady(t *testing.T) {
	t.Parallel()

	// schemaWithoutWriteOnly mirrors schemaInternal but omits the password_wo_version property.
	schemaWithoutWriteOnly := func() *adapter.Schema {
		s := schemaInternal()
		delete(s.Properties, "password_wo_version")
		delete(s.Properties, "password_wo")
		return s
	}

	tests := []struct {
		name    string
		schema  *adapter.Schema
		state   map[string]any
		wantErr string
	}{
		{
			name:  "password present in state",
			state: map[string]any{"password": "SomePass$1"},
		},
		{
			name:    "password empty while backend regenerates",
			state:   map[string]any{},
			wantErr: "password is not available yet",
		},
		{
			name:  "write-only mode keeps password out of state",
			state: map[string]any{"password_wo_version": 1},
		},
		{
			name:   "schema without password_wo_version still checks password",
			schema: schemaWithoutWriteOnly(),
			state:  map[string]any{"password": "SomePass$1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := tt.schema
			if s == nil {
				s = schemaInternal()
			}

			d, err := adapter.NewResourceData(s, idFields(),
				adapter.WithTestState(tt.state),
			)
			require.NoError(t, err)

			err = serviceuser.PasswordIsReady(d)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
