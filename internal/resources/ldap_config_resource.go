package resources

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

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

// ldapConfigResourceModel describes the resource data model.
type ldapConfigResourceModel struct {
	ID                    types.String              `tfsdk:"id"`
	Enabled               types.Bool                `tfsdk:"enabled"`
	URL                   types.String              `tfsdk:"url"`
	BindDN                types.String              `tfsdk:"bind_dn"`
	BindPassword          types.String              `tfsdk:"bind_password"`
	BaseDN                types.String              `tfsdk:"base_dn"`
	SkipCertVerify        types.Bool                `tfsdk:"skip_cert_verify"`
	UserSearchFilter      types.String              `tfsdk:"user_search_filter"`
	UserGroupSearchFilter types.String              `tfsdk:"user_group_search_filter"`
	UserAttributes        *ldapUserAttributesModel  `tfsdk:"user_attributes"`
	GroupAttributes       *ldapGroupAttributesModel `tfsdk:"group_attributes"`
	SoftDeleteUsers       types.Bool                `tfsdk:"soft_delete_users"`
}

// ldapUserAttributesModel describes the user attribute mappings.
type ldapUserAttributesModel struct {
	UniqueIdentifier types.String `tfsdk:"unique_identifier"`
	Username         types.String `tfsdk:"username"`
	Email            types.String `tfsdk:"email"`
	FirstName        types.String `tfsdk:"first_name"`
	LastName         types.String `tfsdk:"last_name"`
	ProfilePicture   types.String `tfsdk:"profile_picture"`
}

// ldapGroupAttributesModel describes the group attribute mappings.
type ldapGroupAttributesModel struct {
	Member           types.String `tfsdk:"member"`
	UniqueIdentifier types.String `tfsdk:"unique_identifier"`
	Name             types.String `tfsdk:"name"`
	AdminGroupName   types.String `tfsdk:"admin_group_name"`
}

// Metadata returns the resource type name.
func (r *ldapConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ldap_config"
}

// Schema defines the schema for the resource.
func (r *ldapConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages LDAP configuration for Pocket-ID.",
		MarkdownDescription: `Manages LDAP configuration for Pocket-ID. This resource allows you to configure LDAP/Active Directory integration for user and group synchronization.

~> **Note** Setting enabled to false effectively "deletes" the LDAP configuration. The bind_password is sensitive and will not be shown in logs or plan output.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the LDAP configuration (always 'ldap').",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable or disable LDAP integration. When false, all other settings are ignored by Pocket ID.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "LDAP server URL. Must be in format: ldap://host:port or ldaps://host:port.",
				Optional:    true,
				Validators: []validator.String{
					&ldapURLValidator{},
				},
			},
			"bind_dn": schema.StringAttribute{
				Description: "Distinguished Name for LDAP bind authentication. Example: cn=admin,dc=example,dc=com. Anonymous binding is not supported.",
				Optional:    true,
				Validators: []validator.String{
					&dnValidator{},
				},
			},
			"bind_password": schema.StringAttribute{
				Description: "Password for bind DN. Anonymous binding is not supported.",
				Optional:    true,
				Sensitive:   true,
			},
			"base_dn": schema.StringAttribute{
				Description: "Base DN for LDAP searches. Example: dc=example,dc=com.",
				Optional:    true,
				Validators: []validator.String{
					&dnValidator{},
				},
			},
			"skip_cert_verify": schema.BoolAttribute{
				Description: "Skip TLS certificate verification for LDAPS connections. Use with caution in production.",
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
			"soft_delete_users": schema.BoolAttribute{
				Description: "When true, users not found in LDAP are disabled. When false, they are deleted.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
		Blocks: map[string]schema.Block{
			"user_attributes": schema.SingleNestedBlock{
				Description: "LDAP attribute mappings for users.",
				Attributes: map[string]schema.Attribute{
					"unique_identifier": schema.StringAttribute{
						Description: "LDAP attribute containing unique user identifier. Examples: objectGUID, entryUUID, uid.",
						Optional:    true,
					},
					"username": schema.StringAttribute{
						Description: "LDAP attribute for username. Examples: sAMAccountName, uid, cn.",
						Optional:    true,
					},
					"email": schema.StringAttribute{
						Description: "LDAP attribute for email address. Example: mail.",
						Optional:    true,
					},
					"first_name": schema.StringAttribute{
						Description: "LDAP attribute for first name. Example: givenName.",
						Optional:    true,
					},
					"last_name": schema.StringAttribute{
						Description: "LDAP attribute for last name. Example: sn.",
						Optional:    true,
					},
					"profile_picture": schema.StringAttribute{
						Description: "LDAP attribute for profile picture. Can be URL, base64, or binary data. Example: thumbnailPhoto.",
						Optional:    true,
					},
				},
			},
			"group_attributes": schema.SingleNestedBlock{
				Description: "LDAP attribute mappings for groups.",
				Attributes: map[string]schema.Attribute{
					"member": schema.StringAttribute{
						Description: "LDAP attribute for group members.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("member"),
					},
					"unique_identifier": schema.StringAttribute{
						Description: "LDAP attribute for unique group identifier. Examples: objectGUID, entryUUID, cn.",
						Optional:    true,
					},
					"name": schema.StringAttribute{
						Description: "LDAP attribute for group name. Example: cn.",
						Optional:    true,
					},
					"admin_group_name": schema.StringAttribute{
						Description: "Name of LDAP group that grants admin role. Members of this group become Pocket ID admins.",
						Optional:    true,
					},
				},
			},
		},
	}
}

// ldapURLValidator validates LDAP URLs.
type ldapURLValidator struct{}

// Description returns a plain text description of the validator's behavior.
func (v ldapURLValidator) Description(_ context.Context) string {
	return "value must be a valid LDAP URL (ldap:// or ldaps://)"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior.
func (v ldapURLValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v ldapURLValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	parsedURL, err := url.Parse(value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid LDAP URL",
			fmt.Sprintf("The provided URL %q is not valid: %s", value, err),
		)
		return
	}

	if parsedURL.Scheme != "ldap" && parsedURL.Scheme != "ldaps" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid LDAP URL Scheme",
			fmt.Sprintf("The URL must use ldap:// or ldaps:// scheme, got %q", parsedURL.Scheme),
		)
		return
	}

	if parsedURL.Host == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Host in LDAP URL",
			"The LDAP URL must specify a host",
		)
		return
	}
}

// dnValidator validates Distinguished Names.
type dnValidator struct{}

// Description returns a plain text description of the validator's behavior.
func (v dnValidator) Description(_ context.Context) string {
	return "value must be a valid Distinguished Name (DN)"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior.
func (v dnValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v dnValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	// Basic DN validation - check for key=value pairs
	if !strings.Contains(value, "=") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Distinguished Name",
			fmt.Sprintf("The value %q must be a valid DN (e.g., cn=admin,dc=example,dc=com)", value),
		)
		return
	}

	// More thorough DN validation using regex
	dnPattern := regexp.MustCompile(`^([A-Za-z][A-Za-z0-9-]*=[^,=]+)(,[A-Za-z][A-Za-z0-9-]*=[^,=]+)*$`)
	if !dnPattern.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Distinguished Name Format",
			fmt.Sprintf("The value %q is not a properly formatted DN", value),
		)
		return
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
	// Retrieve values from plan
	var plan ldapConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate conditional requirements when enabled=true
	if plan.Enabled.ValueBool() {
		if plan.URL.IsNull() || plan.URL.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("url"),
				"Missing Required Attribute",
				"The url attribute is required when enabled is true",
			)
		}
		if plan.BindDN.IsNull() || plan.BindDN.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("bind_dn"),
				"Missing Required Attribute",
				"The bind_dn attribute is required when enabled is true",
			)
		}
		if plan.BindPassword.IsNull() || plan.BindPassword.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("bind_password"),
				"Missing Required Attribute",
				"The bind_password attribute is required when enabled is true",
			)
		}
		if plan.BaseDN.IsNull() || plan.BaseDN.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("base_dn"),
				"Missing Required Attribute",
				"The base_dn attribute is required when enabled is true",
			)
		}
		if plan.UserAttributes == nil || plan.UserAttributes.UniqueIdentifier.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("user_attributes").AtName("unique_identifier"),
				"Missing Required Attribute",
				"The user_attributes.unique_identifier attribute is required when enabled is true",
			)
		}
		if plan.UserAttributes == nil || plan.UserAttributes.Username.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("user_attributes").AtName("username"),
				"Missing Required Attribute",
				"The user_attributes.username attribute is required when enabled is true",
			)
		}

		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Set the ID to "ldap" (singleton resource)
	plan.ID = types.StringValue("ldap")

	// Convert plan to API request
	updateReq := r.planToAPIRequest(&plan)

	// Call API to create/update LDAP configuration
	tflog.Debug(ctx, "Creating LDAP configuration", map[string]interface{}{
		"enabled": plan.Enabled.ValueBool(),
	})

	config, err := r.client.UpdateApplicationConfigurationWithContext(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating LDAP Configuration",
			fmt.Sprintf("Could not create LDAP configuration: %s", err),
		)
		return
	}

	// Update state from API response
	r.apiResponseToState(config, &plan)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ldapConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ldapConfigResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to read LDAP configuration
	tflog.Debug(ctx, "Reading LDAP configuration")

	config, err := r.client.GetApplicationConfigurationWithContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading LDAP Configuration",
			fmt.Sprintf("Could not read LDAP configuration: %s", err),
		)
		return
	}

	// Handle case where LDAP config doesn't exist or is disabled
	if config.LDAP == nil || config.LDAP.Enabled != "true" {
		// Resource has been deleted or disabled outside of Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state from API response
	r.apiResponseToState(config, &state)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ldapConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan ldapConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate conditional requirements when enabled=true
	if plan.Enabled.ValueBool() {
		if plan.URL.IsNull() || plan.URL.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("url"),
				"Missing Required Attribute",
				"The url attribute is required when enabled is true",
			)
		}
		if plan.BindDN.IsNull() || plan.BindDN.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("bind_dn"),
				"Missing Required Attribute",
				"The bind_dn attribute is required when enabled is true",
			)
		}
		if plan.BindPassword.IsNull() || plan.BindPassword.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("bind_password"),
				"Missing Required Attribute",
				"The bind_password attribute is required when enabled is true",
			)
		}
		if plan.BaseDN.IsNull() || plan.BaseDN.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("base_dn"),
				"Missing Required Attribute",
				"The base_dn attribute is required when enabled is true",
			)
		}
		if plan.UserAttributes == nil || plan.UserAttributes.UniqueIdentifier.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("user_attributes").AtName("unique_identifier"),
				"Missing Required Attribute",
				"The user_attributes.unique_identifier attribute is required when enabled is true",
			)
		}
		if plan.UserAttributes == nil || plan.UserAttributes.Username.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("user_attributes").AtName("username"),
				"Missing Required Attribute",
				"The user_attributes.username attribute is required when enabled is true",
			)
		}

		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Convert plan to API request
	updateReq := r.planToAPIRequest(&plan)

	// Call API to update LDAP configuration
	tflog.Debug(ctx, "Updating LDAP configuration", map[string]interface{}{
		"enabled": plan.Enabled.ValueBool(),
	})

	config, err := r.client.UpdateApplicationConfigurationWithContext(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating LDAP Configuration",
			fmt.Sprintf("Could not update LDAP configuration: %s", err),
		)
		return
	}

	// Update state from API response
	r.apiResponseToState(config, &plan)

	// Update the state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ldapConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state ldapConfigResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to disable LDAP configuration (set enabled=false)
	tflog.Debug(ctx, "Deleting LDAP configuration (setting enabled=false)")

	// Create a minimal request to disable LDAP
	updateReq := &client.ApplicationConfigurationUpdateRequest{
		LDAP: &client.LDAPConfiguration{
			Enabled: "false",
		},
	}

	_, err := r.client.UpdateApplicationConfigurationWithContext(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting LDAP Configuration",
			fmt.Sprintf("Could not disable LDAP configuration: %s", err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *ldapConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The import ID should be "ldap" (singleton resource)
	if req.ID != "ldap" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"The import ID must be 'ldap' for the LDAP configuration resource",
		)
		return
	}

	// Set the ID in state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "ldap")...)
}

// planToAPIRequest converts the Terraform plan to an API request
func (r *ldapConfigResource) planToAPIRequest(plan *ldapConfigResourceModel) *client.ApplicationConfigurationUpdateRequest {
	req := &client.ApplicationConfigurationUpdateRequest{
		LDAP: &client.LDAPConfiguration{
			Enabled: boolToString(plan.Enabled.ValueBool()),
		},
	}

	// Only populate other fields if LDAP is enabled
	if plan.Enabled.ValueBool() {
		req.LDAP.URL = plan.URL.ValueString()
		req.LDAP.BindDN = plan.BindDN.ValueString()
		req.LDAP.BindPassword = plan.BindPassword.ValueString()
		req.LDAP.BaseDN = plan.BaseDN.ValueString()
		req.LDAP.SkipCertVerify = boolToString(plan.SkipCertVerify.ValueBool())
		req.LDAP.UserSearchFilter = plan.UserSearchFilter.ValueString()
		req.LDAP.UserGroupSearchFilter = plan.UserGroupSearchFilter.ValueString()
		req.LDAP.SoftDeleteUsers = boolToString(plan.SoftDeleteUsers.ValueBool())

		// User attributes
		if plan.UserAttributes != nil {
			req.LDAP.UserAttributes = &client.LDAPUserAttributes{
				UniqueIdentifier: plan.UserAttributes.UniqueIdentifier.ValueString(),
				Username:         plan.UserAttributes.Username.ValueString(),
				Email:            plan.UserAttributes.Email.ValueString(),
				FirstName:        plan.UserAttributes.FirstName.ValueString(),
				LastName:         plan.UserAttributes.LastName.ValueString(),
				ProfilePicture:   plan.UserAttributes.ProfilePicture.ValueString(),
			}
		}

		// Group attributes
		if plan.GroupAttributes != nil {
			req.LDAP.GroupAttributes = &client.LDAPGroupAttributes{
				Member:           plan.GroupAttributes.Member.ValueString(),
				UniqueIdentifier: plan.GroupAttributes.UniqueIdentifier.ValueString(),
				Name:             plan.GroupAttributes.Name.ValueString(),
				AdminGroupName:   plan.GroupAttributes.AdminGroupName.ValueString(),
			}
		}
	}

	return req
}

// apiResponseToState updates the Terraform state from an API response
func (r *ldapConfigResource) apiResponseToState(config *client.ApplicationConfiguration, state *ldapConfigResourceModel) {
	if config.LDAP == nil {
		// LDAP not configured, set to disabled
		state.Enabled = types.BoolValue(false)
		return
	}

	// Core settings
	state.ID = types.StringValue("ldap")
	state.Enabled = types.BoolValue(config.LDAP.Enabled == "true")

	// Connection settings
	if config.LDAP.URL != "" {
		state.URL = types.StringValue(config.LDAP.URL)
	}
	if config.LDAP.BindDN != "" {
		state.BindDN = types.StringValue(config.LDAP.BindDN)
	}
	// Don't update bind_password from API response as it's not returned
	if config.LDAP.BaseDN != "" {
		state.BaseDN = types.StringValue(config.LDAP.BaseDN)
	}
	if config.LDAP.SkipCertVerify != "" {
		state.SkipCertVerify = types.BoolValue(config.LDAP.SkipCertVerify == "true")
	} else {
		state.SkipCertVerify = types.BoolValue(false)
	}

	// Search filters
	if config.LDAP.UserSearchFilter != "" {
		state.UserSearchFilter = types.StringValue(config.LDAP.UserSearchFilter)
	} else {
		state.UserSearchFilter = types.StringValue("(objectClass=person)")
	}
	if config.LDAP.UserGroupSearchFilter != "" {
		state.UserGroupSearchFilter = types.StringValue(config.LDAP.UserGroupSearchFilter)
	} else {
		state.UserGroupSearchFilter = types.StringValue("(objectClass=groupOfNames)")
	}

	// Behavior settings
	if config.LDAP.SoftDeleteUsers != "" {
		state.SoftDeleteUsers = types.BoolValue(config.LDAP.SoftDeleteUsers == "true")
	} else {
		state.SoftDeleteUsers = types.BoolValue(true)
	}

	// User attributes
	if config.LDAP.UserAttributes != nil {
		if state.UserAttributes == nil {
			state.UserAttributes = &ldapUserAttributesModel{}
		}
		state.UserAttributes.UniqueIdentifier = types.StringValue(config.LDAP.UserAttributes.UniqueIdentifier)
		state.UserAttributes.Username = types.StringValue(config.LDAP.UserAttributes.Username)
		state.UserAttributes.Email = types.StringValue(config.LDAP.UserAttributes.Email)
		state.UserAttributes.FirstName = types.StringValue(config.LDAP.UserAttributes.FirstName)
		state.UserAttributes.LastName = types.StringValue(config.LDAP.UserAttributes.LastName)
		state.UserAttributes.ProfilePicture = types.StringValue(config.LDAP.UserAttributes.ProfilePicture)
	}

	// Group attributes
	if config.LDAP.GroupAttributes != nil {
		if state.GroupAttributes == nil {
			state.GroupAttributes = &ldapGroupAttributesModel{}
		}
		if config.LDAP.GroupAttributes.Member != "" {
			state.GroupAttributes.Member = types.StringValue(config.LDAP.GroupAttributes.Member)
		} else {
			state.GroupAttributes.Member = types.StringValue("member")
		}
		state.GroupAttributes.UniqueIdentifier = types.StringValue(config.LDAP.GroupAttributes.UniqueIdentifier)
		state.GroupAttributes.Name = types.StringValue(config.LDAP.GroupAttributes.Name)
		state.GroupAttributes.AdminGroupName = types.StringValue(config.LDAP.GroupAttributes.AdminGroupName)
	}
}

// boolToString converts a boolean to a string ("true" or "false")
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
