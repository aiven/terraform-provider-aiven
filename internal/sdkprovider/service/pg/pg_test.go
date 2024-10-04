package pg_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const pgResourceTemplate = `
data "aiven_project" "foo" {
  project = "{{.Project}}"
}

{{- if .StaticIPCount }}
resource "aiven_static_ip" "ips" {
  count      = {{.StaticIPCount}}
  project    = data.aiven_project.foo.project
  cloud_name = "{{.CloudName}}"
}
{{- end }}

resource "aiven_pg" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "{{.CloudName}}"
  plan                    = "{{.Plan}}"
  service_name            = "{{.ServiceName}}"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  {{- if .DiskSpace }}
  disk_space              = "{{.DiskSpace}}"
  {{- end }}
  {{- if .AdditionalDiskSpace }}
  additional_disk_space   = "{{.AdditionalDiskSpace}}"
  {{- end }}
  {{- if .StaticIPCount }}
  static_ips              = toset(aiven_static_ip.ips[*].static_ip_address_id)
  {{- end }}
  {{- if .TerminationProtection }}
  termination_protection  = {{.TerminationProtection}}
  {{- end }}
  {{- if or .AdminUsername .StaticIPCount .PublicAccess .PGConfig }}
  pg_user_config {
    {{- if .AdminUsername }}
    admin_username = "{{.AdminUsername}}"
    admin_password = "{{.AdminPassword}}"
    {{- end }}
    {{- if .StaticIPCount }}
    static_ips = true
    {{- end }}
    {{- if .PublicAccess }}
    public_access {
      pg         = {{.PublicAccess.PG}}
      prometheus = {{.PublicAccess.Prometheus}}
    }
    {{- end }}
    {{- if .PGConfig }}
    pg {
      idle_in_transaction_session_timeout = {{.PGConfig.IdleInTransactionSessionTimeout}}
      log_min_duration_statement          = {{.PGConfig.LogMinDurationStatement}}
    }
    {{- end }}
  }
  {{- end }}
  {{- if .Timeouts }}
  timeouts {
    create = "{{.Timeouts.Create}}"
    update = "{{.Timeouts.Update}}"
  }
  {{- end }}
  tag {
    key   = "test"
    value = "val"
  }
}

{{- if .ReadReplica }}
resource "aiven_pg" "bar-replica" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "{{.CloudName}}"
  plan                    = "{{.Plan}}"
  service_name            = "{{.ReadReplica.ServiceName}}"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  pg_user_config {
    backup_hour   = 19
    backup_minute = 30
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }

  service_integrations {
    integration_type    = "read_replica"
    source_service_name = aiven_pg.bar.service_name
  }

  depends_on = [aiven_pg.bar]
}

resource "aiven_service_integration" "pg-readreplica" {
  project                  = data.aiven_project.foo.project
  integration_type         = "read_replica"
  source_service_name      = aiven_pg.bar.service_name
  destination_service_name = aiven_pg.bar-replica.service_name

  depends_on = [aiven_pg.bar-replica]
}
{{- end }}

data "aiven_pg" "common" {
  service_name = aiven_pg.bar.service_name
  project      = aiven_pg.bar.project

  depends_on = [aiven_pg.bar]
}
`

type PGResourceConfig struct {
	Project               string
	CloudName             string
	Plan                  string
	ServiceName           string
	DiskSpace             string
	AdditionalDiskSpace   string
	StaticIPCount         int
	TerminationProtection bool
	AdminUsername         string
	AdminPassword         string
	PublicAccess          *PublicAccessConfig
	PGConfig              *PGConfig
	Timeouts              *TimeoutsConfig
	ReadReplica           *ReadReplicaConfig
}

type PublicAccessConfig struct {
	PG         bool
	Prometheus bool
}

type PGConfig struct {
	IdleInTransactionSessionTimeout int
	LogMinDurationStatement         int
}

type TimeoutsConfig struct {
	Create string
	Update string
}

type ReadReplicaConfig struct {
	ServiceName string
}

func createBaseConfig(rName string) PGResourceConfig {
	return PGResourceConfig{
		Project:     os.Getenv("AIVEN_PROJECT_NAME"),
		CloudName:   "google-europe-west1",
		ServiceName: fmt.Sprintf("test-acc-sr-%s", rName),
		PublicAccess: &PublicAccessConfig{
			PG:         true,
			Prometheus: false,
		},
		PGConfig: &PGConfig{
			IdleInTransactionSessionTimeout: 900,
			LogMinDurationStatement:         -1,
		},
	}
}

func generateConfigWithChanges(baseConfig PGResourceConfig, changes map[string]interface{}) string {
	for key, value := range changes {
		switch key {
		case "Plan":
			baseConfig.Plan = value.(string)
		case "AdditionalDiskSpace":
			baseConfig.AdditionalDiskSpace = value.(string)
		case "DiskSpace":
			baseConfig.DiskSpace = value.(string)
		case "StaticIPCount":
			baseConfig.StaticIPCount = value.(int)
		case "AdminUsername":
			baseConfig.AdminUsername = value.(string)
		case "AdminPassword":
			baseConfig.AdminPassword = value.(string)
		case "TerminationProtection":
			baseConfig.TerminationProtection = value.(bool)
		case "PublicAccessPG":
			if baseConfig.PublicAccess == nil {
				baseConfig.PublicAccess = &PublicAccessConfig{}
			}
			baseConfig.PublicAccess.PG = value.(bool)
		case "PublicAccessPrometheus":
			if baseConfig.PublicAccess == nil {
				baseConfig.PublicAccess = &PublicAccessConfig{}
			}
			baseConfig.PublicAccess.Prometheus = value.(bool)
		case "IdleInTransactionSessionTimeout":
			if baseConfig.PGConfig == nil {
				baseConfig.PGConfig = &PGConfig{}
			}
			baseConfig.PGConfig.IdleInTransactionSessionTimeout = value.(int)
		case "LogMinDurationStatement":
			if baseConfig.PGConfig == nil {
				baseConfig.PGConfig = &PGConfig{}
			}
			baseConfig.PGConfig.LogMinDurationStatement = value.(int)
		case "TimeoutCreate":
			if baseConfig.Timeouts == nil {
				baseConfig.Timeouts = &TimeoutsConfig{}
			}
			baseConfig.Timeouts.Create = value.(string)
		case "TimeoutUpdate":
			if baseConfig.Timeouts == nil {
				baseConfig.Timeouts = &TimeoutsConfig{}
			}
			baseConfig.Timeouts.Update = value.(string)
		case "ReadReplicaServiceName":
			if baseConfig.ReadReplica == nil {
				baseConfig.ReadReplica = &ReadReplicaConfig{}
			}
			baseConfig.ReadReplica.ServiceName = value.(string)
			// Add more cases as needed
		}
	}

	configStr, err := renderConfig(baseConfig)
	if err != nil {
		panic(fmt.Sprintf("Error generating config: %s", err))
	}
	return configStr
}

func renderConfig(params PGResourceConfig) (string, error) {
	tmpl, err := template.New("pgResource").Parse(pgResourceTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func TestAccAivenPG_no_existing_project(t *testing.T) {
	config, err := renderConfig(PGResourceConfig{
		Project:     "wrong-project-name",
		CloudName:   "google-europe-west1",
		Plan:        "startup-4",
		ServiceName: "test-acc-sr-1",
		DiskSpace:   "100GiB",
	})
	if err != nil {
		t.Fatalf("Error generating config: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAivenPG_invalid_disk_size(t *testing.T) {
	expectErrorRegexBadString := regexp.MustCompile(regexp.QuoteMeta("configured string must match ^[1-9][0-9]*(G|GiB)"))
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	badConfigs := []struct {
		diskSize    string
		expectError *regexp.Regexp
	}{
		{"abc", expectErrorRegexBadString},
		{"01MiB", expectErrorRegexBadString},
		{"1234", expectErrorRegexBadString},
		{"5TiB", expectErrorRegexBadString},
		{" 1Gib ", expectErrorRegexBadString},
		{"1 GiB", expectErrorRegexBadString},
		{"1GiB", regexp.MustCompile("requested disk size is too small")},
		{"100000GiB", regexp.MustCompile("requested disk size is too large")},
		{"127GiB", regexp.MustCompile("requested disk size has to increase from: '.*' in increments of '.*'")},
	}

	additionalDiskConfigs := []struct {
		diskSize    string
		expectError *regexp.Regexp
	}{
		{"127GiB", regexp.MustCompile("requested disk size has to increase from: '.*' in increments of '.*'")},
		{"100000GiB", regexp.MustCompile("requested disk size is too large")},
		{"abc", expectErrorRegexBadString},
		{"01MiB", expectErrorRegexBadString},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: func() []resource.TestStep {
			var steps []resource.TestStep

			for _, cfg := range badConfigs {
				config, err := renderConfig(PGResourceConfig{
					Project:     os.Getenv("AIVEN_PROJECT_NAME"),
					CloudName:   "google-europe-west1",
					Plan:        "startup-4",
					ServiceName: fmt.Sprintf("test-acc-sr-%s", rName),
					DiskSpace:   cfg.diskSize,
				})
				if err != nil {
					t.Fatalf("Error generating config: %s", err)
				}

				steps = append(steps, resource.TestStep{
					Config:      config,
					PlanOnly:    true,
					ExpectError: cfg.expectError,
				})
			}

			for _, cfg := range additionalDiskConfigs {
				config, err := renderConfig(PGResourceConfig{
					Project:             os.Getenv("AIVEN_PROJECT_NAME"),
					CloudName:           "google-europe-west1",
					Plan:                "startup-4",
					ServiceName:         fmt.Sprintf("test-acc-sr-%s", rName),
					AdditionalDiskSpace: cfg.diskSize,
				})
				if err != nil {
					t.Fatalf("Error generating config: %s", err)
				}

				steps = append(steps, resource.TestStep{
					Config:      config,
					PlanOnly:    true,
					ExpectError: cfg.expectError,
				})
			}

			return steps
		}(),
	})
}

func TestAccAivenPG_static_ips(t *testing.T) {
	resourceName := "aiven_pg.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	baseConfig := createBaseConfig(rName)

	staticIPCounts := []int{2, 3, 4, 3, 4, 2}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: func() []resource.TestStep {
			var steps []resource.TestStep

			for _, count := range staticIPCounts {
				configStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
					"StaticIPCount": count,
				})

				steps = append(steps, resource.TestStep{
					Config: configStr,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
						resource.TestCheckResourceAttr(resourceName, "static_ips.#", fmt.Sprintf("%d", count)),
					),
				})
			}

			return steps
		}(),
	})
}

func TestAccAivenPG_changing_plan(t *testing.T) {
	resourceName := "aiven_pg.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	baseConfig := createBaseConfig(rName)

	initialConfigStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
		"Plan": "business-8",
	})

	updatedConfigStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
		"Plan": "business-4",
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: initialConfigStr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "plan", "business-8"),
				),
			},
			{
				Config: updatedConfigStr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "plan", "business-4"),
				),
			},
		},
	})
}

func TestAccAivenPG_deleting_additional_disk_size(t *testing.T) {
	resourceName := "aiven_pg.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	baseConfig := createBaseConfig(rName)

	initialConfigStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
		"AdditionalDiskSpace": "100GiB",
	})

	updatedConfigStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
		"AdditionalDiskSpace": "",
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: initialConfigStr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "additional_disk_space", "100GiB"),
				),
			},
			{
				Config: updatedConfigStr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckNoResourceAttr(resourceName, "additional_disk_space"),
				),
			},
		},
	})
}

func TestAccAivenPG_deleting_disk_size(t *testing.T) {
	resourceName := "aiven_pg.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	baseConfig := createBaseConfig(rName)

	initialConfigStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
		"DiskSpace": "90GiB",
	})

	updatedConfigStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
		"DiskSpace": "",
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: initialConfigStr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disk_space", "90GiB"),
				),
			},
			{
				Config: updatedConfigStr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "disk_space"),
				),
			},
		},
	})
}

func TestAccAivenPG_changing_disk_size(t *testing.T) {
	resourceName := "aiven_pg.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	baseConfig := createBaseConfig(rName)

	initialConfigStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
		"DiskSpace": "90GiB",
	})

	updatedConfigStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
		"DiskSpace": "100GiB",
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: initialConfigStr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disk_space", "90GiB"),
				),
			},
			{
				Config: updatedConfigStr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disk_space", "100GiB"),
				),
			},
		},
	})
}

// TestAccAivenPG_admin_creds tests admin creds in user_config
func TestAccAivenPG_admin_creds(t *testing.T) {
	resourceName := "aiven_pg.pg"
	prefix := "test-tf-acc-" + acctest.RandString(7)
	expectedURLPrefix := fmt.Sprintf("postgres://root:%s-password", prefix)

	baseConfig := createBaseConfig(prefix)

	configStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
		"Plan":          "startup-4",
		"AdminUsername": "root",
		"AdminPassword": prefix + "-password",
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: configStr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "service_uri", func(value string) error {
						if !strings.HasPrefix(value, expectedURLPrefix) {
							return fmt.Errorf("invalid service_uri, doesn't contain admin_username: %q", value)
						}
						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "pg_user_config.0.admin_username", "root"),
					resource.TestCheckResourceAttr(resourceName, "pg_user_config.0.admin_password", prefix+"-password"),
				),
			},
		},
	})
}

// PG service tests
func TestAccAivenServicePG_basic(t *testing.T) {
	resourceName := "aiven_pg.bar-pg"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	project := os.Getenv("AIVEN_PROJECT_NAME")

	baseConfig := createBaseConfig(rName)

	configStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
		"Plan": "startup-4",
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: configStr,
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_pg.common-pg"),
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common-pg"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
		},
	})
}

func TestAccAivenServicePG_termination_protection(t *testing.T) {
	resourceName := "aiven_pg.bar-pg"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	baseConfig := createBaseConfig(rName)

	configStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
		"Plan":                  "startup-4",
		"TerminationProtection": true,
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: configStr,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceTerminationProtection("data.aiven_pg.common-pg"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAivenServicePG_read_replica(t *testing.T) {
	resourceName := "aiven_pg.bar-pg"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	project := os.Getenv("AIVEN_PROJECT_NAME")

	baseConfig := createBaseConfig(rName)

	configStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
		"Plan":                   "startup-4",
		"ReadReplicaServiceName": fmt.Sprintf("test-acc-sr-replica-%s", rName),
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:                    configStr,
				PreventPostDestroyRefresh: true,
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_pg.common-pg"),
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common-pg"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
		},
	})
}

func TestAccAivenServicePG_custom_timeouts(t *testing.T) {
	resourceName := "aiven_pg.bar-pg"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	baseConfig := createBaseConfig(rName)

	configStr := generateConfigWithChanges(baseConfig, map[string]interface{}{
		"Plan":          "startup-4",
		"TimeoutCreate": "25m",
		"TimeoutUpdate": "30m",
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: configStr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "timeouts.create", "30m"),
					resource.TestCheckResourceAttr(resourceName, "timeouts.update", "30m"),
				),
			},
		},
	})
}

func testAccCheckAivenServiceTerminationProtection(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		projectName, serviceName, err := schemautil.SplitResourceID2(a["id"])
		if err != nil {
			return err
		}

		c := acc.GetTestAivenClient()

		ctx := context.Background()

		service, err := c.Services.Get(ctx, projectName, serviceName)
		if err != nil {
			return fmt.Errorf("cannot get service %s err: %w", serviceName, err)
		}

		if service.TerminationProtection == false {
			return fmt.Errorf("expected to get a termination_protection=true from Aiven")
		}

		// try to delete Aiven service with termination_protection enabled
		// should be an error from Aiven API
		err = c.Services.Delete(ctx, projectName, serviceName)
		if err == nil {
			return fmt.Errorf("termination_protection enabled should prevent from deletion of a service, deletion went OK")
		}

		// set service termination_protection to false to make Terraform Destroy plan work
		_, err = c.Services.Update(
			ctx,
			projectName,
			service.Name,
			aiven.UpdateServiceRequest{
				Cloud:                 service.CloudName,
				MaintenanceWindow:     &service.MaintenanceWindow,
				Plan:                  service.Plan,
				ProjectVPCID:          service.ProjectVPCID,
				Powered:               true,
				TerminationProtection: false,
				UserConfig:            service.UserConfig,
			},
		)
		if err != nil {
			return fmt.Errorf("unable to update Aiven service to set termination_protection=false err: %w", err)
		}

		return nil
	}
}

func testAccCheckAivenServicePGAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if !strings.Contains(a["service_type"], "pg") {
			return fmt.Errorf("expected to get a correct service_type from Aiven, got :%s", a["service_type"])
		}

		if a["pg_user_config.0.pg.0.idle_in_transaction_session_timeout"] != "900" {
			return fmt.Errorf("expected to get a correct idle_in_transaction_session_timeout from Aiven")
		}

		if a["pg_user_config.0.public_access.0.pg"] != "true" {
			return fmt.Errorf("expected to get a correct public_access.pg from Aiven")
		}

		if a["pg_user_config.0.public_access.0.pgbouncer"] != "false" {
			return fmt.Errorf("expected to get a correct public_access.pgbouncer from Aiven")
		}

		if a["pg_user_config.0.public_access.0.prometheus"] != "false" {
			return fmt.Errorf("expected to get a correct public_access.prometheus from Aiven")
		}

		if a["pg_user_config.0.service_to_fork_from"] != "" {
			return fmt.Errorf("expected to get a service_to_fork_from not set to any value")
		}

		if a["pg.0.uri"] == "" {
			return fmt.Errorf("expected to get a correct uri from Aiven")
		}

		if a["pg.0.uris.#"] == "" {
			return fmt.Errorf("expected to get uris from Aiven")
		}

		if a["pg.0.host"] == "" {
			return fmt.Errorf("expected to get a correct host from Aiven")
		}

		if a["pg.0.port"] == "" {
			return fmt.Errorf("expected to get a correct port from Aiven")
		}

		if a["pg.0.sslmode"] != "require" {
			return fmt.Errorf("expected to get a correct sslmode from Aiven")
		}

		if a["pg.0.user"] != "avnadmin" {
			return fmt.Errorf("expected to get a correct user from Aiven")
		}

		if a["pg.0.password"] == "" {
			return fmt.Errorf("expected to get a correct password from Aiven")
		}

		if a["pg.0.dbname"] != "defaultdb" {
			return fmt.Errorf("expected to get a correct dbname from Aiven")
		}

		if a["pg.0.params.#"] == "" {
			return fmt.Errorf("expected to get params from Aiven")
		}

		if a["pg.0.params.0.host"] == "" {
			return fmt.Errorf("expected to get a correct host from Aiven")
		}

		if a["pg.0.params.0.port"] == "" {
			return fmt.Errorf("expected to get a correct port from Aiven")
		}

		if a["pg.0.params.0.sslmode"] != "require" {
			return fmt.Errorf("expected to get a correct sslmode from Aiven")
		}

		if a["pg.0.params.0.user"] != "avnadmin" {
			return fmt.Errorf("expected to get a correct user from Aiven")
		}

		if a["pg.0.params.0.password"] == "" {
			return fmt.Errorf("expected to get a correct password from Aiven")
		}

		if a["pg.0.params.0.database_name"] != "defaultdb" {
			return fmt.Errorf("expected to get a correct database_name from Aiven")
		}

		if a["pg.0.max_connections"] != "100" && a["pg.0.max_connections"] != "200" {
			return fmt.Errorf("expected to get a correct max_connections from Aiven")
		}

		return nil
	}
}
