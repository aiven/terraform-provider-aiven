// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"os"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aiven_kafka_connector", &resource.Sweeper{
		Name: "aiven_kafka_connector",
		F:    sweepKafkaConnectors,
	})
}

func sweepKafkaConnectors(region string) error {
	client, err := sharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*aiven.Client)

	projects, err := conn.Projects.List()
	if err != nil {
		return fmt.Errorf("error retrieving a list of projects : %s", err)
	}

	for _, project := range projects {
		if project.Name == os.Getenv("AIVEN_PROJECT_NAME") {
			services, err := conn.Services.List(project.Name)
			if err != nil {
				return fmt.Errorf("error retrieving a list of services for a project `%s`: %s", project.Name, err)
			}

			for _, service := range services {
				if service.Type != "kafka" {
					continue
				}

				connectorsList, err := conn.KafkaConnectors.List(project.Name, service.Name)
				if err != nil {
					if err.(aiven.Error).Status == 403 || err.(aiven.Error).Status == 501 {
						continue
					}

					return fmt.Errorf("error retrieving a list of kafka connectors for a service `%s`: %s", service.Name, err)
				}

				for _, c := range connectorsList.Connectors {
					err = conn.KafkaConnectors.Delete(project.Name, service.Name, c.Name)
					if err != nil {
						return fmt.Errorf("error destroying kafka connector `%s` during sweep: %s", c.Name, err)
					}
				}
			}
		}
	}

	return nil
}

func TestAccAivenKafkaConnector_basic(t *testing.T) {
	resourceName := "aiven_kafka_connector.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenKafkaConnectorResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaConnectorResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenKafkaConnectorAttributes("data.aiven_kafka_connector.connector"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "connector_name", fmt.Sprintf("test-acc-con-%s", rName)),
				),
			},
		},
	})
}

func TestAccAivenKafkaConnector_mogosink(t *testing.T) {
	if os.Getenv("MONGO_URI") == "" {
		t.Skip("MONGO_URI environment variable is required to run this test")
	}

	resourceName := "aiven_kafka_connector.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenKafkaConnectorResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaConnectorMonoSinkResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "connector_name", fmt.Sprintf("test-acc-con-mongo-sink-%s", rName)),
				),
			},
		},
	})
}

func testAccCheckAivenKafkaConnectorResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each aiven_kafka_connector is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_kafka" {
			continue
		}

		projectName, serviceName := schemautil.SplitResourceID2(rs.Primary.ID)
		_, err := c.Services.Get(projectName, serviceName)
		if err != nil {
			if err.(aiven.Error).Status == 404 {
				return nil
			}

			return err
		}

		list, err := c.KafkaConnectors.List(projectName, serviceName)
		if err != nil {
			if err.(aiven.Error).Status == 404 {
				return nil
			}

			return err
		}

		for _, connector := range list.Connectors {
			res, err := c.KafkaConnectors.GetByName(projectName, serviceName, connector.Name)
			if err != nil {
				if err.(aiven.Error).Status == 404 {
					return nil
				}

				return err
			}

			if res != nil {
				return fmt.Errorf("kafka connector (%s) still exists", connector.Name)
			}
		}

	}

	return nil
}

// nosemgrep: kafka connectors need kafka with business plans
func testAccKafkaConnectorResource(name string) string {
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
		    kafka_connect = true
		
		    kafka {
		      group_max_session_timeout_ms = 70000
		      log_retention_bytes          = 1000000000
		    }
		  }
		}
		
		resource "aiven_kafka_topic" "foo" {
		  project      = data.aiven_project.foo.project
		  service_name = aiven_kafka.bar.service_name
		  topic_name   = "test-acc-topic-%s"
		  partitions   = 3
		  replication  = 2
		}
		
		resource "aiven_elasticsearch" "dest" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "startup-4"
		  service_name            = "test-acc-sr2-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		}
		
		resource "aiven_kafka_connector" "foo" {
		  project        = data.aiven_project.foo.project
		  service_name   = aiven_kafka.bar.service_name
		  connector_name = "test-acc-con-%s"
		
		  config = {
		    "topics" = aiven_kafka_topic.foo.topic_name
		    "connector.class" : "io.aiven.connect.elasticsearch.ElasticsearchSinkConnector"
		    "type.name"      = "es-connector"
		    "name"           = "test-acc-con-%s"
		    "connection.url" = aiven_elasticsearch.dest.service_uri
		  }
		}
		
		data "aiven_kafka_connector" "connector" {
		  project        = aiven_kafka_connector.foo.project
		  service_name   = aiven_kafka_connector.foo.service_name
		  connector_name = aiven_kafka_connector.foo.connector_name
		
		  depends_on = [aiven_kafka_connector.foo]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, name, name, name, name)
}

func testAccKafkaConnectorMonoSinkResource(name string) string {
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
		    kafka_connect   = true
		    schema_registry = true
		
		    kafka {
		      group_max_session_timeout_ms = 70000
		      log_retention_bytes          = 1000000000
		    }
		  }
		}
		
		resource "aiven_kafka_topic" "foo" {
		  project      = data.aiven_project.foo.project
		  service_name = aiven_kafka.bar.service_name
		  topic_name   = "test-acc-topic-%s"
		  partitions   = 3
		  replication  = 2
		}
		
		resource "aiven_kafka_connector" "foo" {
		  project        = data.aiven_project.foo.project
		  service_name   = aiven_kafka.bar.service_name
		  connector_name = "test-acc-con-mongo-sink-%s"
		
		  config = {
		    "name" = "test-acc-con-mongo-sink-%s"
		    "connector.class" : "com.mongodb.kafka.connect.MongoSinkConnector"
		    "topics"    = aiven_kafka_topic.foo.topic_name
		    "tasks.max" = 1
		
		    # mongo connect settings
		    "connection.uri" = "%s"
		    "database"       = "acc-test-mongo"
		    "collection"     = "mongo_collection_name"
		    "max.batch.size" = 1
		  }
		}
		
		data "aiven_kafka_connector" "connector" {
		  project        = aiven_kafka_connector.foo.project
		  service_name   = aiven_kafka_connector.foo.service_name
		  connector_name = aiven_kafka_connector.foo.connector_name
		
		  depends_on = [aiven_kafka_connector.foo]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, name, name, name, os.Getenv("MONGO_URI"))
}

func testAccCheckAivenKafkaConnectorAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven)")
		}

		if a["plugin_doc_url"] == "" {
			return fmt.Errorf("expected to get a plugin_doc_url from Aiven)")
		}

		if a["plugin_title"] == "" {
			return fmt.Errorf("expected to get a plugin_title from Aiven)")
		}

		if a["plugin_version"] == "" {
			return fmt.Errorf("expected to get a plugin_version from Aiven)")
		}

		if a["config.connector.class"] != "io.aiven.connect.elasticsearch.ElasticsearchSinkConnector" {
			return fmt.Errorf("expected to get a correct config.connector.class from Aiven)")
		}

		if a["config.connection.url"] == "" {
			return fmt.Errorf("expected to get a config.connection.url from Aiven)")
		}

		if a["config.topics"] == "" {
			return fmt.Errorf("expected to get a config.topics from Aiven)")
		}

		if a["config.type.name"] != "es-connector" {
			return fmt.Errorf("expected to get a corect config.type.name from Aiven)")
		}

		if a["config.name"] == "" {
			return fmt.Errorf("expected to get a config.name from Aiven)")
		}

		if a["plugin_type"] != "sink" {
			return fmt.Errorf("expected to get a correct plugin_type from Aiven)")
		}

		return nil
	}
}
