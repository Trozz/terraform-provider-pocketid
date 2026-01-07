package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/Trozz/terraform-provider-pocketid/internal/resources"
)

func TestLDAPURLValidator_Valid(t *testing.T) {
	testCases := []string{
		"ldap://localhost:389",
		"ldaps://ldap.example.com:636",
		"ldap://192.168.1.1:389",
		"ldaps://ldap.example.com",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("url"),
				ConfigValue: types.StringValue(tc),
			}
			resp := &validator.StringResponse{
				Diagnostics: diag.Diagnostics{},
			}

			resources.LDAPURLValidator{}.ValidateString(context.Background(), req, resp)

			assert.False(t, resp.Diagnostics.HasError(), "expected no error for %s", tc)
		})
	}
}

func TestLDAPURLValidator_Invalid(t *testing.T) {
	testCases := []string{
		"http://localhost:389",
		"https://ldap.example.com",
		"ftp://ldap.example.com",
		"not-a-url",
		"",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("url"),
				ConfigValue: types.StringValue(tc),
			}
			resp := &validator.StringResponse{
				Diagnostics: diag.Diagnostics{},
			}

			resources.LDAPURLValidator{}.ValidateString(context.Background(), req, resp)

			assert.True(t, resp.Diagnostics.HasError(), "expected error for %s", tc)
		})
	}
}

func TestLDAPURLValidator_NullValue(t *testing.T) {
	req := validator.StringRequest{
		Path:        path.Root("url"),
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{
		Diagnostics: diag.Diagnostics{},
	}

	resources.LDAPURLValidator{}.ValidateString(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "null values should pass validation")
}

func TestDNValidator_Valid(t *testing.T) {
	testCases := []string{
		"cn=admin,dc=example,dc=com",
		"dc=example,dc=com",
		"ou=users,dc=example,dc=com",
		"CN=Admin,OU=Users,DC=example,DC=com",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("bind_dn"),
				ConfigValue: types.StringValue(tc),
			}
			resp := &validator.StringResponse{
				Diagnostics: diag.Diagnostics{},
			}

			resources.DNValidator{}.ValidateString(context.Background(), req, resp)

			assert.False(t, resp.Diagnostics.HasError(), "expected no error for %s", tc)
		})
	}
}

func TestDNValidator_Invalid(t *testing.T) {
	testCases := []string{
		"not-a-dn",
		"admin",
		"",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("bind_dn"),
				ConfigValue: types.StringValue(tc),
			}
			resp := &validator.StringResponse{
				Diagnostics: diag.Diagnostics{},
			}

			resources.DNValidator{}.ValidateString(context.Background(), req, resp)

			assert.True(t, resp.Diagnostics.HasError(), "expected error for %s", tc)
		})
	}
}

func TestDNValidator_NullValue(t *testing.T) {
	req := validator.StringRequest{
		Path:        path.Root("bind_dn"),
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{
		Diagnostics: diag.Diagnostics{},
	}

	resources.DNValidator{}.ValidateString(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "null values should pass validation")
}
