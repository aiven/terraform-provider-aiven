package aiven

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"reflect"
	"testing"
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
