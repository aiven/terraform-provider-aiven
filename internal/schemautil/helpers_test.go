package schemautil

import (
	"context"
	"testing"

	"github.com/aiven/go-client-codegen/handler/organization"
	"github.com/stretchr/testify/assert"
)

type mockAvngenClient struct {
	get func(ctx context.Context, id string) (*organization.OrganizationGetOut, error)
}

func (c *mockAvngenClient) OrganizationGet(ctx context.Context, id string) (*organization.OrganizationGetOut, error) {
	return c.get(ctx, id)
}

func TestDetermineMixedOrganizationConstraintIDToStore(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()
	)

	type testCase struct {
		name    string
		input   string
		want    string
		wantErr bool
	}

	tests := []testCase{
		{
			name:    "provided Organization ID",
			input:   "org-123",
			want:    "acc-123",
			wantErr: false,
		},
		{
			name:    "provided Account ID",
			input:   "acc-123",
			want:    "acc-123",
			wantErr: false,
		},
		{
			name:    "provided an empty ID",
			input:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "error when fetching organization",
			input:   "org-123",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := new(mockAvngenClient)
			client.get = func(_ context.Context, _ string) (*organization.OrganizationGetOut, error) {
				if tt.wantErr {
					return nil, assert.AnError
				}

				return &organization.OrganizationGetOut{AccountId: tt.want}, nil
			}

			got, err := ConvertOrganizationToAccountID(ctx, tt.input, client)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
			assert.Equalf(t, tt.want, got, "expected %s, got %s", tt.want, got)
		})
	}

}
