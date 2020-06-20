package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
	"time"
)

var (
	testAccProviders map[string]terraform.ResourceProvider
	testAccProvider  *schema.Provider
)

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"aiven": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderImpl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("AIVEN_TOKEN"); v == "" {
		t.Log(v)
		t.Fatal("AIVEN_TOKEN must be set for acceptance tests")
	}

	if v := os.Getenv("AIVEN_CARD_ID"); v == "" {
		t.Fatal("AIVEN_CARD_ID must be set for acceptance tests")
	}
}

func Test_validateDurationString(t *testing.T) {
	type args struct {
		v interface{}
		k string
	}
	tests := []struct {
		name       string
		args       args
		wantWs     []string
		wantErrors bool
	}{
		{
			"basic",
			args{
				v: "2m",
				k: "",
			},
			nil,
			false,
		},
		{
			"wrong-duration",
			args{
				v: "123qweert",
				k: "",
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWs, gotErrors := validateDurationString(tt.args.v, tt.args.k)
			if !reflect.DeepEqual(gotWs, tt.wantWs) {
				t.Errorf("validateDurationString() gotWs = %v, want %v", gotWs, tt.wantWs)
			}
			if !(tt.wantErrors == (len(gotErrors) > 0)) {
				t.Errorf("validateDurationString() gotErrors = %v", gotErrors)
			}
		})
	}
}

func Test_generateClientTimeoutsSchema(t *testing.T) {
	type args struct {
		timeouts map[string]time.Duration
	}
	tests := []struct {
		name string
		args args
		want *schema.Schema
	}{
		{
			"basic",
			args{map[string]time.Duration{"create": 1 * time.Minute}},
			&schema.Schema{
				Type:        schema.TypeSet,
				MaxItems:    1,
				Description: "Custom Terraform Client timeouts",
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"create": {
							Type:         schema.TypeString,
							Description:  "create timeout",
							Optional:     true,
							ValidateFunc: validateDurationString,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateClientTimeoutsSchema(tt.args.timeouts)
			assert.Equal(t, tt.want.Type, got.Type)
			assert.Equal(t, tt.want.MaxItems, got.MaxItems)
			assert.Equal(t, tt.want.Description, got.Description)
			assert.Equal(t, tt.want.Optional, got.Optional)

			for name, s := range got.Elem.(*schema.Resource).Schema {
				want := tt.want.Elem.(*schema.Resource).Schema[name]

				assert.Equal(t, want.Type, s.Type)
				assert.Equal(t, want.Description, s.Description)
				assert.Equal(t, want.Optional, s.Optional)
			}
		})
	}
}

func Test_getTimeoutHelper(t *testing.T) {
	type args struct {
		d               *schema.ResourceData
		name            string
		defaultDuration time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    time.Duration
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				d: resourceService().Data(&terraform.InstanceState{
					ID:         "",
					Attributes: nil,
					Ephemeral:  terraform.EphemeralState{},
					Meta:       nil,
					Tainted:    false,
				}),
				name:            "create",
				defaultDuration: 1 * time.Minute,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTimeoutHelper(tt.args.d, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("getTimeoutHelper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getTimeoutHelper() got = %v, want %v", got, tt.want)
			}
		})
	}
}
