package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"os"
	"reflect"
	"testing"
)

var (
	testAccProviders         map[string]*schema.Provider
	testAccProvider          *schema.Provider
	testAccProviderFactories map[string]func() (*schema.Provider, error)
)

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"aiven": testAccProvider,
	}
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"aiven": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderImpl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("AIVEN_TOKEN"); v == "" {
		t.Log(v)
		t.Fatal("AIVEN_TOKEN must be set for acceptance tests")
	}

	// Provider a project name with enough credits to run acceptance
	// tests or project name with the assigned payment card.
	if v := os.Getenv("AIVEN_PROJECT_NAME"); v == "" {
		log.Print("[WARNING] AIVEN_PROJECT_NAME must be set for some acceptance tests")
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
