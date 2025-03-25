package alloydbomni_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/alloydbomni"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenAlloyDBOmni_invalid_disk_size(t *testing.T) {
	expectErrorRegexBadString := regexp.MustCompile(regexp.QuoteMeta("configured string must match ^[1-9][0-9]*(G|GiB)"))
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// bad strings
			{
				Config:      testAccAlloyDBOmniResourceWithDiskSize(rName, "abc"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccAlloyDBOmniResourceWithDiskSize(rName, "01MiB"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccAlloyDBOmniResourceWithDiskSize(rName, "1234"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccAlloyDBOmniResourceWithDiskSize(rName, "5TiB"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccAlloyDBOmniResourceWithDiskSize(rName, " 1Gib "),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccAlloyDBOmniResourceWithDiskSize(rName, "1 GiB"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			// bad disk sizes
			{
				Config:      testAccAlloyDBOmniResourceWithDiskSize(rName, "1GiB"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("requested disk size is too small"),
			},
			{
				Config:      testAccAlloyDBOmniResourceWithDiskSize(rName, "100000GiB"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("requested disk size is too large"),
			},
			{
				Config:      testAccAlloyDBOmniResourceWithDiskSize(rName, "127GiB"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("requested disk size has to increase from: '.*' in increments of '.*'"),
			},
			{
				Config:      testAccAlloyDBOmniResourceWithAdditionalDiskSize(rName, "127GiB"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("requested disk size has to increase from: '.*' in increments of '.*'"),
			},
			{
				Config:      testAccAlloyDBOmniResourceWithAdditionalDiskSize(rName, "100000GiB"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("requested disk size is too large"),
			},
			{
				Config:      testAccAlloyDBOmniResourceWithAdditionalDiskSize(rName, "abc"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccAlloyDBOmniResourceWithAdditionalDiskSize(rName, "01MiB"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:      testAccAlloyDBOmniResourceWithAdditionalDiskSize(rName, "1234"),
				PlanOnly:    true,
				ExpectError: expectErrorRegexBadString,
			},
			{
				Config:             testAccAlloyDBOmniDoubleTagResource(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("tag keys should be unique"),
			},
		},
	})
}

func TestAccAivenAlloyDBOmni_static_ips(t *testing.T) {
	resourceName := "aiven_alloydbomni.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniWithStaticIps(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "alloydbomni"),
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
				Config: testAccAlloyDBOmniWithStaticIps(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "static_ips.#", "3"),
				),
			},
			{
				Config: testAccAlloyDBOmniWithStaticIps(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "static_ips.#", "4"),
				),
			},
			{
				Config: testAccAlloyDBOmniWithStaticIps(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "static_ips.#", "3"),
				),
			},
			{
				Config: testAccAlloyDBOmniWithStaticIps(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "static_ips.#", "4"),
				),
			},
			{
				Config: testAccAlloyDBOmniWithStaticIps(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "static_ips.#", "2"),
				),
			},
		},
	})
}

func TestAccAivenAlloyDBOmni_changing_plan(t *testing.T) {
	resourceName := "aiven_alloydbomni.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniResourcePlanChange(rName, "business-8"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceAlloyDBOmniAttributes("data.aiven_alloydbomni.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "alloydbomni"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "disk_space_used"),
				),
			},
			{
				Config: testAccAlloyDBOmniResourcePlanChange(rName, "business-4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceAlloyDBOmniAttributes("data.aiven_alloydbomni.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "alloydbomni"),
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

func TestAccAivenAlloyDBOmni_deleting_additional_disk_size(t *testing.T) {
	resourceName := "aiven_alloydbomni.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniResourceWithAdditionalDiskSize(rName, "20GiB"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceAlloyDBOmniAttributes("data.aiven_alloydbomni.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "alloydbomni"),
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
				Config: testAccAlloyDBOmniResourceWithoutDiskSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceAlloyDBOmniAttributes("data.aiven_alloydbomni.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "alloydbomni"),
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

func TestAccAivenAlloyDBOmni_deleting_disk_size(t *testing.T) {
	resourceName := "aiven_alloydbomni.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniResourceWithDiskSize(rName, "90GiB"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceAlloyDBOmniAttributes("data.aiven_alloydbomni.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "alloydbomni"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "disk_space", "90GiB"),
					resource.TestCheckResourceAttr(resourceName, "disk_space_used", "90GiB"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config: testAccAlloyDBOmniResourceWithoutDiskSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceAlloyDBOmniAttributes("data.aiven_alloydbomni.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "alloydbomni"),
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

func TestAccAivenAlloyDBOmni_changing_disk_size(t *testing.T) {
	resourceName := "aiven_alloydbomni.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniResourceWithDiskSize(rName, "90GiB"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceAlloyDBOmniAttributes("data.aiven_alloydbomni.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "alloydbomni"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "disk_space", "90GiB"),
					resource.TestCheckResourceAttr(resourceName, "disk_space_used", "90GiB"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config: testAccAlloyDBOmniResourceWithDiskSize(rName, "100GiB"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceAlloyDBOmniAttributes("data.aiven_alloydbomni.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", acc.ProjectName()),
					resource.TestCheckResourceAttr(resourceName, "service_type", "alloydbomni"),
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

func testAccAlloyDBOmniWithStaticIps(name string, count int) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_static_ip" "ips" {
  count      = %d
  project    = data.aiven_project.foo.project
  cloud_name = "google-europe-west1"
}

resource "aiven_alloydbomni" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  static_ips              = toset(aiven_static_ip.ips[*].static_ip_address_id)

  alloydbomni_user_config {
    static_ips = true
  }
}

data "aiven_alloydbomni" "common" {
  service_name = aiven_alloydbomni.bar.service_name
  project      = aiven_alloydbomni.bar.project

  depends_on = [aiven_alloydbomni.bar]
}`, acc.ProjectName(), count, name)
}

func testAccAlloyDBOmniResourceWithDiskSize(name, diskSize string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  disk_space              = "%s"

  alloydbomni_user_config {
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

data "aiven_alloydbomni" "common" {
  service_name = aiven_alloydbomni.bar.service_name
  project      = aiven_alloydbomni.bar.project

  depends_on = [aiven_alloydbomni.bar]
}`, acc.ProjectName(), name, diskSize)
}

func testAccAlloyDBOmniResourceWithAdditionalDiskSize(name, diskSize string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  additional_disk_space   = "%s"

  alloydbomni_user_config {
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

data "aiven_alloydbomni" "common" {
  service_name = aiven_alloydbomni.bar.service_name
  project      = aiven_alloydbomni.bar.project

  depends_on = [aiven_alloydbomni.bar]
}`, acc.ProjectName(), name, diskSize)
}

func testAccAlloyDBOmniResourceWithoutDiskSize(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar" {
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

  alloydbomni_user_config {
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

data "aiven_alloydbomni" "common" {
  service_name = aiven_alloydbomni.bar.service_name
  project      = aiven_alloydbomni.bar.project

  depends_on = [aiven_alloydbomni.bar]
}`, acc.ProjectName(), name)
}

func testAccAlloyDBOmniResourcePlanChange(name, plan string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar" {
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

  alloydbomni_user_config {
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

data "aiven_alloydbomni" "common" {
  service_name = aiven_alloydbomni.bar.service_name
  project      = aiven_alloydbomni.bar.project

  depends_on = [aiven_alloydbomni.bar]
}`, acc.ProjectName(), plan, name)
}

func testAccAlloyDBOmniDoubleTagResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar" {
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

  alloydbomni_user_config {
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

data "aiven_alloydbomni" "common" {
  service_name = aiven_alloydbomni.bar.service_name
  project      = aiven_alloydbomni.bar.project

  depends_on = [aiven_alloydbomni.bar]
}`, acc.ProjectName(), name)
}

// TestAccAivenAlloyDBOmni_admin_creds tests admin creds in user_config
func TestAccAivenAlloyDBOmni_admin_creds(t *testing.T) {
	resourceName := "aiven_alloydbomni.alloydbomni"
	prefix := "test-tf-acc-" + acctest.RandString(7)
	project := acc.ProjectName()
	expectedURLPrefix := fmt.Sprintf("postgres://root:%s-password", prefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniResourceAdminCreds(prefix, project),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "service_uri", func(value string) error {
						if !strings.HasPrefix(value, expectedURLPrefix) {
							return fmt.Errorf("invalid service_uri, doesn't contain admin_username: %q", value)
						}
						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "alloydbomni_user_config.0.admin_username", "root"),
					resource.TestCheckResourceAttr(resourceName, "alloydbomni_user_config.0.admin_password", prefix+"-password"),
				),
			},
		},
	})
}

// testAccAlloyDBOmniResourceAdminCreds returns config TestAccAivenAlloyDBOmni_admin_creds
func testAccAlloyDBOmniResourceAdminCreds(prefix, project string) string {
	return fmt.Sprintf(`
resource "aiven_alloydbomni" "alloydbomni" {
  project      = %[2]q
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "%[1]s-alloydbomni"

  alloydbomni_user_config {
    admin_username = "root"
    admin_password = "%[1]s-password"
  }
}
	`, prefix, project)
}

// AlloyDBOmni service tests
func TestAccAivenServiceAlloyDBOmni_basic(t *testing.T) {
	resourceName := "aiven_alloydbomni.bar-alloydbomni"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_alloydbomni.common-alloydbomni"),
					testAccCheckAivenServiceAlloyDBOmniAttributes("data.aiven_alloydbomni.common-alloydbomni"),
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

func TestAccAivenServiceAlloyDBOmni_termination_protection(t *testing.T) {
	resourceName := "aiven_alloydbomni.bar-alloydbomni"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniTerminationProtectionServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceTerminationProtection("data.aiven_alloydbomni.common-alloydbomni"),
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_alloydbomni.common-alloydbomni"),
					testAccCheckAivenServiceAlloyDBOmniAttributes("data.aiven_alloydbomni.common-alloydbomni"),
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

func TestAccAivenServiceAlloyDBOmni_read_replica(t *testing.T) {
	resourceName := "aiven_alloydbomni.bar-alloydbomni"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:                    testAccAlloyDBOmniReadReplicaServiceResource(rName),
				PreventPostDestroyRefresh: true,
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_alloydbomni.common-alloydbomni"),
					testAccCheckAivenServiceAlloyDBOmniAttributes("data.aiven_alloydbomni.common-alloydbomni"),
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

func TestAccAivenServiceAlloyDBOmni_custom_timeouts(t *testing.T) {
	resourceName := "aiven_alloydbomni.bar-alloydbomni"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlloyDBOmniServiceCustomTimeoutsResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_alloydbomni.common-alloydbomni"),
					testAccCheckAivenServiceAlloyDBOmniAttributes("data.aiven_alloydbomni.common-alloydbomni"),
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

func testAccAlloyDBOmniServiceResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo-alloydbomni" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar-alloydbomni" {
  project                 = data.aiven_project.foo-alloydbomni.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  alloydbomni_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

data "aiven_alloydbomni" "common-alloydbomni" {
  service_name = aiven_alloydbomni.bar-alloydbomni.service_name
  project      = aiven_alloydbomni.bar-alloydbomni.project

  depends_on = [aiven_alloydbomni.bar-alloydbomni]
}`, acc.ProjectName(), name)
}

func testAccAlloyDBOmniServiceCustomTimeoutsResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo-alloydbomni" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar-alloydbomni" {
  project                 = data.aiven_project.foo-alloydbomni.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  timeouts {
    create = "25m"
    update = "30m"
  }

  alloydbomni_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

data "aiven_alloydbomni" "common-alloydbomni" {
  service_name = aiven_alloydbomni.bar-alloydbomni.service_name
  project      = aiven_alloydbomni.bar-alloydbomni.project

  depends_on = [aiven_alloydbomni.bar-alloydbomni]
}`, acc.ProjectName(), name)
}

func testAccAlloyDBOmniTerminationProtectionServiceResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo-alloydbomni" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar-alloydbomni" {
  project                 = data.aiven_project.foo-alloydbomni.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  termination_protection  = true

  alloydbomni_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

data "aiven_alloydbomni" "common-alloydbomni" {
  service_name = aiven_alloydbomni.bar-alloydbomni.service_name
  project      = aiven_alloydbomni.bar-alloydbomni.project

  depends_on = [aiven_alloydbomni.bar-alloydbomni]
}`, acc.ProjectName(), name)
}

func testAccAlloyDBOmniReadReplicaServiceResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo-alloydbomni" {
  project = "%s"
}

resource "aiven_alloydbomni" "bar-alloydbomni" {
  project                 = data.aiven_project.foo-alloydbomni.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  alloydbomni_user_config {
    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
    }
  }
}

resource "aiven_alloydbomni" "bar-replica" {
  project                 = data.aiven_project.foo-alloydbomni.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-repica-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  alloydbomni_user_config {
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
    source_service_name = aiven_alloydbomni.bar-alloydbomni.service_name
  }

  depends_on = [aiven_alloydbomni.bar-alloydbomni]
}

resource "aiven_service_integration" "alloydbomni-readreplica" {
  project                  = data.aiven_project.foo-alloydbomni.project
  integration_type         = "read_replica"
  source_service_name      = aiven_alloydbomni.bar-alloydbomni.service_name
  destination_service_name = aiven_alloydbomni.bar-replica.service_name

  depends_on = [aiven_alloydbomni.bar-replica]
}

data "aiven_alloydbomni" "common-alloydbomni" {
  service_name = aiven_alloydbomni.bar-alloydbomni.service_name
  project      = aiven_alloydbomni.bar-alloydbomni.project

  depends_on = [aiven_alloydbomni.bar-alloydbomni]
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

func testAccCheckAivenServiceAlloyDBOmniAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if !strings.Contains(a["service_type"], "alloydbomni") {
			return fmt.Errorf("expected to get a correct service_type from Aiven, got :%s", a["service_type"])
		}

		if a["alloydbomni_user_config.0.pg.0.idle_in_transaction_session_timeout"] != "900" {
			return fmt.Errorf("expected to get a correct idle_in_transaction_session_timeout from Aiven")
		}

		if a["alloydbomni_user_config.0.public_access.0.pg"] != "true" {
			return fmt.Errorf("expected to get a correct public_access.alloydbomni from Aiven")
		}

		if a["alloydbomni_user_config.0.public_access.0.pgbouncer"] != "false" {
			return fmt.Errorf("expected to get a correct public_access.alloydbomnibouncer from Aiven")
		}

		if a["alloydbomni_user_config.0.public_access.0.prometheus"] != "false" {
			return fmt.Errorf("expected to get a correct public_access.prometheus from Aiven")
		}

		if a["alloydbomni_user_config.0.service_to_fork_from"] != "" {
			return fmt.Errorf("expected to get a service_to_fork_from not set to any value")
		}

		if a["alloydbomni.0.uri"] == "" {
			return fmt.Errorf("expected to get a correct uri from Aiven")
		}

		if a["alloydbomni.0.uris.#"] == "" {
			return fmt.Errorf("expected to get uris from Aiven")
		}

		if a["alloydbomni.0.host"] == "" {
			return fmt.Errorf("expected to get a correct host from Aiven")
		}

		if a["alloydbomni.0.port"] == "" {
			return fmt.Errorf("expected to get a correct port from Aiven")
		}

		if a["alloydbomni.0.sslmode"] != "require" {
			return fmt.Errorf("expected to get a correct sslmode from Aiven")
		}

		if a["alloydbomni.0.user"] != "avnadmin" {
			return fmt.Errorf("expected to get a correct user from Aiven")
		}

		if a["alloydbomni.0.password"] == "" {
			return fmt.Errorf("expected to get a correct password from Aiven")
		}

		if a["alloydbomni.0.dbname"] != "defaultdb" {
			return fmt.Errorf("expected to get a correct dbname from Aiven")
		}

		if a["alloydbomni.0.params.#"] == "" {
			return fmt.Errorf("expected to get params from Aiven")
		}

		if a["alloydbomni.0.params.0.host"] == "" {
			return fmt.Errorf("expected to get a correct host from Aiven")
		}

		if a["alloydbomni.0.params.0.port"] == "" {
			return fmt.Errorf("expected to get a correct port from Aiven")
		}

		if a["alloydbomni.0.params.0.sslmode"] != "require" {
			return fmt.Errorf("expected to get a correct sslmode from Aiven")
		}

		if a["alloydbomni.0.params.0.user"] != "avnadmin" {
			return fmt.Errorf("expected to get a correct user from Aiven")
		}

		if a["alloydbomni.0.params.0.password"] == "" {
			return fmt.Errorf("expected to get a correct password from Aiven")
		}

		if a["alloydbomni.0.params.0.database_name"] != "defaultdb" {
			return fmt.Errorf("expected to get a correct database_name from Aiven")
		}

		if a["alloydbomni.0.max_connections"] != "100" && a["alloydbomni.0.max_connections"] != "200" {
			return fmt.Errorf("expected to get a correct max_connections from Aiven")
		}

		return nil
	}
}

func TestAccAivenServiceAlloyDBOmni_service_account_credentials(t *testing.T) {
	ctx := context.Background()
	client, err := acc.GetTestGenAivenClient()
	if err != nil {
		t.Skipf("cannot get aiven client: %s", err)
	}

	project := acc.ProjectName()
	resourceName := "aiven_alloydbomni.foo"
	serviceName := fmt.Sprintf("test-acc-sr-%s", acc.RandStr())

	// Service account credentials are managed by its own API
	// When Terraform fails to create a service because of this field,
	// the whole resource is tainted, and must be replaced
	serviceAccountCredentialsInvalid := testAccAivenServiceAlloyDBOmniServiceAccountCredentials(
		project, serviceName,
		`{
		  "private_key": "-----BEGIN PRIVATE KEY--.........----END PRIVATE KEY-----\n",
		  "client_email": "example@aiven.io",
		  "client_id": "example_user_id",
		  "type": "service_account",
		  "project_id": "example_project_id"
		}`,
	)
	serviceAccountCredentialsEmpty := testAccAivenServiceAlloyDBOmniServiceAccountCredentials(
		project, serviceName, "",
	)
	serviceAccountCredentialsValid := testAccAivenServiceAlloyDBOmniServiceAccountCredentials(
		project, serviceName, getTestServiceAccountCredentials("foo"),
	)
	serviceAccountCredentialsValidModified := testAccAivenServiceAlloyDBOmniServiceAccountCredentials(
		project, serviceName, getTestServiceAccountCredentials("bar"),
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				// 0. Invalid credentials
				Config:      serviceAccountCredentialsInvalid,
				ExpectError: regexp.MustCompile(`private_key_id is required`),
			},
			{
				// 1. No credential initially
				Config: serviceAccountCredentialsEmpty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "service_account_credentials"),
				),
			},
			{
				// 2. Credentials are set
				Config: serviceAccountCredentialsValid,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "service_account_credentials"),
				),
			},
			{
				// 3. Updates the credentials
				Config: serviceAccountCredentialsValidModified,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "service_account_credentials"),
					// Validates they key has been updated
					func(_ *terraform.State) error {
						rsp, err := client.AlloyDbOmniGoogleCloudPrivateKeyIdentify(ctx, project, serviceName)
						if err != nil {
							return fmt.Errorf("cannot get alloydbomni service account credentials: %w", err)
						}
						if rsp.PrivateKeyId != "bar" {
							return fmt.Errorf(`invalid private_key_id: expected "bar", got %q`, rsp.PrivateKeyId)
						}
						return nil
					},
				),
			},
			{
				// 4. Removes the credentials
				Config: serviceAccountCredentialsEmpty,
				Check: resource.ComposeTestCheckFunc(
					// It looks like TF can't unset an attribute, when it was set.
					// So I can't use TestCheckNoResourceAttr here.
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials", ""),
					// Validates they key has been deleted
					func(_ *terraform.State) error {
						rsp, err := client.AlloyDbOmniGoogleCloudPrivateKeyIdentify(ctx, project, serviceName)
						if err != nil {
							return fmt.Errorf("cannot get alloydbomni service account credentials: %w", err)
						}
						if rsp.PrivateKeyId != "" {
							return fmt.Errorf(`invalid private_key_id: expected empty, got %q`, rsp.PrivateKeyId)
						}
						return nil
					},
				),
			},
			{
				// 5. Re-applies the credential, so we can check unexpected remote state changes in the next step
				Config: serviceAccountCredentialsValidModified,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "service_account_credentials"),
				),
			},
			{
				// 6. Same config. Modifies the remove state, expects non-empty plan
				Config:             serviceAccountCredentialsValidModified,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				PreConfig: func() {
					// Modifies remote state to simulate a change
					req := &alloydbomni.AlloyDbOmniGoogleCloudPrivateKeySetIn{
						PrivateKey: getTestServiceAccountCredentials("egg"),
					}
					_, err = client.AlloyDbOmniGoogleCloudPrivateKeySet(ctx, project, serviceName, req)
					if err != nil {
						t.Fatal(err)
					}
				},
			},
		},
	})
}

func getTestServiceAccountCredentials(privateKeyID string) string {
	return fmt.Sprintf(`{
	  "private_key_id": %q,
	  "private_key": "-----BEGIN PRIVATE KEY--.........----END PRIVATE KEY-----\n",
	  "client_email": "example@aiven.io",
	  "client_id": "example_user_id",
	  "type": "service_account",
	  "project_id": "example_project_id"
	}`, privateKeyID)
}

func testAccAivenServiceAlloyDBOmniServiceAccountCredentials(project, name, privateKey string) string {
	var p string
	if privateKey != "" {
		p = fmt.Sprintf(`
  service_account_credentials = <<EOF
    %s
    EOF`, privateKey)
	}

	return fmt.Sprintf(`
resource "aiven_alloydbomni" "foo" {
  project                 = %q
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = %q
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  %s
}
`, project, name, p)
}
