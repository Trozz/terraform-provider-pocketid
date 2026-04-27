package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// helper to run the validator against a string value
func runValidate(t *testing.T, v urlValidator, input string) (hasError bool) {
	t.Helper()
	var req validator.StringRequest
	var resp validator.StringResponse

	req.Path = path.Root("callback_urls")
	req.ConfigValue = types.StringValue(input)

	v.ValidateString(context.Background(), req, &resp)

	return resp.Diagnostics.HasError()
}

func TestURLValidator_ValidURLs(t *testing.T) {
	v := urlValidator{}

	cases := []string{
		// standard URLs
		"https://example.com/callback",
		"http://localhost:8080/",
		"https://example.com:8443/path?query=1",
		"https://user:pass@example.com/",

		// custom URL scheme (mobile deep link)
		"app.immich:///oauth-callback",
		"myapp://auth/redirect",
		"com.example.app://callback",
		"ftp://even.this.scheme/passes",
		"javascript:alert('xss')",

		// wildcard URLs
		"https://*.example.com/*",
		"http://localhost:*",
		"myapp://*",
	}

	for _, c := range cases {
		if runValidate(t, v, c) {
			t.Fatalf("expected valid URL, got error for %q", c)
		}
	}
}

func TestURLValidator_InvalidURLs(t *testing.T) {
	v := urlValidator{}

	cases := []string{
		"not-a-url",
		"example.com/no-scheme",
		"http://", // missing host
		"://missing-scheme",
		"",    // empty string
		"   ", // whitespace only
	}

	for _, c := range cases {
		if !runValidate(t, v, c) {
			t.Fatalf("expected invalid URL, got no error for %q", c)
		}
	}
}

func TestURLValidator_TrimsWhitespace(t *testing.T) {
	v := urlValidator{}

	input := "  https://example.com/path  "
	if runValidate(t, v, input) {
		t.Fatalf("expected trimmed URL to be valid, got error for %q", input)
	}
}
