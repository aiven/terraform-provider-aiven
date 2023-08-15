package redis_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

<<<<<<< HEAD
	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
=======
	acctest3 "github.com/aiven/terraform-provider-aiven/internal/acctest"
>>>>>>> fd0b89f6 (feat(frameworkprovider): organization resource and data source (#1283))
)

func TestAccAiven_redis(t *testing.T) {
	resourceName := "aiven_redis.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRedisResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_redis.common"),
					testAccCheckAivenServiceRedisAttributes("data.aiven_redis.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "redis"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "service_username"),
					resource.TestCheckResourceAttrSet(resourceName, "service_password"),
					resource.TestCheckResourceAttrSet(resourceName, "service_host"),
					resource.TestCheckResourceAttrSet(resourceName, "service_port"),
					resource.TestCheckResourceAttrSet(resourceName, "service_uri"),
				),
			},
			{
				Config:             testAccRedisDoubleTagResource(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("tag keys should be unique"),
			},
		},
	})
}

func testAccRedisResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_redis" "bar" {
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

  redis_user_config {
    redis_maxmemory_policy = "allkeys-random"

    public_access {
      redis = true
    }
  }
}

data "aiven_redis" "common" {
  service_name = aiven_redis.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_redis.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccRedisDoubleTagResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_redis" "bar" {
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

  redis_user_config {
    redis_maxmemory_policy = "allkeys-random"

    public_access {
      redis = true
    }
  }
}

data "aiven_redis" "common" {
  service_name = aiven_redis.bar.service_name
  project      = data.aiven_project.foo.project

  depends_on = [aiven_redis.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

// Redis service tests
func TestAccAivenService_redis(t *testing.T) {
	resourceName := "aiven_redis.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRedisServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_redis.common"),
					testAccCheckAivenServiceRedisAttributes("data.aiven_redis.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "redis"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config: testAccRedisServiceResourceWithPersistenceOff(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_redis.common"),
					testAccCheckAivenServiceRedisAttributes("data.aiven_redis.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "redis"),
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

func testAccRedisServiceResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_redis" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  redis_user_config {
    redis_maxmemory_policy = "allkeys-random"

    public_access {
      redis = true
    }
  }
}

data "aiven_redis" "common" {
  service_name = aiven_redis.bar.service_name
  project      = aiven_redis.bar.project

  depends_on = [aiven_redis.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccRedisServiceResourceWithPersistenceOff(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_redis" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  redis_user_config {
    redis_persistence      = "off"
    redis_maxmemory_policy = "allkeys-random"

    public_access {
      redis = true
    }
  }
}

data "aiven_redis" "common" {
  service_name = aiven_redis.bar.service_name
  project      = aiven_redis.bar.project

  depends_on = [aiven_redis.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccCheckAivenServiceRedisAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["service_type"] != "redis" {
			return fmt.Errorf("expected to get a correct service type from Aiven, got :%s", a["service_type"])
		}

		if a["redis_user_config.0.redis_maxmemory_policy"] != "allkeys-random" {
			return fmt.Errorf("expected to get a correct redis_maxmemory_policy from Aiven")
		}

		if a["redis_user_config.0.public_access.0.redis"] != "true" {
			return fmt.Errorf("expected to get a correct public_access.redis from Aiven")
		}

		if a["redis_user_config.0.public_access.0.prometheus"] != "false" {
			return fmt.Errorf("expected to get a correct public_access.prometheus from Aiven")
		}

		return nil
	}
}
