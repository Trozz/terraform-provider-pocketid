//go:build acc
// +build acc

package provider_test

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

func TestAccResourceApplicationConfig_basic(t *testing.T) {
	resourceName := "pocketid_application_config.test"
	// Pocket-ID enforces appName max length of 30 characters, so keep both the
	// create and update names short and distinct.
	suffix := acctest.RandString(8)
	appName := "tf-acc-" + suffix
	appNameUpdated := "tf-upd-" + suffix

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
				Config: testAccResourceApplicationConfigConfig_basic(appNameUpdated, "120"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "app_name", appNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "120"),
				),
			},
		},
	})
}

func TestAccResourceApplicationConfig_dataSource(t *testing.T) {
	// Pocket-ID enforces appName max length of 30 characters.
	appName := "tf-acc-" + acctest.RandString(8)

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

// applicationImages maps each image attribute to its application image.
var applicationImages = []struct {
	attribute string
	image     client.ApplicationImage
}{
	{"logo", client.ApplicationImageLogo},
	{"dark_logo", client.ApplicationImageDarkLogo},
	{"email_logo", client.ApplicationImageEmailLogo},
	{"background_image", client.ApplicationImageBackground},
	{"favicon", client.ApplicationImageFavicon},
	{"default_profile_picture", client.ApplicationImageDefaultProfilePicture},
}

// getApplicationImage returns the custom version of an application image
// Pocket ID serves, and false when none is set.
func getApplicationImage(image client.ApplicationImage) ([]byte, bool, error) {
	u, err := url.Parse(os.Getenv("POCKETID_BASE_URL") + "/api/application-images/" + string(image))
	if err != nil {
		return nil, false, err
	}
	q := u.Query()
	q.Set("default", "false")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("X-API-KEY", os.Getenv("POCKETID_API_TOKEN"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		data, err := io.ReadAll(resp.Body)
		return data, true, err
	case http.StatusNotFound:
		return nil, false, nil
	default:
		return nil, false, fmt.Errorf("GET %s: %s", u, resp.Status)
	}
}

// checkApplicationImagesSet asserts which images have a custom version.
func checkApplicationImagesSet(want map[string]bool) resource.TestCheckFunc {
	return func(*terraform.State) error {
		for _, img := range applicationImages {
			_, set, err := getApplicationImage(img.image)
			if err != nil {
				return err
			}
			if set != want[img.attribute] {
				return fmt.Errorf("%s: custom image set = %v, want %v", img.attribute, set, want[img.attribute])
			}
		}
		return nil
	}
}

// testAccResourceApplicationConfigConfig_images renders the configuration
// with the given image sources, keyed by attribute.
func testAccResourceApplicationConfigConfig_images(sources map[string]string) string {
	var images string
	for _, img := range applicationImages {
		if source, ok := sources[img.attribute]; ok {
			images += fmt.Sprintf("  %s = {\n    source = %q\n  }\n", img.attribute, source)
		}
	}
	return fmt.Sprintf(`
resource "pocketid_application_config" "test" {
%s}
`, images)
}

// TestAccResourceApplicationConfig_images walks every image through its
// lifecycle: upload, re-upload after the file changes, and removal.
func TestAccResourceApplicationConfig_images(t *testing.T) {
	resourceName := "pocketid_application_config.test"
	dir := t.TempDir()
	sources := map[string]string{}
	for i, img := range applicationImages {
		sources[img.attribute] = filepath.Join(dir, img.attribute+".png")
		writeTestPNG(t, sources[img.attribute], color.RGBA{R: uint8(40 * i), G: 100, A: 255})
	}
	allSet := map[string]bool{}
	for _, img := range applicationImages {
		allSet[img.attribute] = true
	}
	var firstLogo []byte

	withOnly := func(attributes ...string) map[string]string {
		m := map[string]string{}
		for _, a := range attributes {
			m[a] = sources[a]
		}
		return m
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceApplicationConfigConfig_images(sources),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "logo.source", sources["logo"]),
					resource.TestCheckResourceAttrSet(resourceName, "favicon.sha256"),
					checkApplicationImagesSet(allSet),
					func(*terraform.State) error {
						var err error
						firstLogo, _, err = getApplicationImage(client.ApplicationImageLogo)
						return err
					},
				),
			},
			{
				PreConfig: func() {
					writeTestPNG(t, sources["logo"], color.RGBA{B: 255, A: 255})
				},
				Config: testAccResourceApplicationConfigConfig_images(sources),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: func(*terraform.State) error {
					logo, _, err := getApplicationImage(client.ApplicationImageLogo)
					if err != nil {
						return err
					}
					if bytes.Equal(logo, firstLogo) {
						return fmt.Errorf("logo was not replaced after the file changed")
					}
					return nil
				},
			},
			{
				Config: testAccResourceApplicationConfigConfig_images(sources),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				// Pocket ID can reset these four; the other two stay set.
				Config: testAccResourceApplicationConfigConfig_images(withOnly("email_logo", "favicon")),
				Check:  checkApplicationImagesSet(map[string]bool{"email_logo": true, "favicon": true}),
			},
			{
				// Removing images Pocket ID cannot delete only stops managing them.
				Config: testAccResourceApplicationConfigConfig_images(withOnly()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "favicon.source"),
					checkApplicationImagesSet(map[string]bool{"email_logo": true, "favicon": true}),
				),
			},
		},
	})
}

// TestAccResourceApplicationConfig_imageRejected checks that a file Pocket ID
// rejects fails the apply. Pocket ID takes the image type from the extension.
func TestAccResourceApplicationConfig_imageRejected(t *testing.T) {
	notAnImage := filepath.Join(t.TempDir(), "logo.txt")
	if err := os.WriteFile(notAnImage, []byte("not an image"), 0o600); err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceApplicationConfigConfig_images(map[string]string{"logo": notAnImage}),
				ExpectError: regexp.MustCompile(`Pocket ID rejected "logo.txt"`),
			},
		},
	})
}
