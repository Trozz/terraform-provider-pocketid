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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	ID                    types.String `tfsdk:"id"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	SyncOnChange          types.Bool   `tfsdk:"sync_on_change"`
	URL                   types.String `tfsdk:"url"`
	BindDN                types.String `tfsdk:"bind_dn"`
	BindPassword          types.String `tfsdk:"bind_password"`
	BaseDN                types.String `tfsdk:"base_dn"`
	SkipCertVerify        types.Bool   `tfsdk:"skip_cert_verify"`
	UserSearchFilter      types.String `tfsdk:"user_search_filter"`
	UserGroupSearchFilter types.String `tfsdk:"user_group_search_filter"`
	UserAttributes        types.Object `tfsdk:"user_attributes"`
	GroupAttributes       types.Object `tfsdk:"group_attributes"`
	SoftDeleteUsers       types.Bool   `tfsdk:"soft_delete_users"`
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
		LdapEnabled:           plan.Enabled.ValueBool(),
		LdapUrl:               plan.URL.ValueString(),
		LdapBindDn:            plan.BindDN.ValueString(),
		LdapBindPassword:      plan.BindPassword.ValueString(),
		LdapBase:              plan.BaseDN.ValueString(),
		LdapSkipCertVerify:    plan.SkipCertVerify.ValueBool(),
		LdapUserSearchFilter:  plan.UserSearchFilter.ValueString(),
		LdapGroupSearchFilter: plan.UserGroupSearchFilter.ValueString(),
		LdapSoftDeleteUsers:   plan.SoftDeleteUsers.ValueBool(),
	}

	// Extract user attributes
	if !plan.UserAttributes.IsNull() {
		var userAttrs userAttributesModel
		diags := plan.UserAttributes.As(ctx, &userAttrs, basetypes.ObjectAsOptions{})
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
		diags := plan.GroupAttributes.As(ctx, &groupAttrs, basetypes.ObjectAsOptions{})
		if !diags.HasError() {
			req.LdapAttributeGroupMember = groupAttrs.Member.ValueString()
			req.LdapAttributeGroupUniqueIdentifier = groupAttrs.UniqueIdentifier.ValueString()
			req.LdapAttributeGroupName = groupAttrs.Name.ValueString()
			req.LdapAttributeAdminGroup = groupAttrs.AdminGroup.ValueString()
		}
	}

	return req
}
