package resources

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// LDAPURLValidator validates that a string is a valid LDAP URL
type LDAPURLValidator struct{}

// Description returns the validator description
func (v LDAPURLValidator) Description(ctx context.Context) string {
	return "value must be a valid LDAP URL (ldap:// or ldaps://)"
}

// MarkdownDescription returns the markdown description
func (v LDAPURLValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation
func (v LDAPURLValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid LDAP URL",
			"LDAP URL cannot be empty",
		)
		return
	}

	parsedURL, err := url.Parse(value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid LDAP URL",
			fmt.Sprintf("The value %q is not a valid URL: %s", value, err),
		)
		return
	}

	if parsedURL.Scheme != "ldap" && parsedURL.Scheme != "ldaps" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid LDAP URL Scheme",
			fmt.Sprintf("The URL scheme must be 'ldap' or 'ldaps', got %q", parsedURL.Scheme),
		)
		return
	}

	if parsedURL.Host == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid LDAP URL",
			"The URL must specify a host",
		)
	}
}

// DNValidator validates that a string is a valid Distinguished Name
type DNValidator struct{}

// Description returns the validator description
func (v DNValidator) Description(ctx context.Context) string {
	return "value must be a valid Distinguished Name (DN)"
}

// MarkdownDescription returns the markdown description
func (v DNValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation
func (v DNValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Distinguished Name",
			"Distinguished Name cannot be empty",
		)
		return
	}

	// Basic DN validation - must contain at least one key=value pair
	if !strings.Contains(value, "=") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Distinguished Name",
			fmt.Sprintf("The value %q is not a valid DN. Expected format: cn=admin,dc=example,dc=com", value),
		)
	}
}
