package kafka_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func TestAccAiven_kafka(t *testing.T) {
	resourceName := "aiven_kafka.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rName2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccKafkaDoubleTagResource(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ExpectError:        regexp.MustCompile("tag keys should be unique"),
			},
			{
				Config: testAccKafkaResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_kafka.common"),
					testAccCheckAivenServiceKafkaAttributes("data.aiven_kafka.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "kafka"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_acl", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "service_username"),
					resource.TestCheckResourceAttrSet(resourceName, "service_password"),
					resource.TestCheckResourceAttrSet(resourceName, "service_host"),
					resource.TestCheckResourceAttrSet(resourceName, "service_port"),
					resource.TestCheckResourceAttrSet(resourceName, "service_uri"),
				),
			},
			{
				Config: testAccKafkaWithoutDefaultACLResource(rName2),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_kafka.common"),
					testAccCheckAivenServiceKafkaAttributes("data.aiven_kafka.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName2)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "kafka"),
					resource.TestCheckResourceAttr(resourceName, "cloud_name", "google-europe-west1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_dow", "monday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_time", "10:00:00"),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_acl", "false"),
					resource.TestCheckResourceAttr(resourceName, "components.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "components.0.ssl", "true"),
					resource.TestCheckResourceAttr(resourceName, "components.0.kafka_authentication_method", "certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "service_username"),
					resource.TestCheckResourceAttrSet(resourceName, "service_password"),
					resource.TestCheckResourceAttrSet(resourceName, "service_host"),
					resource.TestCheckResourceAttrSet(resourceName, "service_port"),
					resource.TestCheckResourceAttrSet(resourceName, "service_uri"),
					func(state *terraform.State) error {
						c := acc.GetTestAivenClient()
						a, err := c.KafkaACLs.List(os.Getenv("AIVEN_PROJECT_NAME"), rName2)
						if err != nil && !aiven.IsNotFound(err) {
							return fmt.Errorf("cannot get a list of kafka ACLs: %s", err)
						}

						if len(a) > 0 {
							return fmt.Errorf("list of ACLs should be empty")
						}

						s, err := c.KafkaSchemaRegistryACLs.List(os.Getenv("AIVEN_PROJECT_NAME"), rName2)
						if err != nil && !aiven.IsNotFound(err) {
							return fmt.Errorf("cannot get a list of Kafka Schema ACLs: %s", err)
						}

						if len(s) > 0 {
							return fmt.Errorf("list of Kafka Schema ACLs should be empty")
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccKafkaResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  default_acl             = false

  kafka_user_config {
    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }

    public_access {
      kafka_rest    = true
      kafka_connect = true
    }
  }
}

data "aiven_kafka" "common" {
  service_name = aiven_kafka.bar.service_name
  project      = aiven_kafka.bar.project

  depends_on = [aiven_kafka.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccKafkaWithoutDefaultACLResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  default_acl             = false

  tag {
    key   = "test"
    value = "val"
  }

  kafka_user_config {
    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }

    public_access {
      kafka_rest    = true
      kafka_connect = true
    }
  }
}
data "aiven_kafka" "common" {
  service_name = aiven_kafka.bar.service_name
  project      = aiven_kafka.bar.project

  depends_on = [aiven_kafka.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccKafkaDoubleTagResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  default_acl             = false

  tag {
    key   = "test"
    value = "val"
  }
  tag {
    key   = "test"
    value = "val2"
  }

  kafka_user_config {
    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }

    public_access {
      kafka_rest    = true
      kafka_connect = true
    }
  }
}

data "aiven_kafka" "common" {
  service_name = aiven_kafka.bar.service_name
  project      = aiven_kafka.bar.project

  depends_on = [aiven_kafka.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

// Kafka service tests
func TestAccAivenService_kafka(t *testing.T) {
	resourceName := "aiven_kafka.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaServiceResource(rName),
				Check: resource.ComposeTestCheckFunc(
					acc.TestAccCheckAivenServiceCommonAttributes("data.aiven_kafka.common"),
					testAccCheckAivenServiceKafkaAttributes("data.aiven_kafka.common"),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_type", "kafka"),
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
		},
	})
}

func testAccKafkaServiceResource(name string) string {
	return fmt.Sprintf(`
data "aiven_project" "foo" {
  project = "%s"
}

resource "aiven_kafka" "bar" {
  project                 = data.aiven_project.foo.project
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  service_name            = "test-acc-sr-%s"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  kafka_user_config {
    kafka_rest      = true
    kafka_connect   = true
    schema_registry = true

    kafka {
      group_max_session_timeout_ms = 70000
      log_retention_bytes          = 1000000000
    }

    public_access {
      kafka_rest    = true
      kafka_connect = true
    }
  }
}

data "aiven_kafka" "common" {
  service_name = aiven_kafka.bar.service_name
  project      = aiven_kafka.bar.project

  depends_on = [aiven_kafka.bar]
}`, os.Getenv("AIVEN_PROJECT_NAME"), name)
}

func testAccCheckAivenServiceKafkaAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["kafka_user_config.0.public_access.0.kafka_connect"] != "true" {
			return fmt.Errorf("expected to get a correct public_access.kafka_connect from Aiven")
		}

		if a["kafka_user_config.0.public_access.0.kafka_rest"] != "true" {
			return fmt.Errorf("expected to get a correct public_access.kafka_rest from Aiven")
		}

		if a["kafka_user_config.0.public_access.0.kafka"] != "false" {
			return fmt.Errorf("expected to get a correct public_access.kafka from Aiven")
		}

		if a["kafka_user_config.0.public_access.0.prometheus"] != "false" {
			return fmt.Errorf("expected to get a correct public_access.prometheus from Aiven")
		}

		if a["kafka.0.rest_uri"] == "" {
			return fmt.Errorf("expected to get a rest_uri from Aiven")
		}

		if a["kafka.0.schema_registry_uri"] == "" {
			return fmt.Errorf("expected to get a schema_registry_uri from Aiven")
		}

		if a["kafka.0.access_key"] == "" {
			return fmt.Errorf("expected to get an access_key from Aiven")
		}

		if a["kafka.0.access_cert"] == "" {
			return fmt.Errorf("expected to get an access_cert from Aiven")
		}

		return nil
	}
}

func testAccKafkaResourceUserConfigKafkaNullFieldsOnly(project, prefix string) string {
	return fmt.Sprintf(`
resource "aiven_kafka" "kafka" {
  project                 = "%s"
  cloud_name              = "google-europe-west1"
  plan                    = "startup-2"
  service_name            = "%s-kafka"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  kafka_user_config {
    kafka {
      group_max_session_timeout_ms = null
      log_retention_bytes          = null
    }
    public_access {
      kafka_rest    = true
      kafka_connect = true
    }
  }
}
`, project, prefix)
}

func TestAccAiven_kafka_userconfig_kafka_null_fields_only(t *testing.T) {
	project := os.Getenv("AIVEN_PROJECT_NAME")
	prefix := "test-tf-acc-" + acctest.RandString(7)
	resourceName := "aiven_kafka.kafka"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             acc.TestAccCheckAivenServiceResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaResourceUserConfigKafkaNullFieldsOnly(project, prefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "RUNNING"),
					resource.TestCheckResourceAttr(resourceName, "kafka_user_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_user_config.0.kafka.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_user_config.0.kafka.0.group_max_session_timeout_ms", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka_user_config.0.kafka.0.log_retention_bytes", "0"),
				),
			},
		},
	})
}
