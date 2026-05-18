package userlist_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
)

func testAccAivenOrganizationUserListByName(name string) string {
	return fmt.Sprintf(`
data "aiven_organization_user_list" "org" {
  name = "%s"
}
`, name)
}

func TestAccAivenOrganizationUserListByName(t *testing.T) {
	resourceName := "data.aiven_organization_user_list.org"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAivenOrganizationUserListByName(acc.OrganizationName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "users.#"),
					resource.TestMatchResourceAttr(resourceName, "users.0.user_info.0.user_email", regexp.MustCompile(`.*@.*`)),
				),
			},
		},
	})
}

func testAccAivenOrganizationUserListByID(id string) string {
	return fmt.Sprintf(`
data "aiven_organization_user_list" "org" {
  id = "%s"
}
`, id)
}

func TestAccAivenOrganizationUserListByID(t *testing.T) {
	// Skip test if TF_ACC is not set
	acc.TestAccPreCheck(t)

	// This test creates Aiven client before running PreCheck part
	// Runs checks manually
	resourceName := "data.aiven_organization_user_list.org"
	client, err := acc.GetTestGenAivenClient()
	require.NoError(t, err)

	id, err := getOrganizationByName(
		context.Background(),
		client,
		acc.OrganizationName(),
	)
	require.NoError(t, err)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAivenOrganizationUserListByID(id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "users.#"),
					resource.TestMatchResourceAttr(resourceName, "users.0.user_info.0.user_email", regexp.MustCompile(`.*@.*`)),
				),
			},
		},
	})
}

func TestAccAivenOrganizationUserList_InvalidInput(t *testing.T) {
	t.Run("both id and name set", func(t *testing.T) {
		tfConfig := fmt.Sprintf(`
data "aiven_organization_user_list" "invalid" {
  id   = "%s"
  name = "%s"
}
`, "dummy-id", "dummy-name")

		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      tfConfig,
					ExpectError: regexp.MustCompile(`Exactly one of these attributes must be configured: \[id,name]`),
				},
			},
		})
	})

	t.Run("neither id nor name set", func(t *testing.T) {
		tfConfig := `
data "aiven_organization_user_list" "invalid" {
  # neither id nor name
}
`
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      tfConfig,
					ExpectError: regexp.MustCompile(`Exactly one of these attributes must be configured: \[id,name]`),
				},
			},
		})
	})
}

// GetOrganizationByName resolves an organization name to its ID. It is kept as an exported
// helper for tests; the runtime lookup is handled by the generated readView's id-empty branch.
func getOrganizationByName(ctx context.Context, client avngen.Client, name string) (string, error) {
	ids := make([]string, 0)
	list, err := client.UserOrganizationsList(ctx)
	if err != nil {
		return "", err
	}

	for _, o := range list {
		// Organization name is not unique
		if o.OrganizationName == name {
			ids = append(ids, o.OrganizationId)
		}
	}

	switch len(ids) {
	case 0:
		return "", fmt.Errorf("organization %q not found", name)
	case 1:
		return ids[0], nil
	}
	return "", fmt.Errorf("multiple organizations %q found, ids: %s", name, strings.Join(ids, ", "))
}
