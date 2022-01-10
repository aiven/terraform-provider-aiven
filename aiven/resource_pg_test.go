// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAiven_pg(t *testing.T) {
	t.Parallel()

	resourceName := "aiven_pg.bar"

	t.Run("invalid disk sizes", func(tt *testing.T) {
		var (
			expectErrorRegexBadString = regexp.MustCompile(regexp.QuoteMeta("configured string must match ^[1-9][0-9]*(G|GiB)"))
		)
		rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

		resource.ParallelTest(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
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
			},
		})
	})

	t.Run("changing disk sizes", func(tt *testing.T) {
		rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

		resource.Test(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckAivenServiceResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccPGResourceWithDiskSize(rName, "90GiB"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenServicePGAttributes("data.aiven_pg.service"),
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
						testAccCheckAivenServicePGAttributes("data.aiven_pg.service"),
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
	})

	t.Run("deleting a disc size from the manifest", func(tt *testing.T) {
		rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

		resource.Test(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckAivenServiceResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccPGResourceWithDiskSize(rName, "90GiB"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenServicePGAttributes("data.aiven_pg.service"),
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
						testAccCheckAivenServicePGAttributes("data.aiven_pg.service"),
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
	})

	t.Run("changing plan of a service when disc size is not set", func(tt *testing.T) {
		rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

		resource.Test(tt, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(tt) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckAivenServiceResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccPGResourcePlanChange(rName, "business-8"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAivenServicePGAttributes("data.aiven_pg.service"),
						resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
						resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
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
						testAccCheckAivenServicePGAttributes("data.aiven_pg.service"),
						resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
						resource.TestCheckResourceAttr(resourceName, "state", "REBALANCING"),
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
	})
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
		
		data "aiven_pg" "service" {
		  service_name = aiven_pg.bar.service_name
		  project      = aiven_pg.bar.project
		
		  depends_on = [aiven_pg.bar]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, diskSize)
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
		
		data "aiven_pg" "service" {
		  service_name = aiven_pg.bar.service_name
		  project      = aiven_pg.bar.project
		
		  depends_on = [aiven_pg.bar]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name)
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
		
		data "aiven_pg" "service" {
		  service_name = aiven_pg.bar.service_name
		  project      = aiven_pg.bar.project
		
		  depends_on = [aiven_pg.bar]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), plan, name)
}
