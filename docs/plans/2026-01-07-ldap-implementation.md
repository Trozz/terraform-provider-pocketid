# LDAP Resources Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement `pocketid_ldap_config` and `pocketid_ldap_sync` Terraform resources for managing LDAP configuration in Pocket ID.

**Architecture:** Singleton resource pattern for LDAP config (one per instance), triggers-based pattern for sync. Uses existing client HTTP infrastructure with new app-config endpoints. Nested blocks for attribute mappings.

**Tech Stack:** Go 1.23, Terraform Plugin Framework v1.15.0, testify for assertions

**Design Document:** `docs/plans/2026-01-07-ldap-resources-design.md`

---

## Task Dependencies

```
Task 1 (Models) ─────┬──► Task 3 (Client) ──┬──► Task 4 (Config Resource) ──┬──► Task 6 (Register) ──► Task 7 (Tests)
                     │                      │                               │                         Task 8 (Docs)
Task 2 (Validators) ─┘                      └──► Task 5 (Sync Resource) ────┘
```

**Parallel Groups:**
- **Group A** (no deps): Tasks 1, 2
- **Group B** (after Group A): Task 3
- **Group C** (after Task 3): Tasks 4, 5
- **Group D** (after Group C): Task 6
- **Group E** (after Task 6): Tasks 7, 8

---

## Task 1: Add LDAP Models

**Files:**
- Modify: `internal/client/models.go`

**Step 1: Add LDAP configuration structs**

Add at the end of `internal/client/models.go`:

```go
// AppConfigVariable represents a single configuration variable from the API
type AppConfigVariable struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// LDAPConfig represents the LDAP configuration for Pocket ID
type LDAPConfig struct {
	Enabled                bool   `json:"ldapEnabled"`
	URL                    string `json:"ldapUrl"`
	BindDN                 string `json:"ldapBindDn"`
	BindPassword           string `json:"ldapBindPassword"`
	BaseDN                 string `json:"ldapBase"`
	SkipCertVerify         bool   `json:"ldapSkipCertVerify"`
	UserSearchFilter       string `json:"ldapUserSearchFilter"`
	UserGroupSearchFilter  string `json:"ldapGroupSearchFilter"`
	UserUniqueAttribute    string `json:"ldapAttributeUserUniqueIdentifier"`
	UserUsernameAttribute  string `json:"ldapAttributeUserUsername"`
	UserEmailAttribute     string `json:"ldapAttributeUserEmail"`
	UserFirstNameAttribute string `json:"ldapAttributeUserFirstName"`
	UserLastNameAttribute  string `json:"ldapAttributeUserLastName"`
	GroupMemberAttribute   string `json:"ldapAttributeGroupMember"`
	GroupUniqueAttribute   string `json:"ldapAttributeGroupUniqueIdentifier"`
	GroupNameAttribute     string `json:"ldapAttributeGroupName"`
	AdminGroupName         string `json:"ldapAttributeAdminGroup"`
	SoftDeleteUsers        bool   `json:"ldapSoftDeleteUsers"`
}

// LDAPConfigUpdateRequest represents the request to update LDAP configuration
type LDAPConfigUpdateRequest struct {
	LdapEnabled                        bool   `json:"ldapEnabled"`
	LdapUrl                            string `json:"ldapUrl,omitempty"`
	LdapBindDn                         string `json:"ldapBindDn,omitempty"`
	LdapBindPassword                   string `json:"ldapBindPassword,omitempty"`
	LdapBase                           string `json:"ldapBase,omitempty"`
	LdapSkipCertVerify                 bool   `json:"ldapSkipCertVerify"`
	LdapUserSearchFilter               string `json:"ldapUserSearchFilter,omitempty"`
	LdapGroupSearchFilter              string `json:"ldapGroupSearchFilter,omitempty"`
	LdapAttributeUserUniqueIdentifier  string `json:"ldapAttributeUserUniqueIdentifier,omitempty"`
	LdapAttributeUserUsername          string `json:"ldapAttributeUserUsername,omitempty"`
	LdapAttributeUserEmail             string `json:"ldapAttributeUserEmail,omitempty"`
	LdapAttributeUserFirstName         string `json:"ldapAttributeUserFirstName,omitempty"`
	LdapAttributeUserLastName          string `json:"ldapAttributeUserLastName,omitempty"`
	LdapAttributeGroupMember           string `json:"ldapAttributeGroupMember,omitempty"`
	LdapAttributeGroupUniqueIdentifier string `json:"ldapAttributeGroupUniqueIdentifier,omitempty"`
	LdapAttributeGroupName             string `json:"ldapAttributeGroupName,omitempty"`
	LdapAttributeAdminGroup            string `json:"ldapAttributeAdminGroup,omitempty"`
	LdapSoftDeleteUsers                bool   `json:"ldapSoftDeleteUsers"`
}
```

**Step 2: Verify build**

Run: `go build ./...`
Expected: Build succeeds with no errors

**Step 3: Commit**

```bash
git add internal/client/models.go
git commit -m "feat(client): add LDAP configuration model structs"
```

---

## Task 2: Create Validators

**Files:**
- Create: `internal/resources/validators.go`
- Create: `internal/resources/validators_test.go`

**Step 1: Write validator tests**

Create `internal/resources/validators_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/resources/... -run TestLDAPURLValidator -v`
Expected: FAIL - undefined: resources.LDAPURLValidator

**Step 3: Create validators implementation**

Create `internal/resources/validators.go`:

```go
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
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/resources/... -run "TestLDAPURLValidator|TestDNValidator" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/resources/validators.go internal/resources/validators_test.go
git commit -m "feat(resources): add LDAP URL and DN validators"
```

---

## Task 3: Create App Config Client Methods

**Files:**
- Create: `internal/client/app_config.go`
- Create: `internal/client/app_config_test.go`

**Step 1: Write client tests**

Create `internal/client/app_config_test.go`:

```go
package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

func TestGetLDAPConfig(t *testing.T) {
	// Mock API response - array of config variables
	configResponse := []map[string]any{
		{"key": "ldapEnabled", "type": "boolean", "value": true},
		{"key": "ldapUrl", "type": "string", "value": "ldaps://ldap.example.com:636"},
		{"key": "ldapBindDn", "type": "string", "value": "cn=admin,dc=example,dc=com"},
		{"key": "ldapBase", "type": "string", "value": "dc=example,dc=com"},
		{"key": "ldapSkipCertVerify", "type": "boolean", "value": false},
		{"key": "ldapUserSearchFilter", "type": "string", "value": "(objectClass=person)"},
		{"key": "ldapGroupSearchFilter", "type": "string", "value": "(objectClass=groupOfNames)"},
		{"key": "ldapAttributeUserUniqueIdentifier", "type": "string", "value": "objectGUID"},
		{"key": "ldapAttributeUserUsername", "type": "string", "value": "sAMAccountName"},
		{"key": "ldapAttributeUserEmail", "type": "string", "value": "mail"},
		{"key": "ldapAttributeUserFirstName", "type": "string", "value": "givenName"},
		{"key": "ldapAttributeUserLastName", "type": "string", "value": "sn"},
		{"key": "ldapAttributeGroupMember", "type": "string", "value": "member"},
		{"key": "ldapAttributeGroupUniqueIdentifier", "type": "string", "value": "objectGUID"},
		{"key": "ldapAttributeGroupName", "type": "string", "value": "cn"},
		{"key": "ldapAttributeAdminGroup", "type": "string", "value": "PocketID-Admins"},
		{"key": "ldapSoftDeleteUsers", "type": "boolean", "value": true},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/application-configuration/all", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(configResponse)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	config, err := c.GetLDAPConfig()
	require.NoError(t, err)

	assert.True(t, config.Enabled)
	assert.Equal(t, "ldaps://ldap.example.com:636", config.URL)
	assert.Equal(t, "cn=admin,dc=example,dc=com", config.BindDN)
	assert.Equal(t, "dc=example,dc=com", config.BaseDN)
	assert.Equal(t, "objectGUID", config.UserUniqueAttribute)
	assert.Equal(t, "sAMAccountName", config.UserUsernameAttribute)
	assert.True(t, config.SoftDeleteUsers)
}

func TestUpdateLDAPConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/application-configuration", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)

		var req client.LDAPConfigUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.True(t, req.LdapEnabled)
		assert.Equal(t, "ldaps://ldap.example.com:636", req.LdapUrl)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	updateReq := &client.LDAPConfigUpdateRequest{
		LdapEnabled: true,
		LdapUrl:     "ldaps://ldap.example.com:636",
		LdapBindDn:  "cn=admin,dc=example,dc=com",
		LdapBase:    "dc=example,dc=com",
	}

	err = c.UpdateLDAPConfig(updateReq)
	assert.NoError(t, err)
}

func TestSyncLDAP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/application-configuration/sync-ldap", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	err = c.SyncLDAP()
	assert.NoError(t, err)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/client/... -run "TestGetLDAPConfig|TestUpdateLDAPConfig|TestSyncLDAP" -v`
Expected: FAIL - undefined methods

**Step 3: Create app_config.go implementation**

Create `internal/client/app_config.go`:

```go
package client

import (
	"encoding/json"
	"fmt"
)

// GetLDAPConfig retrieves the current LDAP configuration
func (c *Client) GetLDAPConfig() (*LDAPConfig, error) {
	body, err := c.doRequest("GET", "/api/application-configuration/all", nil)
	if err != nil {
		return nil, err
	}

	// Parse the array of config variables
	var configVars []AppConfigVariable
	if err := json.Unmarshal(body, &configVars); err != nil {
		return nil, fmt.Errorf("error unmarshaling config response: %w", err)
	}

	// Map config variables to LDAPConfig struct
	config := &LDAPConfig{}
	for _, v := range configVars {
		switch v.Key {
		case "ldapEnabled":
			if b, ok := v.Value.(bool); ok {
				config.Enabled = b
			}
		case "ldapUrl":
			if s, ok := v.Value.(string); ok {
				config.URL = s
			}
		case "ldapBindDn":
			if s, ok := v.Value.(string); ok {
				config.BindDN = s
			}
		case "ldapBindPassword":
			if s, ok := v.Value.(string); ok {
				config.BindPassword = s
			}
		case "ldapBase":
			if s, ok := v.Value.(string); ok {
				config.BaseDN = s
			}
		case "ldapSkipCertVerify":
			if b, ok := v.Value.(bool); ok {
				config.SkipCertVerify = b
			}
		case "ldapUserSearchFilter":
			if s, ok := v.Value.(string); ok {
				config.UserSearchFilter = s
			}
		case "ldapGroupSearchFilter":
			if s, ok := v.Value.(string); ok {
				config.UserGroupSearchFilter = s
			}
		case "ldapAttributeUserUniqueIdentifier":
			if s, ok := v.Value.(string); ok {
				config.UserUniqueAttribute = s
			}
		case "ldapAttributeUserUsername":
			if s, ok := v.Value.(string); ok {
				config.UserUsernameAttribute = s
			}
		case "ldapAttributeUserEmail":
			if s, ok := v.Value.(string); ok {
				config.UserEmailAttribute = s
			}
		case "ldapAttributeUserFirstName":
			if s, ok := v.Value.(string); ok {
				config.UserFirstNameAttribute = s
			}
		case "ldapAttributeUserLastName":
			if s, ok := v.Value.(string); ok {
				config.UserLastNameAttribute = s
			}
		case "ldapAttributeGroupMember":
			if s, ok := v.Value.(string); ok {
				config.GroupMemberAttribute = s
			}
		case "ldapAttributeGroupUniqueIdentifier":
			if s, ok := v.Value.(string); ok {
				config.GroupUniqueAttribute = s
			}
		case "ldapAttributeGroupName":
			if s, ok := v.Value.(string); ok {
				config.GroupNameAttribute = s
			}
		case "ldapAttributeAdminGroup":
			if s, ok := v.Value.(string); ok {
				config.AdminGroupName = s
			}
		case "ldapSoftDeleteUsers":
			if b, ok := v.Value.(bool); ok {
				config.SoftDeleteUsers = b
			}
		}
	}

	return config, nil
}

// UpdateLDAPConfig updates the LDAP configuration
func (c *Client) UpdateLDAPConfig(req *LDAPConfigUpdateRequest) error {
	_, err := c.doRequest("PUT", "/api/application-configuration", req)
	return err
}

// SyncLDAP triggers an LDAP synchronization
func (c *Client) SyncLDAP() error {
	_, err := c.doRequest("POST", "/api/application-configuration/sync-ldap", nil)
	return err
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/client/... -run "TestGetLDAPConfig|TestUpdateLDAPConfig|TestSyncLDAP" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/client/app_config.go internal/client/app_config_test.go
git commit -m "feat(client): add app config API methods for LDAP"
```

---

## Task 4: Create LDAP Config Resource

**Files:**
- Create: `internal/resources/ldap_config_resource.go`

**Step 1: Create the resource implementation**

Create `internal/resources/ldap_config_resource.go`:

```go
package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ldapConfigResource{}
	_ resource.ResourceWithConfigure   = &ldapConfigResource{}
	_ resource.ResourceWithImportState = &ldapConfigResource{}
)

// NewLDAPConfigResource is a helper function to simplify the provider implementation.
func NewLDAPConfigResource() resource.Resource {
	return &ldapConfigResource{}
}

// ldapConfigResource is the resource implementation.
type ldapConfigResource struct {
	client *client.Client
}

// ldapConfigResourceModel maps the resource schema data.
type ldapConfigResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	SyncOnChange            types.Bool   `tfsdk:"sync_on_change"`
	URL                     types.String `tfsdk:"url"`
	BindDN                  types.String `tfsdk:"bind_dn"`
	BindPassword            types.String `tfsdk:"bind_password"`
	BaseDN                  types.String `tfsdk:"base_dn"`
	SkipCertVerify          types.Bool   `tfsdk:"skip_cert_verify"`
	UserSearchFilter        types.String `tfsdk:"user_search_filter"`
	UserGroupSearchFilter   types.String `tfsdk:"user_group_search_filter"`
	UserAttributes          types.Object `tfsdk:"user_attributes"`
	GroupAttributes         types.Object `tfsdk:"group_attributes"`
	SoftDeleteUsers         types.Bool   `tfsdk:"soft_delete_users"`
}

// userAttributesModel maps the user_attributes nested block
type userAttributesModel struct {
	UniqueIdentifier types.String `tfsdk:"unique_identifier"`
	Username         types.String `tfsdk:"username"`
	Email            types.String `tfsdk:"email"`
	FirstName        types.String `tfsdk:"first_name"`
	LastName         types.String `tfsdk:"last_name"`
}

// groupAttributesModel maps the group_attributes nested block
type groupAttributesModel struct {
	Member           types.String `tfsdk:"member"`
	UniqueIdentifier types.String `tfsdk:"unique_identifier"`
	Name             types.String `tfsdk:"name"`
	AdminGroup       types.String `tfsdk:"admin_group"`
}

var userAttributesAttrTypes = map[string]attr.Type{
	"unique_identifier": types.StringType,
	"username":          types.StringType,
	"email":             types.StringType,
	"first_name":        types.StringType,
	"last_name":         types.StringType,
}

var groupAttributesAttrTypes = map[string]attr.Type{
	"member":            types.StringType,
	"unique_identifier": types.StringType,
	"name":              types.StringType,
	"admin_group":       types.StringType,
}

// Metadata returns the resource type name.
func (r *ldapConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ldap_config"
}

// Schema defines the schema for the resource.
func (r *ldapConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages LDAP configuration in Pocket-ID.",
		MarkdownDescription: "Manages LDAP configuration in Pocket-ID. This is a singleton resource - only one LDAP configuration exists per Pocket-ID instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the LDAP configuration (always 'ldap').",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable or disable LDAP integration.",
				Required:    true,
			},
			"sync_on_change": schema.BoolAttribute{
				Description: "Trigger LDAP sync after configuration changes.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"url": schema.StringAttribute{
				Description: "LDAP server URL (ldap:// or ldaps://).",
				Optional:    true,
				Validators: []validator.String{
					LDAPURLValidator{},
				},
			},
			"bind_dn": schema.StringAttribute{
				Description: "Distinguished Name for LDAP bind authentication.",
				Optional:    true,
				Validators: []validator.String{
					DNValidator{},
				},
			},
			"bind_password": schema.StringAttribute{
				Description: "Password for bind DN.",
				Optional:    true,
				Sensitive:   true,
			},
			"base_dn": schema.StringAttribute{
				Description: "Base DN for LDAP searches.",
				Optional:    true,
				Validators: []validator.String{
					DNValidator{},
				},
			},
			"skip_cert_verify": schema.BoolAttribute{
				Description: "Skip TLS certificate verification for LDAPS.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"user_search_filter": schema.StringAttribute{
				Description: "LDAP filter for finding users.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("(objectClass=person)"),
			},
			"user_group_search_filter": schema.StringAttribute{
				Description: "LDAP filter for finding groups.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("(objectClass=groupOfNames)"),
			},
			"user_attributes": schema.SingleNestedAttribute{
				Description: "User attribute mappings from LDAP.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"unique_identifier": schema.StringAttribute{
						Description: "LDAP attribute for unique user identifier.",
						Optional:    true,
					},
					"username": schema.StringAttribute{
						Description: "LDAP attribute for username.",
						Optional:    true,
					},
					"email": schema.StringAttribute{
						Description: "LDAP attribute for email.",
						Optional:    true,
					},
					"first_name": schema.StringAttribute{
						Description: "LDAP attribute for first name.",
						Optional:    true,
					},
					"last_name": schema.StringAttribute{
						Description: "LDAP attribute for last name.",
						Optional:    true,
					},
				},
			},
			"group_attributes": schema.SingleNestedAttribute{
				Description: "Group attribute mappings from LDAP.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"member": schema.StringAttribute{
						Description: "LDAP attribute for group members.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("member"),
					},
					"unique_identifier": schema.StringAttribute{
						Description: "LDAP attribute for unique group identifier.",
						Optional:    true,
					},
					"name": schema.StringAttribute{
						Description: "LDAP attribute for group name.",
						Optional:    true,
					},
					"admin_group": schema.StringAttribute{
						Description: "Name of LDAP group that grants admin role.",
						Optional:    true,
					},
				},
			},
			"soft_delete_users": schema.BoolAttribute{
				Description: "When true, users not in LDAP are disabled instead of deleted.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ldapConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *ldapConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ldapConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields when enabled
	if plan.Enabled.ValueBool() {
		if plan.URL.IsNull() || plan.URL.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(path.Root("url"), "Missing Required Attribute", "url is required when enabled is true")
		}
		if plan.BindDN.IsNull() || plan.BindDN.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(path.Root("bind_dn"), "Missing Required Attribute", "bind_dn is required when enabled is true")
		}
		if plan.BindPassword.IsNull() || plan.BindPassword.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(path.Root("bind_password"), "Missing Required Attribute", "bind_password is required when enabled is true")
		}
		if plan.BaseDN.IsNull() || plan.BaseDN.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(path.Root("base_dn"), "Missing Required Attribute", "base_dn is required when enabled is true")
		}
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Build update request
	updateReq := r.buildUpdateRequest(ctx, &plan)

	tflog.Debug(ctx, "Creating LDAP configuration", map[string]any{
		"enabled": updateReq.LdapEnabled,
		"url":     updateReq.LdapUrl,
	})

	// Update the configuration
	err := r.client.UpdateLDAPConfig(updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating LDAP configuration",
			"Could not create LDAP configuration: "+err.Error(),
		)
		return
	}

	// Trigger sync if requested
	if plan.SyncOnChange.ValueBool() && plan.Enabled.ValueBool() {
		tflog.Debug(ctx, "Triggering LDAP sync after config creation")
		if err := r.client.SyncLDAP(); err != nil {
			resp.Diagnostics.AddWarning(
				"LDAP Sync Failed",
				"Configuration was saved but sync failed: "+err.Error(),
			)
		}
	}

	// Set the ID
	plan.ID = types.StringValue("ldap")

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ldapConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ldapConfigResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading LDAP configuration")

	config, err := r.client.GetLDAPConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading LDAP configuration",
			"Could not read LDAP configuration: "+err.Error(),
		)
		return
	}

	// Map API response to state
	state.ID = types.StringValue("ldap")
	state.Enabled = types.BoolValue(config.Enabled)
	state.URL = types.StringValue(config.URL)
	state.BindDN = types.StringValue(config.BindDN)
	// Note: bind_password is not returned by API, preserve from state
	state.BaseDN = types.StringValue(config.BaseDN)
	state.SkipCertVerify = types.BoolValue(config.SkipCertVerify)
	state.UserSearchFilter = types.StringValue(config.UserSearchFilter)
	state.UserGroupSearchFilter = types.StringValue(config.UserGroupSearchFilter)
	state.SoftDeleteUsers = types.BoolValue(config.SoftDeleteUsers)

	// Map user attributes
	userAttrs := userAttributesModel{
		UniqueIdentifier: types.StringValue(config.UserUniqueAttribute),
		Username:         types.StringValue(config.UserUsernameAttribute),
		Email:            types.StringValue(config.UserEmailAttribute),
		FirstName:        types.StringValue(config.UserFirstNameAttribute),
		LastName:         types.StringValue(config.UserLastNameAttribute),
	}
	userAttrsObj, diags := types.ObjectValueFrom(ctx, userAttributesAttrTypes, userAttrs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.UserAttributes = userAttrsObj

	// Map group attributes
	groupAttrs := groupAttributesModel{
		Member:           types.StringValue(config.GroupMemberAttribute),
		UniqueIdentifier: types.StringValue(config.GroupUniqueAttribute),
		Name:             types.StringValue(config.GroupNameAttribute),
		AdminGroup:       types.StringValue(config.AdminGroupName),
	}
	groupAttrsObj, diags := types.ObjectValueFrom(ctx, groupAttributesAttrTypes, groupAttrs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.GroupAttributes = groupAttrsObj

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ldapConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ldapConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields when enabled
	if plan.Enabled.ValueBool() {
		if plan.URL.IsNull() || plan.URL.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(path.Root("url"), "Missing Required Attribute", "url is required when enabled is true")
		}
		if plan.BindDN.IsNull() || plan.BindDN.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(path.Root("bind_dn"), "Missing Required Attribute", "bind_dn is required when enabled is true")
		}
		if plan.BindPassword.IsNull() || plan.BindPassword.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(path.Root("bind_password"), "Missing Required Attribute", "bind_password is required when enabled is true")
		}
		if plan.BaseDN.IsNull() || plan.BaseDN.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(path.Root("base_dn"), "Missing Required Attribute", "base_dn is required when enabled is true")
		}
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateReq := r.buildUpdateRequest(ctx, &plan)

	tflog.Debug(ctx, "Updating LDAP configuration", map[string]any{
		"enabled": updateReq.LdapEnabled,
	})

	err := r.client.UpdateLDAPConfig(updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating LDAP configuration",
			"Could not update LDAP configuration: "+err.Error(),
		)
		return
	}

	// Trigger sync if requested
	if plan.SyncOnChange.ValueBool() && plan.Enabled.ValueBool() {
		tflog.Debug(ctx, "Triggering LDAP sync after config update")
		if err := r.client.SyncLDAP(); err != nil {
			resp.Diagnostics.AddWarning(
				"LDAP Sync Failed",
				"Configuration was saved but sync failed: "+err.Error(),
			)
		}
	}

	plan.ID = types.StringValue("ldap")

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ldapConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Disabling LDAP configuration (delete)")

	// Delete = disable LDAP
	updateReq := &client.LDAPConfigUpdateRequest{
		LdapEnabled: false,
	}

	err := r.client.UpdateLDAPConfig(updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error disabling LDAP configuration",
			"Could not disable LDAP configuration: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *ldapConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildUpdateRequest converts the resource model to an API update request
func (r *ldapConfigResource) buildUpdateRequest(ctx context.Context, plan *ldapConfigResourceModel) *client.LDAPConfigUpdateRequest {
	req := &client.LDAPConfigUpdateRequest{
		LdapEnabled:         plan.Enabled.ValueBool(),
		LdapUrl:             plan.URL.ValueString(),
		LdapBindDn:          plan.BindDN.ValueString(),
		LdapBindPassword:    plan.BindPassword.ValueString(),
		LdapBase:            plan.BaseDN.ValueString(),
		LdapSkipCertVerify:  plan.SkipCertVerify.ValueBool(),
		LdapUserSearchFilter:  plan.UserSearchFilter.ValueString(),
		LdapGroupSearchFilter: plan.UserGroupSearchFilter.ValueString(),
		LdapSoftDeleteUsers:   plan.SoftDeleteUsers.ValueBool(),
	}

	// Extract user attributes
	if !plan.UserAttributes.IsNull() {
		var userAttrs userAttributesModel
		diags := plan.UserAttributes.As(ctx, &userAttrs, types.ObjectAsOptions{})
		if !diags.HasError() {
			req.LdapAttributeUserUniqueIdentifier = userAttrs.UniqueIdentifier.ValueString()
			req.LdapAttributeUserUsername = userAttrs.Username.ValueString()
			req.LdapAttributeUserEmail = userAttrs.Email.ValueString()
			req.LdapAttributeUserFirstName = userAttrs.FirstName.ValueString()
			req.LdapAttributeUserLastName = userAttrs.LastName.ValueString()
		}
	}

	// Extract group attributes
	if !plan.GroupAttributes.IsNull() {
		var groupAttrs groupAttributesModel
		diags := plan.GroupAttributes.As(ctx, &groupAttrs, types.ObjectAsOptions{})
		if !diags.HasError() {
			req.LdapAttributeGroupMember = groupAttrs.Member.ValueString()
			req.LdapAttributeGroupUniqueIdentifier = groupAttrs.UniqueIdentifier.ValueString()
			req.LdapAttributeGroupName = groupAttrs.Name.ValueString()
			req.LdapAttributeAdminGroup = groupAttrs.AdminGroup.ValueString()
		}
	}

	return req
}
```

**Step 2: Verify build**

Run: `go build ./...`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/resources/ldap_config_resource.go
git commit -m "feat(resources): add LDAP config resource"
```

---

## Task 5: Create LDAP Sync Resource

**Files:**
- Create: `internal/resources/ldap_sync_resource.go`

**Step 1: Create the resource implementation**

Create `internal/resources/ldap_sync_resource.go`:

```go
package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ldapSyncResource{}
	_ resource.ResourceWithConfigure   = &ldapSyncResource{}
	_ resource.ResourceWithImportState = &ldapSyncResource{}
)

// NewLDAPSyncResource is a helper function to simplify the provider implementation.
func NewLDAPSyncResource() resource.Resource {
	return &ldapSyncResource{}
}

// ldapSyncResource is the resource implementation.
type ldapSyncResource struct {
	client *client.Client
}

// ldapSyncResourceModel maps the resource schema data.
type ldapSyncResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Triggers types.Map    `tfsdk:"triggers"`
	LastSync types.String `tfsdk:"last_sync"`
}

// Metadata returns the resource type name.
func (r *ldapSyncResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ldap_sync"
}

// Schema defines the schema for the resource.
func (r *ldapSyncResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Triggers LDAP synchronization in Pocket-ID.",
		MarkdownDescription: "Triggers LDAP synchronization in Pocket-ID. Use the `triggers` attribute to control when sync occurs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the sync resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"triggers": schema.MapAttribute{
				Description:         "A map of values that, when changed, will trigger a new LDAP sync.",
				MarkdownDescription: "A map of values that, when changed, will trigger a new LDAP sync. Use `timestamp()` to sync on every apply.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"last_sync": schema.StringAttribute{
				Description: "Timestamp of the last successful LDAP sync.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ldapSyncResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *ldapSyncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ldapSyncResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Triggering LDAP sync (create)")

	err := r.client.SyncLDAP()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error triggering LDAP sync",
			"Could not trigger LDAP sync: "+err.Error(),
		)
		return
	}

	// Set computed values
	plan.ID = types.StringValue("ldap-sync")
	plan.LastSync = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ldapSyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ldapSyncResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Nothing to read from API - sync is fire-and-forget
	// Just preserve the current state

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ldapSyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ldapSyncResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Triggering LDAP sync (update - triggers changed)")

	err := r.client.SyncLDAP()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error triggering LDAP sync",
			"Could not trigger LDAP sync: "+err.Error(),
		)
		return
	}

	// Update last_sync timestamp
	plan.ID = types.StringValue("ldap-sync")
	plan.LastSync = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ldapSyncResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nothing to delete - sync is fire-and-forget
	tflog.Debug(ctx, "Removing LDAP sync resource from state (no API call needed)")
}

// ImportState imports an existing resource into Terraform.
func (r *ldapSyncResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

**Step 2: Verify build**

Run: `go build ./...`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/resources/ldap_sync_resource.go
git commit -m "feat(resources): add LDAP sync resource"
```

---

## Task 6: Register Resources in Provider

**Files:**
- Modify: `internal/provider/provider.go`

**Step 1: Add new resources to provider**

In `internal/provider/provider.go`, update the `Resources` function (around line 208):

```go
// Resources defines the resources implemented in the provider.
func (p *pocketIDProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewClientResource,
		resources.NewUserResource,
		resources.NewGroupResource,
		resources.NewOneTimeAccessTokenResource,
		resources.NewLDAPConfigResource,
		resources.NewLDAPSyncResource,
	}
}
```

**Step 2: Verify build**

Run: `go build ./...`
Expected: Build succeeds

**Step 3: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass

**Step 4: Commit**

```bash
git add internal/provider/provider.go
git commit -m "feat(provider): register LDAP config and sync resources"
```

---

## Task 7: Create Resource Tests

**Files:**
- Create: `internal/resources/ldap_config_resource_test.go`
- Create: `internal/resources/ldap_sync_resource_test.go`

**Step 1: Create LDAP config resource tests**

Create `internal/resources/ldap_config_resource_test.go`:

```go
package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"

	"github.com/Trozz/terraform-provider-pocketid/internal/resources"
)

func TestLDAPConfigResource_Schema(t *testing.T) {
	ctx := context.Background()
	schemaRequest := resource.SchemaRequest{}
	schemaResponse := &resource.SchemaResponse{}

	resources.NewLDAPConfigResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema returned diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Verify required attributes
	enabledAttr, ok := schemaResponse.Schema.Attributes["enabled"]
	assert.True(t, ok, "enabled attribute should exist")
	assert.True(t, enabledAttr.IsRequired(), "enabled should be required")

	// Verify computed attributes
	idAttr, ok := schemaResponse.Schema.Attributes["id"]
	assert.True(t, ok, "id attribute should exist")
	assert.True(t, idAttr.IsComputed(), "id should be computed")

	// Verify sensitive attributes
	bindPasswordAttr, ok := schemaResponse.Schema.Attributes["bind_password"]
	assert.True(t, ok, "bind_password attribute should exist")
	assert.True(t, bindPasswordAttr.IsSensitive(), "bind_password should be sensitive")

	// Verify optional attributes with defaults
	skipCertVerifyAttr, ok := schemaResponse.Schema.Attributes["skip_cert_verify"]
	assert.True(t, ok, "skip_cert_verify attribute should exist")
	assert.True(t, skipCertVerifyAttr.IsOptional(), "skip_cert_verify should be optional")
	assert.True(t, skipCertVerifyAttr.IsComputed(), "skip_cert_verify should be computed")

	softDeleteUsersAttr, ok := schemaResponse.Schema.Attributes["soft_delete_users"]
	assert.True(t, ok, "soft_delete_users attribute should exist")
	assert.True(t, softDeleteUsersAttr.IsOptional(), "soft_delete_users should be optional")
	assert.True(t, softDeleteUsersAttr.IsComputed(), "soft_delete_users should be computed")

	// Verify nested attributes exist
	userAttributesAttr, ok := schemaResponse.Schema.Attributes["user_attributes"]
	assert.True(t, ok, "user_attributes attribute should exist")
	assert.True(t, userAttributesAttr.IsOptional(), "user_attributes should be optional")

	groupAttributesAttr, ok := schemaResponse.Schema.Attributes["group_attributes"]
	assert.True(t, ok, "group_attributes attribute should exist")
	assert.True(t, groupAttributesAttr.IsOptional(), "group_attributes should be optional")

	// Verify sync_on_change attribute
	syncOnChangeAttr, ok := schemaResponse.Schema.Attributes["sync_on_change"]
	assert.True(t, ok, "sync_on_change attribute should exist")
	assert.True(t, syncOnChangeAttr.IsOptional(), "sync_on_change should be optional")
	assert.True(t, syncOnChangeAttr.IsComputed(), "sync_on_change should be computed")
}

func TestLDAPConfigResource_Metadata(t *testing.T) {
	ctx := context.Background()
	metadataRequest := resource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	metadataResponse := &resource.MetadataResponse{}

	resources.NewLDAPConfigResource().Metadata(ctx, metadataRequest, metadataResponse)

	assert.Equal(t, "pocketid_ldap_config", metadataResponse.TypeName)
}
```

**Step 2: Create LDAP sync resource tests**

Create `internal/resources/ldap_sync_resource_test.go`:

```go
package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"

	"github.com/Trozz/terraform-provider-pocketid/internal/resources"
)

func TestLDAPSyncResource_Schema(t *testing.T) {
	ctx := context.Background()
	schemaRequest := resource.SchemaRequest{}
	schemaResponse := &resource.SchemaResponse{}

	resources.NewLDAPSyncResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema returned diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Verify computed attributes
	idAttr, ok := schemaResponse.Schema.Attributes["id"]
	assert.True(t, ok, "id attribute should exist")
	assert.True(t, idAttr.IsComputed(), "id should be computed")

	lastSyncAttr, ok := schemaResponse.Schema.Attributes["last_sync"]
	assert.True(t, ok, "last_sync attribute should exist")
	assert.True(t, lastSyncAttr.IsComputed(), "last_sync should be computed")

	// Verify optional attributes
	triggersAttr, ok := schemaResponse.Schema.Attributes["triggers"]
	assert.True(t, ok, "triggers attribute should exist")
	assert.True(t, triggersAttr.IsOptional(), "triggers should be optional")
}

func TestLDAPSyncResource_Metadata(t *testing.T) {
	ctx := context.Background()
	metadataRequest := resource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	metadataResponse := &resource.MetadataResponse{}

	resources.NewLDAPSyncResource().Metadata(ctx, metadataRequest, metadataResponse)

	assert.Equal(t, "pocketid_ldap_sync", metadataResponse.TypeName)
}
```

**Step 3: Run tests**

Run: `go test ./internal/resources/... -v`
Expected: All tests pass

**Step 4: Commit**

```bash
git add internal/resources/ldap_config_resource_test.go internal/resources/ldap_sync_resource_test.go
git commit -m "test(resources): add LDAP resource unit tests"
```

---

## Task 8: Create Examples and Documentation

**Files:**
- Create: `examples/resources/pocketid_ldap_config/resource.tf`
- Create: `examples/resources/pocketid_ldap_sync/resource.tf`
- Create: `templates/resources/pocketid_ldap_config.md.tmpl`
- Create: `templates/resources/pocketid_ldap_sync.md.tmpl`

**Step 1: Create LDAP config example**

Create `examples/resources/pocketid_ldap_config/resource.tf`:

```hcl
# Basic LDAP configuration
resource "pocketid_ldap_config" "main" {
  enabled        = true
  sync_on_change = true

  # Connection settings
  url              = "ldaps://ldap.example.com:636"
  bind_dn          = "cn=admin,dc=example,dc=com"
  bind_password    = var.ldap_bind_password
  base_dn          = "dc=example,dc=com"
  skip_cert_verify = false

  # Search filters
  user_search_filter       = "(objectClass=person)"
  user_group_search_filter = "(objectClass=groupOfNames)"

  # User attribute mappings
  user_attributes {
    unique_identifier = "objectGUID"
    username          = "sAMAccountName"
    email             = "mail"
    first_name        = "givenName"
    last_name         = "sn"
  }

  # Group attribute mappings
  group_attributes {
    member            = "member"
    unique_identifier = "objectGUID"
    name              = "cn"
    admin_group       = "PocketID-Admins"
  }

  # Behavior settings
  soft_delete_users = true
}

variable "ldap_bind_password" {
  type      = string
  sensitive = true
}
```

**Step 2: Create LDAP sync example**

Create `examples/resources/pocketid_ldap_sync/resource.tf`:

```hcl
# Trigger LDAP sync on every apply
resource "pocketid_ldap_sync" "sync" {
  triggers = {
    timestamp = timestamp()
  }
}

# Trigger LDAP sync only when config changes
resource "pocketid_ldap_sync" "on_config_change" {
  triggers = {
    config_id = pocketid_ldap_config.main.id
  }
}

# Manual sync trigger (change the value to trigger)
resource "pocketid_ldap_sync" "manual" {
  triggers = {
    manual = "2024-01-15"
  }
}
```

**Step 3: Create LDAP config doc template**

Create `templates/resources/pocketid_ldap_config.md.tmpl`:

```markdown
---
page_title: "{{.Name}} {{.Type}} - {{.ProviderName}}"
subcategory: ""
description: |-
{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

This is a singleton resource - only one LDAP configuration exists per Pocket-ID instance.

## Example Usage

{{ tffile "examples/resources/pocketid_ldap_config/resource.tf" }}

## Argument Reference

{{ .SchemaMarkdown | trimspace }}

## Import

Import is supported using the following syntax:

```shell
terraform import pocketid_ldap_config.main ldap
```

After import, you must provide the `bind_password` in your configuration as it is not returned by the API.
```

**Step 4: Create LDAP sync doc template**

Create `templates/resources/pocketid_ldap_sync.md.tmpl`:

```markdown
---
page_title: "{{.Name}} {{.Type}} - {{.ProviderName}}"
subcategory: ""
description: |-
{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/resources/pocketid_ldap_sync/resource.tf" }}

## Argument Reference

{{ .SchemaMarkdown | trimspace }}

## Import

Import is supported using the following syntax:

```shell
terraform import pocketid_ldap_sync.sync ldap-sync
```
```

**Step 5: Create example directories**

Run:
```bash
mkdir -p examples/resources/pocketid_ldap_config
mkdir -p examples/resources/pocketid_ldap_sync
```

**Step 6: Generate documentation**

Run: `go generate ./...`
Expected: Documentation generated in docs/ directory

**Step 7: Commit**

```bash
git add examples/ templates/
git commit -m "docs: add LDAP resource examples and documentation templates"
```

---

## Final Verification

**Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: All tests pass

**Step 2: Build provider**

Run: `go build -o terraform-provider-pocketid`
Expected: Build succeeds

**Step 3: Verify documentation**

Run: `go generate ./...`
Expected: Documentation generated successfully

---

## Summary

This plan implements:

1. **LDAP Models** (Task 1) - Data structures for API communication
2. **Validators** (Task 2) - LDAP URL and DN validation with tests
3. **Client Methods** (Task 3) - GetLDAPConfig, UpdateLDAPConfig, SyncLDAP
4. **LDAP Config Resource** (Task 4) - Full CRUD with nested blocks
5. **LDAP Sync Resource** (Task 5) - Trigger-based sync
6. **Provider Registration** (Task 6) - Register new resources
7. **Unit Tests** (Task 7) - Schema and metadata tests
8. **Documentation** (Task 8) - Examples and templates

**Parallel execution opportunities:**
- Tasks 1 & 2 can run in parallel
- Tasks 4 & 5 can run in parallel (after Task 3)
- Tasks 7 & 8 can run in parallel (after Task 6)
