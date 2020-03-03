package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"reflect"
	"testing"
)

func TestAccAivenElasticsearchAcl_basic(t *testing.T) {
	t.Parallel()

	resourceName := "aiven_elasticsearch_acl.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAivenAleasticsearchAclResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticsearchAclResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenElasticsearchAclAttributes("data.aiven_elasticsearch_acl.acl"),
					resource.TestCheckResourceAttr(resourceName, "project", fmt.Sprintf("test-acc-pr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "service_name", fmt.Sprintf("test-acc-sr-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "extended_acl", "false"),
				),
			},
		},
	})
}

func testAccElasticsearchAclResource(name string) string {
	return fmt.Sprintf(`
		resource "aiven_project" "foo" {
			project = "test-acc-pr-%s"
			card_id="%s"	
		}

		resource "aiven_service" "bar" {
			project = aiven_project.foo.project
			cloud_name = "google-europe-west1"
			plan = "startup-4"
			service_name = "test-acc-sr-%s"
			service_type = "elasticsearch"
			maintenance_window_dow = "monday"
			maintenance_window_time = "10:00:00"
			
			elasticsearch_user_config {
				elasticsearch_version = 7
			}
		}

		resource "aiven_service_user" "foo" {
			service_name = aiven_service.bar.service_name
			project = aiven_project.foo.project
			username = "user-%s"
		}
		
		resource "aiven_elasticsearch_acl" "foo" {
			project = aiven_project.foo.project
			service_name = aiven_service.bar.service_name
			enabled = true
			extended_acl = false

			acl {
				username = aiven_service_user.foo.username

				rule {
					index = "_*"
					permission = "admin"
				}
			
				rule {
					index = "*"
					permission = "admin"
				}
			}
		}
		
		data "aiven_elasticsearch_acl" "acl" {
			project = aiven_elasticsearch_acl.foo.project
			service_name = aiven_elasticsearch_acl.foo.service_name
		}
		`, name, os.Getenv("AIVEN_CARD_ID"), name, name)
}

func testAccCheckAivenElasticsearchAclAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project name from Aiven")
		}

		if a["service_name"] == "" {
			return fmt.Errorf("expected to get a service_name from Aiven")
		}

		if a["enabled"] != "true" {
			return fmt.Errorf("expected to get a enabled from Aiven")
		}

		if a["acl.#"] != "1" {
			return fmt.Errorf("expected to get a atleast one acls reccord from Aiven")
		}

		return nil
	}
}

func testAccCheckAivenAleasticsearchAclResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each ES ACL is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_elasticsearch_acl" {
			continue
		}

		projectName, serviceName := splitResourceID2(rs.Primary.ID)
		acl, err := c.ElasticsearchACLs.Get(projectName, serviceName)
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}
		}

		if acl != nil {
			return fmt.Errorf("elasticsearch acl (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func TestFlattenElasticsearchACL(t *testing.T) {
	type args struct {
		r *aiven.ElasticSearchACLResponse
	}
	tests := []struct {
		name string
		args args
		want []map[string]interface{}
	}{
		{
			"complex-response",
			args{r: &aiven.ElasticSearchACLResponse{
				APIResponse: aiven.APIResponse{},
				ElasticSearchACLConfig: aiven.ElasticSearchACLConfig{
					Enabled:     true,
					ExtendedAcl: true,
					ACLs: []aiven.ElasticSearchACL{
						{
							Username: "test-user1",
							Rules: []aiven.ElasticsearchACLRule{
								{
									Permission: "read",
									Index:      "_*",
								},
								{
									Permission: "admin",
									Index:      "_test*",
								},
							},
						},
						{
							Username: "test-user2",
							Rules: []aiven.ElasticsearchACLRule{
								{
									Permission: "admin",
									Index:      "*",
								},
							},
						},
					},
				},
			}},
			[]map[string]interface{}{
				{
					"username": "test-user1",
					"rule": []map[string]interface{}{
						{
							"permission": "read",
							"index":      "_*",
						},
						{
							"permission": "admin",
							"index":      "_test*",
						},
					},
				},
				{
					"username": "test-user2",
					"rule": []map[string]interface{}{
						{
							"permission": "admin",
							"index":      "*",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flattenElasticsearchACL(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("flattenElasticsearchACL() = %v, want %v", got, tt.want)
			}
		})
	}
}
