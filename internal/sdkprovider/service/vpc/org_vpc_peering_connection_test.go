package vpc

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
	"unsafe"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationvpc"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/aiven/terraform-provider-aiven/mocks"
)

var testMu = &sync.RWMutex{}

func TestCreatePeeringConnection(t *testing.T) {
	t.Parallel()

	var (
		ctx        = context.Background()
		mockClient = mocks.NewMockClient(t)
		d          = schema.TestResourceDataRaw(t, nil, nil)

		orgID = uuid.New().String()
		vpcID = uuid.New().String()
	)

	testCases := []struct {
		name          string
		setupMocks    func() *mocks.MockClient
		expectedState organizationvpc.VpcPeeringConnectionStateType
		expectError   bool
	}{
		{
			name:          "successful creation and approval",
			expectedState: organizationvpc.VpcPeeringConnectionStateTypeActive,
			setupMocks: func() *mocks.MockClient {
				pcID := uuid.New().String()
				mc := mocks.NewMockClient(t)
				setTimeouts(t, d, 10*time.Millisecond, 10*time.Millisecond, 1*time.Second)

				// Setup create response
				mc.On("OrganizationVpcPeeringConnectionCreate",
					ctx,
					orgID,
					vpcID,
					mock.AnythingOfType("*organizationvpc.OrganizationVpcPeeringConnectionCreateIn"),
				).Return(&organizationvpc.OrganizationVpcPeeringConnectionCreateOut{
					PeeringConnectionId: &pcID,
				}, nil).Once()

				// Setup first get response (pending)
				mc.On("OrganizationVpcGet",
					ctx,
					orgID,
					vpcID,
				).Return(&organizationvpc.OrganizationVpcGetOut{
					PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{
						{
							PeeringConnectionId: &pcID,
							State:               organizationvpc.VpcPeeringConnectionStateTypeApproved,
						},
					},
				}, nil).Once()

				// Setup second get response (approved)
				mc.On("OrganizationVpcGet",
					ctx,
					orgID,
					vpcID,
				).Return(&organizationvpc.OrganizationVpcGetOut{
					PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{
						{
							PeeringConnectionId: &pcID,
							State:               organizationvpc.VpcPeeringConnectionStateTypeActive,
						},
					},
				}, nil).Once()

				return mc
			},
		},
		{
			name: "creation fails",
			setupMocks: func() *mocks.MockClient {
				mc := mocks.NewMockClient(t)

				mc.On("OrganizationVpcPeeringConnectionCreate",
					ctx,
					orgID,
					vpcID,
					mock.AnythingOfType("*organizationvpc.OrganizationVpcPeeringConnectionCreateIn"),
				).Return(nil, errors.New("creation failed")).Once()

				return mc
			},
			expectError: true,
		},
		{
			name: "approval timeout",
			setupMocks: func() *mocks.MockClient {
				pcID := uuid.New().String()
				mc := mocks.NewMockClient(t)
				setTimeouts(t, d, 10*time.Millisecond, 10*time.Millisecond, 100*time.Millisecond)

				mc.On("OrganizationVpcPeeringConnectionCreate",
					ctx,
					orgID,
					vpcID,
					mock.AnythingOfType("*organizationvpc.OrganizationVpcPeeringConnectionCreateIn"),
				).Return(&organizationvpc.OrganizationVpcPeeringConnectionCreateOut{
					PeeringConnectionId: &pcID,
				}, nil)

				// Always return pending state
				mc.On("OrganizationVpcGet",
					ctx,
					orgID,
					vpcID,
				).Return(&organizationvpc.OrganizationVpcGetOut{
					PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{
						{
							PeeringConnectionId: &pcID,
							State:               organizationvpc.VpcPeeringConnectionStateTypeApproved,
						},
					},
				}, nil)

				return mc
			},
			expectError: true,
		},
		{
			name: "peering connection disappears",
			setupMocks: func() *mocks.MockClient {
				pcID := uuid.New().String()
				mc := mocks.NewMockClient(t)
				setTimeouts(t, d, 10*time.Millisecond, 10*time.Millisecond, 1*time.Second)

				mc.On("OrganizationVpcPeeringConnectionCreate",
					ctx,
					orgID,
					vpcID,
					mock.AnythingOfType("*organizationvpc.OrganizationVpcPeeringConnectionCreateIn"),
				).Return(&organizationvpc.OrganizationVpcPeeringConnectionCreateOut{
					PeeringConnectionId: &pcID,
				}, nil).Once()

				mc.On("OrganizationVpcGet",
					ctx,
					orgID,
					vpcID,
				).Return(&organizationvpc.OrganizationVpcGetOut{
					PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{},
				}, nil).Once()

				return mc
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mc := tc.setupMocks()

			result, err := createPeeringConnection(
				ctx,
				orgID,
				vpcID,
				mc,
				d,
				organizationvpc.OrganizationVpcPeeringConnectionCreateIn{},
			)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedState, result.State)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func setTimeouts(t *testing.T, rd *schema.ResourceData, delay, interval, timeout time.Duration) {
	t.Helper()

	testMu.Lock()
	originalDelay := pollDelay
	originalInterval := pollInterval
	pollDelay = delay
	pollInterval = interval
	testMu.Unlock()

	t.Cleanup(func() {
		testMu.Lock()
		pollDelay = originalDelay
		pollInterval = originalInterval
		testMu.Unlock()
	})

	timeouts := &schema.ResourceTimeout{
		Create: lo.ToPtr(timeout),
		Read:   lo.ToPtr(timeout),
		Update: lo.ToPtr(timeout),
		Delete: lo.ToPtr(timeout),
	}

	val := reflect.ValueOf(rd).Elem()
	field := val.FieldByName("timeouts")
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(timeouts))
}

func TestDeletePeeringConnection(t *testing.T) {
	t.Parallel()

	var (
		ctx   = context.Background()
		d     = schema.TestResourceDataRaw(t, nil, nil)
		orgID = uuid.New().String()
		vpcID = uuid.New().String()
		pcID  = uuid.New().String()
		pc    = &organizationvpc.OrganizationVpcGetPeeringConnectionOut{
			PeeringConnectionId: &pcID,
			State:               organizationvpc.VpcPeeringConnectionStateTypeActive,
		}
	)

	testCases := []struct {
		name        string
		setupMocks  func() *mocks.MockClient
		inputPC     *organizationvpc.OrganizationVpcGetPeeringConnectionOut
		expectError bool
	}{
		{
			name:    "nil peering connection",
			inputPC: nil,
			setupMocks: func() *mocks.MockClient {
				return mocks.NewMockClient(t)
			},
			expectError: false,
		},
		{
			name:    "successful deletion",
			inputPC: pc,
			setupMocks: func() *mocks.MockClient {
				mc := mocks.NewMockClient(t)
				setTimeouts(t, d, 10*time.Millisecond, 10*time.Millisecond, 2*time.Second)

				// Setup delete response
				mc.On("OrganizationVpcPeeringConnectionDeleteById",
					ctx,
					orgID,
					vpcID,
					pcID,
				).Return(&organizationvpc.OrganizationVpcPeeringConnectionDeleteByIdOut{}, nil).Once()

				// First get after delete shows connection still exists
				mc.On("OrganizationVpcGet",
					ctx,
					orgID,
					vpcID,
				).Return(&organizationvpc.OrganizationVpcGetOut{
					PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{
						{
							PeeringConnectionId: &pcID,
							State:               organizationvpc.VpcPeeringConnectionStateTypeDeleting,
						},
					},
				}, nil).Once()

				// Second get shows connection is gone
				mc.On("OrganizationVpcGet",
					ctx,
					orgID,
					vpcID,
				).Return(&organizationvpc.OrganizationVpcGetOut{
					PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{
						{
							PeeringConnectionId: &pcID,
							State:               organizationvpc.VpcPeeringConnectionStateTypeDeleted,
						},
					},
				}, nil).Once()

				return mc
			},
			expectError: false,
		},
		{
			name:    "delete returns not found",
			inputPC: pc,
			setupMocks: func() *mocks.MockClient {
				mc := mocks.NewMockClient(t)
				setTimeouts(t, d, 10*time.Millisecond, 10*time.Millisecond, 2*time.Second)

				// Setup delete response with not found error
				mc.On("OrganizationVpcPeeringConnectionDeleteById",
					ctx,
					orgID,
					vpcID,
					pcID,
				).Return(nil, avngen.Error{Status: 404}).Once()

				return mc
			},
			expectError: false,
		},
		{
			name:    "delete fails with error",
			inputPC: pc,
			setupMocks: func() *mocks.MockClient {
				mc := mocks.NewMockClient(t)
				setTimeouts(t, d, 10*time.Millisecond, 10*time.Millisecond, 2*time.Second)

				mc.On("OrganizationVpcPeeringConnectionDeleteById",
					ctx,
					orgID,
					vpcID,
					pcID,
				).Return(nil, errors.New("delete failed")).Once()

				return mc
			},
			expectError: true,
		},
		{
			name:    "get after delete fails with non-404 error",
			inputPC: pc,
			setupMocks: func() *mocks.MockClient {
				mc := mocks.NewMockClient(t)
				setTimeouts(t, d, 10*time.Millisecond, 10*time.Millisecond, 2*time.Second)

				// Setup delete response
				mc.On("OrganizationVpcPeeringConnectionDeleteById",
					ctx,
					orgID,
					vpcID,
					pcID,
				).Return(&organizationvpc.OrganizationVpcPeeringConnectionDeleteByIdOut{}, nil).Once()

				// Get fails with error
				mc.On("OrganizationVpcGet",
					ctx,
					orgID,
					vpcID,
				).Return(nil, errors.New("get failed")).Once()

				return mc
			},
			expectError: true,
		},
		{
			name:    "get after delete returns 404",
			inputPC: pc,
			setupMocks: func() *mocks.MockClient {
				mc := mocks.NewMockClient(t)
				setTimeouts(t, d, 10*time.Millisecond, 10*time.Millisecond, 2*time.Second)

				// Setup delete response
				mc.On("OrganizationVpcPeeringConnectionDeleteById",
					ctx,
					orgID,
					vpcID,
					pcID,
				).Return(&organizationvpc.OrganizationVpcPeeringConnectionDeleteByIdOut{}, nil).Once()

				// Get returns 404
				mc.On("OrganizationVpcGet",
					ctx,
					orgID,
					vpcID,
				).Return(nil, avngen.Error{Status: 404}).Once()

				return mc
			},
			expectError: false,
		},
		{
			name:    "deletion timeout",
			inputPC: pc,
			setupMocks: func() *mocks.MockClient {
				mc := mocks.NewMockClient(t)
				setTimeouts(t, d, 10*time.Millisecond, 10*time.Millisecond, 1*time.Second)

				// Setup delete response
				mc.On("OrganizationVpcPeeringConnectionDeleteById",
					ctx,
					orgID,
					vpcID,
					pcID,
				).Return(&organizationvpc.OrganizationVpcPeeringConnectionDeleteByIdOut{}, nil).Once()

				// Always return the connection in deleting state
				mc.On("OrganizationVpcGet",
					ctx,
					orgID,
					vpcID,
				).Return(&organizationvpc.OrganizationVpcGetOut{
					PeeringConnections: []organizationvpc.OrganizationVpcGetPeeringConnectionOut{
						{
							PeeringConnectionId: &pcID,
							State:               organizationvpc.VpcPeeringConnectionStateTypeDeleting,
						},
					},
				}, nil)

				return mc
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mc := tc.setupMocks()

			err := deletePeeringConnection(
				ctx,
				orgID,
				vpcID,
				mc,
				d,
				tc.inputPC,
			)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mc.AssertExpectations(t)
		})
	}
}
