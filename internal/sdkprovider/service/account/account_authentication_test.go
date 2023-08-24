package account_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func TestAccAivenAccountAuthentication_basic(t *testing.T) {
	resourceName := "aiven_account_authentication.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenAccountAuthenticationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAuthenticationResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenAccountAuthenticationAttributes("data.aiven_account_authentication.auth"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-auth-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", "saml"),
				),
			},
		},
	})
}

func testAccAccountAuthenticationResourceSAMLCertificate(rName, certificate string) string {
	return fmt.Sprintf(`
resource "aiven_account" "user" {
  name = "test-acc-account-%[1]s"
}

resource "aiven_account_authentication" "method" {
  account_id       = aiven_account.user.account_id
  type             = "saml"
  name             = "test-acc-auth-method-%[1]s"
  saml_certificate = <<-EOT
  %[2]s
  EOT

  saml_field_mapping {
    email      = "test@aiven.io"
    first_name = "TestName"
    identity   = "1234567"
    last_name  = "TestLastName"
    real_name  = "TestRealName"
  }
}
`, rName, certificate)
}

func TestAccAivenAccountAuthentication_saml_valid_certificate_create_update(t *testing.T) {
	certCreate, err := genX509Certificate(time.Now())
	assert.NoError(t, err)

	certUpdate, err := genX509Certificate(time.Now())
	assert.NoError(t, err)

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := "aiven_account_authentication.method"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenAccountAuthenticationResourceDestroy,
		Steps: []resource.TestStep{
			// Creates
			{
				Config: testAccAccountAuthenticationResourceSAMLCertificate(rName, certCreate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenAccountAuthenticationAttributes(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-acc-auth-method-"+rName),
					resource.TestCheckResourceAttr(resourceName, "saml_certificate", certCreate),
				),
			},
			// Updates
			{
				Config: testAccAccountAuthenticationResourceSAMLCertificate(rName, certUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenAccountAuthenticationAttributes(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-acc-auth-method-"+rName),
					resource.TestCheckResourceAttr(resourceName, "saml_certificate", certUpdate),
				),
			},
		},
	})
}

func TestAccAivenAccountAuthentication_saml_invalid_certificate(t *testing.T) {
	certCreate, err := genX509Certificate(time.Now().AddDate(-30, 0, 0))
	assert.NoError(t, err)

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenAccountAuthenticationResourceDestroy,
		Steps: []resource.TestStep{
			// Creates
			{
				Config: testAccAccountAuthenticationResourceSAMLCertificate(rName, certCreate),
			},
		},
		ErrorCheck: func(err error) error {
			assert.ErrorContains(t, err, "Certificate is no longer valid")
			return nil
		},
	})
}

func TestAccAivenAccountAuthentication_auto_join_team_id(t *testing.T) {
	resourceName := "aiven_account_authentication.foo"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenAccountAuthenticationWithAutoJoinTeamIDResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAuthenticationWithAutoJoinTeamIDResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenAccountAuthenticationAttributes("data.aiven_account_authentication.auth"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-acc-auth-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", "saml"),
					resource.TestCheckResourceAttrPair(resourceName, "auto_join_team_id", "aiven_account_team.foo", "team_id"),
				),
			},
		},
	})
}

func testAccAccountAuthenticationResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_account" "foo" {
  name = "test-acc-ac-%s"
}

resource "aiven_account_authentication" "foo" {
  account_id = aiven_account.foo.account_id
  name       = "test-acc-auth-%s"
  type       = "saml"
  enabled    = false
}

data "aiven_account_authentication" "auth" {
  account_id = aiven_account_authentication.foo.account_id
  name       = aiven_account_authentication.foo.name

  depends_on = [aiven_account_authentication.foo]
}`, name, name)
}

func testAccAccountAuthenticationWithAutoJoinTeamIDResource(name string) string {
	return fmt.Sprintf(`
resource "aiven_account" "foo" {
  name = "test-acc-ac-%s"
}

resource "aiven_account_team" "foo" {
  account_id = aiven_account.foo.account_id
  name       = "test-acc-team-%s"
}

resource "aiven_account_authentication" "foo" {
  account_id        = aiven_account.foo.account_id
  name              = "test-acc-auth-%s"
  type              = "saml"
  enabled           = false
  auto_join_team_id = aiven_account_team.foo.team_id
}

data "aiven_account_team" "team" {
  name       = aiven_account_team.foo.name
  account_id = aiven_account_team.foo.account_id

  depends_on = [aiven_account_team.foo]
}

data "aiven_account_authentication" "auth" {
  account_id = aiven_account_authentication.foo.account_id
  name       = aiven_account_authentication.foo.name

  depends_on = [aiven_account_authentication.foo]
}`, name, name, name)
}

func testAccCheckAivenAccountAuthenticationResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	// loop through the resources in state, verifying each account authentication is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_account_authentication" {
			continue
		}

		accountID, authID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		r, err := c.Accounts.List()
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}

			return nil
		}

		for _, ac := range r.Accounts {
			if ac.Id == accountID {
				ra, err := c.AccountAuthentications.List(accountID)
				if err != nil {
					if err.(aiven.Error).Status != 404 {
						return err
					}

					return nil
				}

				for _, a := range ra.AuthenticationMethods {
					if a.AuthenticationMethodID == authID {
						return fmt.Errorf("account authentication (%s) still exists", rs.Primary.ID)
					}
				}
			}
		}
	}

	return nil
}

func testAccCheckAivenAccountAuthenticationWithAutoJoinTeamIDResourceDestroy(s *terraform.State) error {
	c := acc.GetTestAivenClient()

	// loop through the resources in state, verifying each account authentication is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_account_team" && rs.Type != "aiven_account_authentication" {
			continue
		}

		isTeam := rs.Type == "aiven_account_team"

		accountID, secondaryID, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		r, err := c.Accounts.List()
		if err != nil {
			if err.(aiven.Error).Status != 404 {
				return err
			}

			return nil
		}

		for _, ac := range r.Accounts {
			if ac.Id == accountID {
				if isTeam {
					rl, err := c.AccountTeams.List(accountID)
					if err != nil {
						if err.(aiven.Error).Status != 404 {
							return err
						}

						return nil
					}

					for _, team := range rl.Teams {
						if team.Id == secondaryID {
							return fmt.Errorf("account team (%s) still exists", rs.Primary.ID)
						}
					}
				} else {
					ra, err := c.AccountAuthentications.List(accountID)
					if err != nil {
						if err.(aiven.Error).Status != 404 {
							return err
						}

						return nil
					}

					for _, a := range ra.AuthenticationMethods {
						if a.AuthenticationMethodID == secondaryID {
							return fmt.Errorf("account authentication (%s) still exists", rs.Primary.ID)
						}
					}
				}
			}
		}
	}

	return nil
}

func testAccCheckAivenAccountAuthenticationAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		if r.Primary == nil {
			return fmt.Errorf("resource %s not found", n)
		}

		a := r.Primary.Attributes
		log.Printf("[DEBUG] account team attributes %v", a)

		if a["enabled"] != "false" {
			return fmt.Errorf("expected to get an enabled from Aiven")
		}

		if a["type"] != "saml" {
			return fmt.Errorf("expected to get a correty type from Aiven")
		}

		if a["account_id"] == "" {
			return fmt.Errorf("expected to get an account id from Aiven")
		}

		if a["authentication_id"] == "" {
			return fmt.Errorf("expected to get an authentication_id from Aiven")
		}

		if a["name"] == "" {
			return fmt.Errorf("expected to get a name from Aiven")
		}

		if a["create_time"] == "" {
			return fmt.Errorf("expected to get a create_time from Aiven")
		}

		if a["saml_acs_url"] == "" {
			return fmt.Errorf("expected to get a saml_acs_url from Aiven")
		}

		if a["saml_metadata_url"] == "" {
			return fmt.Errorf("expected to get a saml_metadata_url from Aiven")
		}

		return nil
	}
}

func genX509Certificate(now time.Time) (string, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2022),
		NotBefore:    now,
		NotAfter:     now.Add(time.Hour),
		Subject: pkix.Name{
			Organization: []string{"aiven"},
		},
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &key.PublicKey, key)
	if err != nil {
		return "", err
	}

	p := new(bytes.Buffer)
	err = pem.Encode(p, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	if err != nil {
		return "", err
	}
	return strings.TrimSpace(p.String()), nil
}
