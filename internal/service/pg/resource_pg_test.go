package pg_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAivenPG_invalid_disc_size(t *testing.T) {
	expectErrorRegexBadString := regexp.MustCompile(regexp.QuoteMeta("configured string must match ^[1-9][0-9]*(G|GiB)"))
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
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
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGWithStaticIps(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
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
				Config: testAccPGWithStaticIpsAddition(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "static_ips.#", "3"),
				),
			},
			{
				Config: testAccPGWithStaticIpsPreDeletion(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "static_ips.#", "2"),
				),
			},
			{
				Config: testAccPGWithStaticIpsDeletion(rName),
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
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGResourcePlanChange(rName, "business-8"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
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
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
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

func TestAccAivenPG_deleting_disc_size(t *testing.T) {
	resourceName := "aiven_pg.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGResourceWithDiskSize(rName, "90GiB"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
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
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
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

func TestAccAivenPG_changing_disc_size(t *testing.T) {
	resourceName := "aiven_pg.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPGResourceWithDiskSize(rName, "90GiB"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServicePGAttributes("data.aiven_pg.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
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
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
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

func testAccPGWithStaticIps(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_static_ip" "ips" {
  count      = 2
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
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccPGWithStaticIpsAddition(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_static_ip" "ips" {
  count      = 3
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
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccPGWithStaticIpsPreDeletion(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_static_ip" "ips" {
  count      = 3
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
  static_ips = toset([
    aiven_static_ip.ips[0].static_ip_address_id,
    aiven_static_ip.ips[1].static_ip_address_id,
  ])

  pg_user_config {
    static_ips = true
  }
}

data "aiven_pg" "common" {
  service_name = aiven_pg.bar.service_name
  project      = aiven_pg.bar.project

  depends_on = [aiven_pg.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccPGWithStaticIpsDeletion(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_static_ip" "ips" {
  count      = 2
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
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
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
}`, os.Getenv("AIVEN_PROJECT_NAME"), name, diskSize)
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
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
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
}`, os.Getenv("AIVEN_PROJECT_NAME"), plan, name)
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
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}
