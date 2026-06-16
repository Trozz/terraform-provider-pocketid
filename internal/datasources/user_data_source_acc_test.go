//go:build acc
// +build acc

package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource_displayNameAndEmailVerified(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	email := fmt.Sprintf("%s@example.com", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_lookupByID(rName, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pocketid_user.test", "username", rName),
					resource.TestCheckResourceAttr("data.pocketid_user.test", "email", email),
					resource.TestCheckResourceAttr("data.pocketid_user.test", "display_name", "Test User"),
					resource.TestCheckResourceAttr("data.pocketid_user.test", "email_verified", "true"),
				),
			},
		},
	})
}

func TestAccUsersDataSource_displayNameAndEmailVerified(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	email := fmt.Sprintf("%s@example.com", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUsersDataSourceConfig_list(rName, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.pocketid_users.test", "users.#"),
				),
			},
		},
	})
}

func testAccUserDataSourceConfig_lookupByID(username, email string) string {
	return fmt.Sprintf(`
resource "pocketid_user" "test" {
  username       = %[1]q
  email          = %[2]q
  first_name     = "Test"
  last_name      = "User"
  email_verified = true
}

data "pocketid_user" "test" {
  id = pocketid_user.test.id
}
`, username, email)
}

func testAccUsersDataSourceConfig_list(username, email string) string {
	return fmt.Sprintf(`
resource "pocketid_user" "test" {
  username       = %[1]q
  email          = %[2]q
  first_name     = "Test"
  last_name      = "User"
  email_verified = true
}

data "pocketid_users" "test" {
  depends_on = [pocketid_user.test]
}
`, username, email)
}
