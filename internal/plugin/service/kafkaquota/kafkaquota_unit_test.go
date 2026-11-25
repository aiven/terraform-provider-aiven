package kafkaquota

import (
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafka"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func newResourceData(t *testing.T, opts ...adapter.ResourceDataOpt) adapter.ResourceData {
	t.Helper()
	opts = append(opts, adapter.WithIdentityFieldsOptional(ResourceOptions.IdentityFieldsOptional))
	d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields, opts...)
	require.NoError(t, err)
	return d
}

func TestCreateView(t *testing.T) {
	cases := []struct {
		name   string
		plan   map[string]any
		req    *kafka.ServiceKafkaQuotaCreateIn
		query  [][2]string
		wantID string
	}{
		{
			name: "user",
			plan: map[string]any{
				"project":            "p",
				"service_name":       "s",
				"user":               "alice",
				"consumer_byte_rate": 1000,
				"request_percentage": 10.5,
			},
			req: &kafka.ServiceKafkaQuotaCreateIn{
				User:              lo.ToPtr("alice"),
				ConsumerByteRate:  lo.ToPtr(1000.0),
				RequestPercentage: lo.ToPtr(10.5),
			},
			query:  [][2]string{kafka.ServiceKafkaQuotaDescribeUser("alice")},
			wantID: "p/s//alice",
		},
		{
			name: "client_id",
			plan: map[string]any{
				"project":            "p",
				"service_name":       "s",
				"client_id":          "c1",
				"producer_byte_rate": 500,
			},
			req: &kafka.ServiceKafkaQuotaCreateIn{
				ClientId:         lo.ToPtr("c1"),
				ProducerByteRate: lo.ToPtr(500.0),
			},
			query:  [][2]string{kafka.ServiceKafkaQuotaDescribeClientId("c1")},
			wantID: "p/s/c1/",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			client := avngen.NewMockClient(t)
			client.EXPECT().ServiceKafkaQuotaCreate(ctx, "p", "s", tc.req).Return(nil).Once()
			client.EXPECT().
				ServiceKafkaQuotaDescribe(mock.Anything, "p", "s", tc.query).
				Return(
					&kafka.ServiceKafkaQuotaDescribeOut{
						ConsumerByteRate:  tc.req.ConsumerByteRate,
						ProducerByteRate:  tc.req.ProducerByteRate,
						RequestPercentage: tc.req.RequestPercentage,
					},
					nil,
				).
				Once()

			d := newResourceData(t, adapter.WithTestPlan(tc.plan))
			require.NoError(t, createView(ctx, client, d))
			assert.Equal(t, tc.wantID, d.ID())
		})
	}
}

func TestUpdateViewWaitsForRates(t *testing.T) {
	ctx := t.Context()
	client := avngen.NewMockClient(t)
	client.EXPECT().ServiceKafkaQuotaCreate(ctx, "p", "s", mock.Anything).Return(nil).Once()
	client.EXPECT().
		ServiceKafkaQuotaDescribe(mock.Anything, "p", "s", mock.Anything).
		Return(nil, avngen.Error{Status: 404}).
		Once()
	client.EXPECT().
		ServiceKafkaQuotaDescribe(mock.Anything, "p", "s", mock.Anything).
		Return(&kafka.ServiceKafkaQuotaDescribeOut{ConsumerByteRate: lo.ToPtr(1000.0)}, nil).
		Once()
	client.EXPECT().
		ServiceKafkaQuotaDescribe(mock.Anything, "p", "s", mock.Anything).
		Return(&kafka.ServiceKafkaQuotaDescribeOut{ConsumerByteRate: lo.ToPtr(2000.0)}, nil).
		Once()

	d := newResourceData(
		t,
		adapter.WithTestPlan(map[string]any{
			"project":            "p",
			"service_name":       "s",
			"user":               "alice",
			"consumer_byte_rate": 2000,
		}),
	)
	require.NoError(t, updateView(ctx, client, d))
}

func TestReadView(t *testing.T) {
	cases := []struct {
		name  string
		state map[string]any
		rsp   *kafka.ServiceKafkaQuotaDescribeOut
		want  map[string]any
	}{
		{
			name:  "import",
			state: map[string]any{"id": "p/s//alice"},
			rsp: &kafka.ServiceKafkaQuotaDescribeOut{
				ConsumerByteRate:  lo.ToPtr(1e9),
				ProducerByteRate:  lo.ToPtr(0.0),
				RequestPercentage: lo.ToPtr(12.5),
			},
			want: map[string]any{"consumer_byte_rate": 1000000000, "producer_byte_rate": 0, "request_percentage": 12.5},
		},
		{
			name: "SDK state with zero for unset rates",
			state: map[string]any{
				"id":                 "p/s//alice",
				"project":            "p",
				"service_name":       "s",
				"user":               "alice",
				"consumer_byte_rate": 0,
				"producer_byte_rate": 0,
				"request_percentage": 10.0,
			},
			rsp:  &kafka.ServiceKafkaQuotaDescribeOut{RequestPercentage: lo.ToPtr(10.0)},
			want: map[string]any{"consumer_byte_rate": nil, "producer_byte_rate": nil, "request_percentage": 10.0},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			client := avngen.NewMockClient(t)
			client.EXPECT().
				ServiceKafkaQuotaDescribe(ctx, "p", "s", [][2]string{kafka.ServiceKafkaQuotaDescribeUser("alice")}).
				Return(tc.rsp, nil).
				Once()

			d := newResourceData(t, adapter.WithTestState(tc.state))
			require.NoError(t, readView(ctx, client, d))
			for k, want := range tc.want {
				got, ok := d.GetOk(k)
				if want == nil {
					assert.False(t, ok, k)
				} else {
					assert.Equal(t, want, got, k)
				}
			}
		})
	}
}

func TestDeleteView(t *testing.T) {
	ctx := t.Context()
	client := avngen.NewMockClient(t)
	client.EXPECT().
		ServiceKafkaQuotaDelete(
			ctx,
			"p",
			"s",
			[][2]string{kafka.ServiceKafkaQuotaDeleteUser("alice"), kafka.ServiceKafkaQuotaDeleteClientId("c1")},
		).
		Return(nil).
		Once()

	d := newResourceData(t, adapter.WithTestState(map[string]any{"id": "p/s/c1/alice"}))
	require.NoError(t, deleteView(ctx, client, d))
}
