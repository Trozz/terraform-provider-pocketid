//go:build acc
// +build acc

package provider_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/Trozz/terraform-provider-pocketid/internal/provider"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"pocketid": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// Check for required environment variables
	if v := os.Getenv("POCKETID_BASE_URL"); v == "" {
		t.Fatal("POCKETID_BASE_URL must be set for acceptance tests")
	}

	if v := os.Getenv("POCKETID_API_TOKEN"); v == "" {
		t.Fatal("POCKETID_API_TOKEN must be set for acceptance tests")
	}
}

// TestAccProviderConfig returns a provider configuration string for acceptance tests
func testAccProviderConfig() string {
	return `
provider "pocketid" {
  # Configuration is provided via environment variables:
  # POCKETID_BASE_URL
  # POCKETID_API_TOKEN
}
`
}

// TestMain is the entry point for acceptance tests
func TestMain(m *testing.M) {
	resource.TestMain(m)
}
