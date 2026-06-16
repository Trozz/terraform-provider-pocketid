//go:build acc
// +build acc

package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceScimServiceProvider_basic(t *testing.T) {
	resourceName := "pocketid_scim_service_provider.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	clientName := rName + "-client"
	endpoint := "https://scim.example.com/v2"
	token := "test-bearer-token"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing, including round-trip of the token.
			{
				Config: testAccResourceScimServiceProviderConfig_basic(clientName, endpoint, token),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint),
					resource.TestCheckResourceAttr(resourceName, "token", token),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					testAccCheckScimServiceProviderExists(resourceName),
				),
			},
			// ImportState testing using the OIDC client ID.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Import is keyed on the OIDC client ID, not the SCIM config ID,
				// so the import ID must be the client_id attribute value.
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Not found: %s", resourceName)
					}
					return rs.Primary.Attributes["client_id"], nil
				},
				// The token is returned by the API on read, so it round-trips
				// and is verified against the imported state.
				// pocket-id returns created_at with nanosecond precision on
				// create but truncated to seconds on GET, so it cannot be
				// verified byte-for-byte after import.
				ImportStateVerifyIgnore: []string{"created_at"},
			},
			// Update and Read testing.
			{
				Config: testAccResourceScimServiceProviderConfig_basic(clientName, endpoint+"/updated", token),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint+"/updated"),
					resource.TestCheckResourceAttr(resourceName, "token", token),
				),
			},
		},
	})
}

func testAccCheckScimServiceProviderExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SCIM service provider ID is set")
		}

		return nil
	}
}

func testAccResourceScimServiceProviderConfig_basic(clientName, endpoint, token string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "pocketid_client" "test" {
  name          = %[1]q
  callback_urls = ["https://example.com/callback"]
}

resource "pocketid_scim_service_provider" "test" {
  client_id = pocketid_client.test.id
  endpoint  = %[2]q
  token     = %[3]q
}
`, clientName, endpoint, token)
}
