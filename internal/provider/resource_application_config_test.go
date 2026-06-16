//go:build acc
// +build acc

package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceApplicationConfig_basic(t *testing.T) {
	resourceName := "pocketid_application_config.test"
	appName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceApplicationConfigConfig_basic(appName, "60"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "app_name", appName),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "60"),
					resource.TestCheckResourceAttr(resourceName, "id", "application-configuration"),
					// Computed defaults should be populated by the server.
					resource.TestCheckResourceAttrSet(resourceName, "allow_user_signups"),
				),
			},
			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccResourceApplicationConfigConfig_basic(appName+"-updated", "120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "app_name", appName+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "120"),
				),
			},
		},
	})
}

func TestAccResourceApplicationConfig_dataSource(t *testing.T) {
	appName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceApplicationConfigConfig_withDataSource(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pocketid_application_config.test", "app_name", appName),
					resource.TestCheckResourceAttr("data.pocketid_application_config.test", "app_name", appName),
					resource.TestCheckResourceAttr("data.pocketid_application_config.test", "id", "application-configuration"),
				),
			},
		},
	})
}

func testAccResourceApplicationConfigConfig_basic(appName, sessionDuration string) string {
	return fmt.Sprintf(`
resource "pocketid_application_config" "test" {
  app_name         = %[1]q
  session_duration = %[2]q
}
`, appName, sessionDuration)
}

func testAccResourceApplicationConfigConfig_withDataSource(appName string) string {
	return fmt.Sprintf(`
resource "pocketid_application_config" "test" {
  app_name = %[1]q
}

data "pocketid_application_config" "test" {
  depends_on = [pocketid_application_config.test]
}
`, appName)
}
