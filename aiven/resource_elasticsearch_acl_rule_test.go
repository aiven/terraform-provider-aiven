// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAivenElasticsearchACLRule_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resourceName := "aiven_elasticsearch_acl_rule.foo"
	serviceName := fmt.Sprintf("test-acc-sr-aclrule-%s", rName)
	userName := fmt.Sprintf("user-%s", rName)
	indexName := "test-index"
	permissionName := "readwrite"
	projectName := os.Getenv("AIVEN_PROJECT_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAivenElasticsearchACLRuleResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchACLRuleResource(projectName, serviceName, userName, indexName, permissionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", projectName),
					resource.TestCheckResourceAttr(resourceName, "service_name", serviceName),
					resource.TestCheckResourceAttr(resourceName, "index", indexName),
					resource.TestCheckResourceAttr(resourceName, "username", userName),
					resource.TestCheckResourceAttr(resourceName, "permission", permissionName),
				),
			},
			{
				Config:        testAccElasticsearchACLRuleResourceImport(projectName, serviceName, userName, indexName, permissionName),
				ResourceName:  resourceName,
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/%s/%s/%s", projectName, serviceName, userName, indexName),
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if len(s) != 1 {
						return fmt.Errorf("expected only one instance to be imported, state: %#v", s)
					}
					rs := s[0]
					attributes := rs.Attributes
					if attServiceName := attributes["service_name"]; !strings.EqualFold(attServiceName, serviceName) {
						return fmt.Errorf("expected service_name to match '%s', got: '%s'", serviceName, attServiceName)
					}
					if attProjectName := attributes["project"]; !strings.EqualFold(attProjectName, projectName) {
						return fmt.Errorf("expected project to match '%s', got: '%s'", projectName, attProjectName)
					}
					if attUserName := attributes["username"]; !strings.EqualFold(attUserName, userName) {
						return fmt.Errorf("expected username '%s' after import, got :%s", userName, attUserName)
					}
					if attIndexName := attributes["index"]; !strings.EqualFold(attIndexName, indexName) {
						return fmt.Errorf("expected index '%s' after import, got :%s", indexName, attIndexName)
					}
					if attPermissionName := attributes["permission"]; !strings.EqualFold(attPermissionName, permissionName) {
						return fmt.Errorf("expected permission '%s' after import, got :%s", permissionName, attPermissionName)
					}
					return nil
				},
			},
		},
	})
}

func testAccElasticsearchACLRuleResource(project, service, user, index, permission string) string {
	return fmt.Sprintf(`
    data "aiven_project" "foo" {
      project = "%s"
    }

    resource "aiven_elasticsearch" "bar" {
      project = data.aiven_project.foo.project
      cloud_name = "google-europe-west1"
      plan = "startup-4"
      service_name = "%s"
      maintenance_window_dow = "monday"
      maintenance_window_time = "10:00:00"
    }

    resource "aiven_service_user" "foo" {
      service_name = aiven_elasticsearch.bar.service_name
      project = data.aiven_project.foo.project
      username = "%s"
    }

    resource "aiven_elasticsearch_acl_config" "foo" {
      project = data.aiven_project.foo.project
      service_name = aiven_elasticsearch.bar.service_name
      enabled = true
      extended_acl = false
    }

    resource "aiven_elasticsearch_acl_rule" "foo" {
      project = data.aiven_project.foo.project
      service_name = aiven_elasticsearch.bar.service_name
      username = aiven_service_user.foo.username
      index = "%s"
      permission = "%s"
    }
    `, project, service, user, index, permission)
}

func testAccElasticsearchACLRuleResourceImport(project, service, user, index, permission string) string {
	return fmt.Sprintf(`
    resource "aiven_elasticsearch_acl_rule" "foo" {
      project = "%s"
      service_name = "%s"
      username = "%s"
      index = "%s"
      permission = "%s"
    }
    `, project, service, user, index, permission)
}

func testAccCheckAivenElasticsearchACLRuleResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each OS ACL is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_opensearch_acl_rule" {
			continue
		}

		projectName, serviceName, username, index := splitResourceID4(rs.Primary.ID)

		r, err := c.ElasticsearchACLs.Get(projectName, serviceName)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}
		if r == nil {
			return nil
		}

		for _, acl := range r.ElasticSearchACLConfig.ACLs {
			if acl.Username != username {
				continue
			}
			for _, rule := range acl.Rules {
				if rule.Index == index {
					return fmt.Errorf("opensearch acl (%s) still exists", rs.Primary.ID)
				}
			}
		}
	}
	return nil
}
