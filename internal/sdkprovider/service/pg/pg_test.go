package pg_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenPG_no_existing_project(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccPGProjectDoesntExist(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAivenPG_invalid_disk_size(t *testing.T) {
	expectErrorRegexBadString := regexp.MustCompile(regexp.QuoteMeta("configured string must match ^[1-9][0-9]*(G|GiB)"))
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// bad strings
			{
				Config:      testAccPGResourceWithDiskSize(rName, "abc"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccPGResourceWithDiskSize(rName, "01MiB"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccPGResourceWithDiskSize(rName, "1234"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccPGResourceWithDiskSize(rName, "5TiB"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccPGResourceWithDiskSize(rName, " 1Gib "),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccPGResourceWithDiskSize(rName, "1 GiB"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			// bad disk sizes
			{
				Config:      testAccPGResourceWithDiskSize(rName, "1GiB"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("requested disk size is too small"),
			},
			{
				Config:      testAccPGResourceWithDiskSize(rName, "100000GiB"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("requested disk size is too large"),
			},
			{
				Config:      testAccPGResourceWithDiskSize(rName, "127GiB"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("requested disk size has to increase from: '.*' in increments of '.*'"),
			},
			{
				Config:      testAccPGResourceWithAdditionalDiskSize(rName, "127GiB"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("requested disk size has to increase from: '.*' in increments of '.*'"),
			},
			{
				Config:      testAccPGResourceWithAdditionalDiskSize(rName, "100000GiB"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("requested disk size is too large"),
			},
			{
				Config:      testAccPGResourceWithAdditionalDiskSize(rName, "abc"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccPGResourceWithAdditionalDiskSize(rName, "01MiB"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccPGResourceWithAdditionalDiskSize(rName, "1234"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:             testAccPGDoubleTagResource(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("tag keys should be unique"),
			},
		},
	})
}

func TestAccAivenPG_static_ips(t *testing.T) {
	resourceName := "aiven_pg.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGWithStaticIps(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "pg"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "service_username"),
					resource.TestCheckResourceAttrSet(resourceName, "service_password"),
					resource.TestCheckResourceAttrSet(resourceName, "service_host"),
					resource.TestCheckResourceAttrSet(resourceName, "service_port"),
					resource.TestCheckResourceAttrSet(resourceName, "service_uri"),
					resource.TestCheckResourceAttr(resourceName, "static_ips.#", "2"),
				),
			},
			{
				Config: testAccPGWithStaticIps(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "static_ips.#", "3"),
				),
			},
			{
				Config: testAccPGWithStaticIps(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "static_ips.#", "4"),
				),
			},
			{
				Config: testAccPGWithStaticIps(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "static_ips.#", "3"),
				),
			},
			{
				Config: testAccPGWithStaticIps(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "static_ips.#", "4"),
				),
			},
			{
				Config: testAccPGWithStaticIps(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "static_ips.#", "2"),
				),
			},
		},
	})
}

func TestAccAivenPG_changing_plan(t *testing.T) {
	resourceName := "aiven_pg.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGResourcePlanChange(rName, "business-8"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "pg"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "disk_space_used"),
				),
			},
			{
				Config: testAccPGResourcePlanChange(rName, "business-4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "pg"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "disk_space_used"),
				),
			},
		},
	})
}

func TestAccAivenPG_deleting_additional_disk_size(t *testing.T) {
	resourceName := "aiven_pg.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGResourceWithAdditionalDiskSize(rName, "20GiB"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "pg"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "disk_space_used", "100GiB"),
					resource.TestCheckResourceAttr(resourceName, "disk_space_default", "80GiB"),
					resource.TestCheckResourceAttr(resourceName, "additional_disk_space", "20GiB"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config: testAccPGResourceWithoutDiskSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "pg"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "disk_space_default", "80GiB"),
					resource.TestCheckResourceAttr(resourceName, "disk_space_used", "80GiB"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
		},
	})
}

func TestAccAivenPG_deleting_disk_size(t *testing.T) {
	resourceName := "aiven_pg.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGResourceWithDiskSize(rName, "90GiB"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "pg"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "disk_space", "90GiB"),
					resource.TestCheckResourceAttr(resourceName, "disk_space_used", "90GiB"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config: testAccPGResourceWithoutDiskSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "pg"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "disk_space", ""),
					resource.TestCheckResourceAttrSet(resourceName, "disk_space_used"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
		},
	})
}

func TestAccAivenPG_changing_disk_size(t *testing.T) {
	resourceName := "aiven_pg.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGResourceWithDiskSize(rName, "90GiB"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "pg"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "disk_space", "90GiB"),
					resource.TestCheckResourceAttr(resourceName, "disk_space_used", "90GiB"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config: testAccPGResourceWithDiskSize(rName, "100GiB"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "pg"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "disk_space", "100GiB"),
					resource.TestCheckResourceAttr(resourceName, "disk_space_used", "100GiB"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
		},
	})
}

func testAccPGWithStaticIps(name string, count int) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_static_ip" "ips" {
  count      = %d
  project    = data.aiven_project.foo.project
  cloud_name = "google-europe-west1"
}

resource "aiven_pg" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  static_ips              = toset(aiven_static_ip.ips[*].static_ip_address_id)

  pg_user_config {
    static_ips = true
  }
}

data "aiven_pg" "common" {
  service_name = aiven_pg.bar.service_name
  project      = aiven_pg.bar.project

  depends_on = [aiven_pg.bar]
}`, acc.ProjectName(), count, name)
}

func testAccPGResourceWithDiskSize(name, diskSize string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  disk_space              = "%s"

  pg_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
      log_min_duration_statement          = -1
    }
  }
}

data "aiven_pg" "common" {
  service_name = aiven_pg.bar.service_name
  project      = aiven_pg.bar.project

  depends_on = [aiven_pg.bar]
}`, acc.ProjectName(), name, diskSize)
}

func testAccPGResourceWithAdditionalDiskSize(name, diskSize string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  additional_disk_space   = "%s"

  pg_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
      log_min_duration_statement          = -1
    }
  }
}

data "aiven_pg" "common" {
  service_name = aiven_pg.bar.service_name
  project      = aiven_pg.bar.project

  depends_on = [aiven_pg.bar]
}`, acc.ProjectName(), name, diskSize)
}

func testAccPGResourceWithoutDiskSize(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  tag {
    key   = "test"
    value = "val"
  }

  pg_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
      log_min_duration_statement          = -1
    }
  }
}

data "aiven_pg" "common" {
  service_name = aiven_pg.bar.service_name
  project      = aiven_pg.bar.project

  depends_on = [aiven_pg.bar]
}`, acc.ProjectName(), name)
}

func testAccPGResourcePlanChange(name, plan string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "%s"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  tag {
    key   = "test"
    value = "val"
  }

  pg_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
      log_min_duration_statement          = -1
    }
  }
}

data "aiven_pg" "common" {
  service_name = aiven_pg.bar.service_name
  project      = aiven_pg.bar.project

  depends_on = [aiven_pg.bar]
}`, acc.ProjectName(), plan, name)
}

func testAccPGDoubleTagResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_pg" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  tag {
    key   = "test"
    value = "val"
  }
  tag {
    key   = "test"
    value = "val2"
  }

  pg_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
      log_min_duration_statement          = -1
    }
  }
}

data "aiven_pg" "common" {
  service_name = aiven_pg.bar.service_name
  project      = aiven_pg.bar.project

  depends_on = [aiven_pg.bar]
}`, acc.ProjectName(), name)
}

func testAccPGProjectDoesntExist() string {
	return `
resource "aiven_pg" "bar" {
  project                 = "wrong-project-name"
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-1"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  disk_space              = "100GiB"
}
`
}

// TestAccAivenPG_admin_creds tests admin creds in user_config
func TestAccAivenPG_admin_creds(t *testing.T) {
	resourceName := "aiven_pg.pg"
	prefix := "test-tf-acc-" + acctest.RandString(7)
	project := acc.ProjectName()
	expectedURLPrefix := fmt.Sprintf("postgres://root:%s-password", prefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGResourceAdminCreds(prefix, project),
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

// testAccPGResourceAdminCreds returns config TestAccAivenPG_admin_creds
func testAccPGResourceAdminCreds(prefix, project string) string {
	return fmt.Sprintf(`
resource "aiven_pg" "pg" {
  project      = %[2]q
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "%[1]s-pg"

  pg_user_config {
    admin_username = "root"
    admin_password = "%[1]s-password"
  }
}
`, prefix, project)
}

// PG service tests
func TestAccAivenServicePG_basic(t *testing.T) {
	resourceName := "aiven_pg.bar-pg"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_pg.common-pg"),
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common-pg"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGTerminationProtectionServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceTerminationProtection("data.aiven_pg.common-pg"),
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_pg.common-pg"),
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common-pg"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:                    testAccPGReadReplicaServiceResource(rName),
				PreventPostDestroyRefresh: true,
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_pg.common-pg"),
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common-pg"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
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

func TestAccAivenServicePG_custom_timeouts(t *testing.T) {
	resourceName := "aiven_pg.bar-pg"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGServiceCustomTimeoutsResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_pg.common-pg"),
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common-pg"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
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

func testAccPGServiceResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo-pg" {
  project = "%s"
}

resource "aiven_pg" "bar-pg" {
  project                 = data.aiven_project.foo-pg.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  pg_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

data "aiven_pg" "common-pg" {
  service_name = aiven_pg.bar-pg.service_name
  project      = aiven_pg.bar-pg.project

  depends_on = [aiven_pg.bar-pg]
}`, acc.ProjectName(), name)
}

func testAccPGServiceCustomTimeoutsResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo-pg" {
  project = "%s"
}

resource "aiven_pg" "bar-pg" {
  project                 = data.aiven_project.foo-pg.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  timeouts {
    create = "25m"
    update = "30m"
  }

  pg_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

data "aiven_pg" "common-pg" {
  service_name = aiven_pg.bar-pg.service_name
  project      = aiven_pg.bar-pg.project

  depends_on = [aiven_pg.bar-pg]
}`, acc.ProjectName(), name)
}

func testAccPGTerminationProtectionServiceResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo-pg" {
  project = "%s"
}

resource "aiven_pg" "bar-pg" {
  project                 = data.aiven_project.foo-pg.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  termination_protection  = true

  pg_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

data "aiven_pg" "common-pg" {
  service_name = aiven_pg.bar-pg.service_name
  project      = aiven_pg.bar-pg.project

  depends_on = [aiven_pg.bar-pg]
}`, acc.ProjectName(), name)
}

func testAccPGReadReplicaServiceResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo-pg" {
  project = "%s"
}

resource "aiven_pg" "bar-pg" {
  project                 = data.aiven_project.foo-pg.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  pg_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

resource "aiven_pg" "bar-replica" {
  project                 = data.aiven_project.foo-pg.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-repica-%s"
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
    source_service_name = aiven_pg.bar-pg.service_name
  }

  depends_on = [aiven_pg.bar-pg]
}

resource "aiven_service_integration" "pg-readreplica" {
  project                  = data.aiven_project.foo-pg.project
  integration_type         = "read_replica"
  source_service_name      = aiven_pg.bar-pg.service_name
  destination_service_name = aiven_pg.bar-replica.service_name

  depends_on = [aiven_pg.bar-replica]
}

data "aiven_pg" "common-pg" {
  service_name = aiven_pg.bar-pg.service_name
  project      = aiven_pg.bar-pg.project

  depends_on = [aiven_pg.bar-pg]
}`, acc.ProjectName(), name, name)
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

func TestAccAivenServicePG_disaster_recovery(t *testing.T) {
	primaryName := "aiven_pg.primary"
	secondaryName := "aiven_pg.secondary"
	projectName := acc.ProjectName()
	randStr := acc.RandStr()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGServiceResourceDisasterRecovery(projectName, randStr),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes(primaryName),
					acc.TestAccCheckAivenServiceCommonAttributes(secondaryName),
					// One service integration should be created
					resource.TestCheckResourceAttr(secondaryName, "service_integrations.#", "1"),
					// Which is of type disaster_recovery
					resource.TestCheckTypeSetElemNestedAttrs(secondaryName, "service_integrations.*", map[string]string{
						"integration_type":    "disaster_recovery",
						"source_service_name": "test-acc-primary-" + randStr,
					}),
				),
			},
		},
	})
}

func testAccPGServiceResourceDisasterRecovery(project, randStr string) string {
	return fmt.Sprintf(`
resource "aiven_pg" "primary" {
  project                 = "%[1]s"
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-primary-%[2]s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_pg" "secondary" {
  project                 = aiven_pg.primary.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-secondary-%[2]s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  service_integrations {
    integration_type    = "disaster_recovery"
    source_service_name = "test-acc-primary-%[2]s"
  }
}
`, project, randStr)
}

func TestAccAivenServicePG_user_config_zero_values(t *testing.T) {
	projectName := acc.ProjectName()
	serviceName := "test-acc-" + acc.RandStr()
	resourceName := "aiven_pg.foo"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				// Creates a PG service without a user config.
				// The plan is never suppressed for new services,
				// so the resource must have and ID to make the diff suppressor work.
				Config: testAccAivenServicePGUserConfigZeroValues(projectName, serviceName, ``),
			},
			{
				// Adds `true` value to change it later to `false`
				Config: testAccAivenServicePGUserConfigZeroValues(
					projectName, serviceName,
					`pg_user_config {
						pg {
						  jit = true
						}
					  }`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pg_user_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pg_user_config.0.pg.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pg_user_config.0.pg.0.jit", "true"),
					// public_access will be set with the next step, here it proves the block does not exist
					resource.TestCheckResourceAttr(resourceName, "pg_user_config.0.public_access.#", "0"),
				),
			},
			{
				// Sets existing field to `false` (proves the change is detected for an existing field)
				// Adds `public_access` block with all fields set to `false` (proves the changes are detected for new fields)
				Config: testAccAivenServicePGUserConfigZeroValues(
					projectName, serviceName,
					`pg_user_config {
						pg {
						  jit = false
						}

						public_access {
						  pg         = false
						  pgbouncer  = false
						  prometheus = false
						}
					  }`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pg_user_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pg_user_config.0.pg.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pg_user_config.0.pg.0.jit", "false"),
					resource.TestCheckResourceAttr(resourceName, "pg_user_config.0.public_access.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pg_user_config.0.public_access.0.pg", "false"),
					resource.TestCheckResourceAttr(resourceName, "pg_user_config.0.public_access.0.pgbouncer", "false"),
					resource.TestCheckResourceAttr(resourceName, "pg_user_config.0.public_access.0.prometheus", "false"),
				),
			},
		},
	})
}

func testAccAivenServicePGUserConfigZeroValues(projectName, serviceName, userConfig string) string {
	return fmt.Sprintf(`
resource "aiven_pg" "foo" {
  project      = %[1]q
  service_name = %[2]q
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"

  %s
}
`, projectName, serviceName, userConfig)
}
