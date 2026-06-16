package resources

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &clientResource{}
	_ resource.ResourceWithConfigure      = &clientResource{}
	_ resource.ResourceWithImportState    = &clientResource{}
	_ resource.ResourceWithValidateConfig = &clientResource{}
)

// NewClientResource is a helper function to simplify the provider implementation.
func NewClientResource() resource.Resource {
	return &clientResource{}
}

// clientResource is the resource implementation.
type clientResource struct {
	client *client.Client
}

// clientResourceModel maps the resource schema data.
type clientResourceModel struct {
	ID                                  types.String `tfsdk:"id"`
	Name                                types.String `tfsdk:"name"`
	ClientID                            types.String `tfsdk:"client_id"`
	CallbackURLs                        types.List   `tfsdk:"callback_urls"`
	LogoutCallbackURLs                  types.List   `tfsdk:"logout_callback_urls"`
	IsPublic                            types.Bool   `tfsdk:"is_public"`
	PkceEnabled                         types.Bool   `tfsdk:"pkce_enabled"`
	AllowedUserGroups                   types.List   `tfsdk:"allowed_user_groups"`
	HasLogo                             types.Bool   `tfsdk:"has_logo"`
	RequiresReauthentication            types.Bool   `tfsdk:"requires_reauthentication"`
	RequiresPushedAuthorizationRequests types.Bool   `tfsdk:"requires_pushed_authorization_requests"`
	LaunchURL                           types.String `tfsdk:"launch_url"`
	FederatedIdentities                 types.List   `tfsdk:"federated_identities"`
	ClientSecret                        types.String `tfsdk:"client_secret"`
}

// clientFederatedIdentityModel maps a single federated identity nested object.
type clientFederatedIdentityModel struct {
	Issuer   types.String `tfsdk:"issuer"`
	Subject  types.String `tfsdk:"subject"`
	Audience types.String `tfsdk:"audience"`
	JWKS     types.String `tfsdk:"jwks"`
}

// federatedIdentityAttrTypes is the attribute-type map for a federated identity object.
var federatedIdentityAttrTypes = map[string]attr.Type{
	"issuer":   types.StringType,
	"subject":  types.StringType,
	"audience": types.StringType,
	"jwks":     types.StringType,
}

// Metadata returns the resource type name.
func (r *clientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

// Schema defines the schema for the resource.
func (r *clientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an OIDC client in Pocket-ID.",
		MarkdownDescription: `Manages an OIDC client in Pocket-ID. OIDC clients are applications that can authenticate users through Pocket-ID.

~> **Note** The client secret is only available during resource creation and cannot be retrieved later. Store it securely.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the OIDC client.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the OIDC client.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 50),
				},
			},
			"client_id": schema.StringAttribute{
				Description: "The client ID to use for the OIDC client. If not set, one will be generated. Must be between 2 and 128 characters.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 128), // Matches the API binding (min=2, max=128)
				},
			},
			"callback_urls": schema.ListAttribute{
				Description: "List of allowed callback URLs for the OIDC client.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(urlValidator{}),
				},
			},
			"logout_callback_urls": schema.ListAttribute{
				Description: "List of allowed logout callback URLs for the OIDC client.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(urlValidator{}),
				},
			},
			"is_public": schema.BoolAttribute{
				Description: "Whether this is a public client (no client secret). Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"requires_reauthentication": schema.BoolAttribute{
				Description: "Whether this client requires reauthentication for certain flows. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"requires_pushed_authorization_requests": schema.BoolAttribute{
				Description: "Whether this client requires Pushed Authorization Requests (PAR, RFC 9126). Defaults to false. " +
					"Applies to confidential clients only — Pocket-ID coerces this to false for public clients (is_public = true). " +
					"Enforced only by Pocket-ID versions that support PAR (v2.9.0+); on older versions the value is stored in state but not enforced.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"launch_url": schema.StringAttribute{
				Description: "Optional launch URL associated with the client.",
				Optional:    true,
				Computed:    true,
			},
			"federated_identities": schema.ListNestedAttribute{
				Description: "List of federated identities (workload identity federation) allowed to authenticate as this client.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"issuer": schema.StringAttribute{
							Description: "The issuer of the federated identity token.",
							Required:    true,
						},
						"subject": schema.StringAttribute{
							Description: "The expected subject of the federated identity token.",
							Optional:    true,
						},
						"audience": schema.StringAttribute{
							Description: "The expected audience of the federated identity token.",
							Optional:    true,
						},
						"jwks": schema.StringAttribute{
							Description: "Optional JWKS used to validate the federated identity token.",
							Optional:    true,
						},
					},
				},
			},
			"pkce_enabled": schema.BoolAttribute{
				Description: "Whether PKCE is enabled for this client. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"allowed_user_groups": schema.ListAttribute{
				Description: "List of user group IDs that are allowed to use this client. If empty, all users can use this client.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"has_logo": schema.BoolAttribute{
				Description: "Whether the client has a logo configured.",
				Computed:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "The client secret. Only available during resource creation for non-public clients.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *clientResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig rejects configurations the API would silently coerce, giving a
// clear plan-time error instead of an inconsistent-result error after apply.
func (r *clientResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config clientResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Pocket-ID coerces requires_pushed_authorization_requests to false for
	// public clients, so true + is_public is never satisfiable.
	if config.IsPublic.ValueBool() && config.RequiresPushedAuthorizationRequests.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root("requires_pushed_authorization_requests"),
			"Invalid PAR configuration",
			"requires_pushed_authorization_requests can only be true for confidential clients. "+
				"Set is_public = false to use Pushed Authorization Requests.",
		)
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *clientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan clientResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform types to Go types
	var callbackURLs []string
	diags = plan.CallbackURLs.ElementsAs(ctx, &callbackURLs, false)
	resp.Diagnostics.Append(diags...)

	var logoutCallbackURLs []string
	if !plan.LogoutCallbackURLs.IsNull() {
		diags = plan.LogoutCallbackURLs.ElementsAs(ctx, &logoutCallbackURLs, false)
		resp.Diagnostics.Append(diags...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Build the create request using helper
	createReq := buildCreateRequestFromPlan(ctx, &plan)
	if !plan.ClientID.IsNull() && !plan.ClientID.IsUnknown() && plan.ClientID.ValueString() != "" {
		cid := plan.ClientID.ValueString()
		createReq.ClientID = &cid
	}

	tflog.Debug(ctx, "Creating OIDC client", map[string]any{
		"name":     createReq.Name,
		"isPublic": createReq.IsPublic,
	})

	clientResp, err := r.client.CreateClient(createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating OIDC client",
			"Could not create OIDC client, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Created OIDC client", map[string]any{
		"id": clientResp.ID,
	})

	// Map API response to Terraform model and preserve fields
	apiModel := mapAPIClientToModel(ctx, clientResp)
	plan.ID = apiModel.ID
	plan.HasLogo = apiModel.HasLogo
	plan.RequiresReauthentication = apiModel.RequiresReauthentication
	plan.FederatedIdentities = apiModel.FederatedIdentities
	plan.LaunchURL = apiModel.LaunchURL
	// Preserve the configured PAR value when the server does not return the field
	// (Pocket-ID <= v2.8.0). Only override from the API when it is present.
	if clientResp.RequiresPushedAuthorizationRequests != nil {
		plan.RequiresPushedAuthorizationRequests = types.BoolValue(*clientResp.RequiresPushedAuthorizationRequests)
	}

	// Generate client secret for non-public clients
	if !plan.IsPublic.ValueBool() {
		tflog.Debug(ctx, "Generating client secret for non-public client")
		secret, err := r.client.GenerateClientSecret(clientResp.ID)
		if err != nil {
			// Try to clean up the created client
			_ = r.client.DeleteClient(clientResp.ID)
			resp.Diagnostics.AddError(
				"Error generating client secret",
				"Could not generate client secret, the client was deleted. Error: "+err.Error(),
			)
			return
		}
		plan.ClientSecret = types.StringValue(secret)
	} else {
		plan.ClientSecret = types.StringNull()
	}

	// Handle allowed user groups
	if !plan.AllowedUserGroups.IsNull() && !plan.AllowedUserGroups.IsUnknown() {
		var groupIDs []string
		diags = plan.AllowedUserGroups.ElementsAs(ctx, &groupIDs, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() && len(groupIDs) > 0 {
			tflog.Debug(ctx, "Updating allowed user groups", map[string]any{
				"groups": groupIDs,
			})
			err = r.client.UpdateClientAllowedUserGroups(clientResp.ID, groupIDs)
			if err != nil {
				// Try to clean up the created client
				_ = r.client.DeleteClient(clientResp.ID)
				resp.Diagnostics.AddError(
					"Error updating allowed user groups",
					"Could not update allowed user groups, the client was deleted. Error: "+err.Error(),
				)
				return
			}
		}
	}

	// Set the state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *clientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state clientResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading OIDC client", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Get client from API
	clientResp, err := r.client.GetClient(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading OIDC client",
			"Could not read OIDC client ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state from API response
	state.Name = types.StringValue(clientResp.Name)
	state.IsPublic = types.BoolValue(clientResp.IsPublic)
	state.PkceEnabled = types.BoolValue(clientResp.PkceEnabled)
	state.HasLogo = types.BoolValue(clientResp.HasLogo)
	state.RequiresReauthentication = types.BoolValue(clientResp.RequiresReauthentication)
	// Only refresh PAR from the API when the server returns the field; otherwise
	// preserve the existing state value (Pocket-ID <= v2.8.0 omits it). On import
	// there is no prior value, so fall back to the default of false.
	if clientResp.RequiresPushedAuthorizationRequests != nil {
		state.RequiresPushedAuthorizationRequests = types.BoolValue(*clientResp.RequiresPushedAuthorizationRequests)
	} else if state.RequiresPushedAuthorizationRequests.IsNull() || state.RequiresPushedAuthorizationRequests.IsUnknown() {
		state.RequiresPushedAuthorizationRequests = types.BoolValue(false)
	}
	state.FederatedIdentities = federatedIdentitiesToList(ctx, clientResp.Credentials.FederatedIdentities)
	if clientResp.LaunchURL != "" {
		state.LaunchURL = types.StringValue(clientResp.LaunchURL)
	} else {
		state.LaunchURL = types.StringNull()
	}

	// Update callback URLs
	callbackURLs, diags := types.ListValueFrom(ctx, types.StringType, clientResp.CallbackURLs)
	resp.Diagnostics.Append(diags...)
	state.CallbackURLs = callbackURLs

	// Update logout callback URLs
	if len(clientResp.LogoutCallbackURLs) > 0 {
		logoutCallbackURLs, diags := types.ListValueFrom(ctx, types.StringType, clientResp.LogoutCallbackURLs)
		resp.Diagnostics.Append(diags...)
		state.LogoutCallbackURLs = logoutCallbackURLs
	} else {
		state.LogoutCallbackURLs = types.ListNull(types.StringType)
	}

	// Update allowed user groups
	if len(clientResp.AllowedUserGroups) > 0 {
		var groupIDs []string
		for _, group := range clientResp.AllowedUserGroups {
			groupIDs = append(groupIDs, group.ID)
		}
		allowedGroups, diags := types.ListValueFrom(ctx, types.StringType, groupIDs)
		resp.Diagnostics.Append(diags...)
		state.AllowedUserGroups = allowedGroups
	} else {
		state.AllowedUserGroups = types.ListNull(types.StringType)
	}

	// Note: client_secret is not updated from Read as it's only available during creation

	// Set the state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *clientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan clientResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	// Retrieve current state
	var state clientResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform types to Go types
	var callbackURLs []string
	diags = plan.CallbackURLs.ElementsAs(ctx, &callbackURLs, false)
	resp.Diagnostics.Append(diags...)

	var logoutCallbackURLs []string
	if !plan.LogoutCallbackURLs.IsNull() {
		diags = plan.LogoutCallbackURLs.ElementsAs(ctx, &logoutCallbackURLs, false)
		resp.Diagnostics.Append(diags...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Determine if group restriction is enabled based on allowed_user_groups
	var isGroupRestricted bool
	if !plan.AllowedUserGroups.IsNull() && !plan.AllowedUserGroups.IsUnknown() {
		var groupIDs []string
		_ = plan.AllowedUserGroups.ElementsAs(ctx, &groupIDs, false)
		isGroupRestricted = len(groupIDs) > 0
	}

	// Update the client
	updateReq := &client.OIDCClientCreateRequest{
		Name:                                plan.Name.ValueString(),
		CallbackURLs:                        callbackURLs,
		LogoutCallbackURLs:                  logoutCallbackURLs,
		IsPublic:                            plan.IsPublic.ValueBool(),
		RequiresReauthentication:            plan.RequiresReauthentication.ValueBool(),
		RequiresPushedAuthorizationRequests: plan.RequiresPushedAuthorizationRequests.ValueBool(),
		LaunchURL: func() *string {
			if !plan.LaunchURL.IsNull() && !plan.LaunchURL.IsUnknown() && plan.LaunchURL.ValueString() != "" {
				v := plan.LaunchURL.ValueString()
				return &v
			}
			return nil
		}(),
		PkceEnabled:       plan.PkceEnabled.ValueBool(),
		IsGroupRestricted: isGroupRestricted,
		Credentials:       buildCredentialsFromPlan(ctx, &plan),
	}
	if !plan.ClientID.IsNull() && !plan.ClientID.IsUnknown() && plan.ClientID.ValueString() != "" {
		cid := plan.ClientID.ValueString()
		updateReq.ClientID = &cid
	}

	tflog.Debug(ctx, "Updating OIDC client", map[string]any{
		"id":   plan.ID.ValueString(),
		"name": updateReq.Name,
	})

	clientResp, err := r.client.UpdateClient(plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating OIDC client",
			"Could not update OIDC client, unexpected error: "+err.Error(),
		)
		return
	}

	// Update state values
	plan.HasLogo = types.BoolValue(clientResp.HasLogo)
	plan.RequiresReauthentication = types.BoolValue(clientResp.RequiresReauthentication)
	plan.FederatedIdentities = federatedIdentitiesToList(ctx, clientResp.Credentials.FederatedIdentities)
	// Preserve the configured PAR value unless the server returns the field.
	if clientResp.RequiresPushedAuthorizationRequests != nil {
		plan.RequiresPushedAuthorizationRequests = types.BoolValue(*clientResp.RequiresPushedAuthorizationRequests)
	}
	if clientResp.LaunchURL != "" {
		plan.LaunchURL = types.StringValue(clientResp.LaunchURL)
	} else {
		plan.LaunchURL = types.StringNull()
	}

	// Handle allowed user groups
	var plannedGroupIDs []string
	if !plan.AllowedUserGroups.IsNull() && !plan.AllowedUserGroups.IsUnknown() {
		diags = plan.AllowedUserGroups.ElementsAs(ctx, &plannedGroupIDs, false)
		resp.Diagnostics.Append(diags...)
	}

	var currentGroupIDs []string
	if !state.AllowedUserGroups.IsNull() {
		diags = state.AllowedUserGroups.ElementsAs(ctx, &currentGroupIDs, false)
		resp.Diagnostics.Append(diags...)
	}

	if !resp.Diagnostics.HasError() {
		// Check if groups have changed
		groupsChanged := false
		if len(plannedGroupIDs) != len(currentGroupIDs) {
			groupsChanged = true
		} else {
			// Check if group IDs are different
			groupMap := make(map[string]bool)
			for _, id := range currentGroupIDs {
				groupMap[id] = true
			}
			for _, id := range plannedGroupIDs {
				if !groupMap[id] {
					groupsChanged = true
					break
				}
			}
		}

		if groupsChanged {
			tflog.Debug(ctx, "Updating allowed user groups", map[string]any{
				"groups": plannedGroupIDs,
			})
			err = r.client.UpdateClientAllowedUserGroups(plan.ID.ValueString(), plannedGroupIDs)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating allowed user groups",
					"Could not update allowed user groups: "+err.Error(),
				)
				return
			}
		}
	}

	// Preserve the client secret from state as it cannot be retrieved
	plan.ClientSecret = state.ClientSecret

	// Set the state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *clientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state clientResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting OIDC client", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Delete the client
	err := r.client.DeleteClient(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting OIDC client",
			"Could not delete OIDC client, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted OIDC client", map[string]any{
		"id": state.ID.ValueString(),
	})
}

// ImportState imports an existing resource into Terraform.
func (r *clientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and set it as the resource ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// urlValidator validates that a string is a valid URL
type urlValidator struct{}

func (v urlValidator) Description(ctx context.Context) string {
	return "string must be a valid URL"
}

func (v urlValidator) MarkdownDescription(ctx context.Context) string {
	return "string must be a valid URL"
}

func (v urlValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := strings.TrimSpace(req.ConfigValue.ValueString())

	// Reject empty strings after trimming
	if value == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"invalid callback URL",
			"Callback URL must not be empty",
		)
		return
	}

	// Allow wildcard patterns containing '*'
	if strings.Contains(value, "*") {
		return
	}

	// Parse and require a scheme and some content (host or path)
	u, err := url.Parse(value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"invalid callback URL",
			fmt.Sprintf("The value %q is not a valid URL: %s", value, err),
		)
		return
	}

	if u.Scheme == "" || (u.Host == "" && u.Path == "" && u.Opaque == "") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"invalid callback URL",
			fmt.Sprintf("The value %q is not a valid URL: must include a scheme and a host, path, or opaque data", value),
		)
		return
	}
}

// buildCreateRequestFromPlan converts the Terraform plan model into an API create request.
func buildCreateRequestFromPlan(ctx context.Context, plan *clientResourceModel) *client.OIDCClientCreateRequest {
	var callbackURLs []string
	_ = plan.CallbackURLs.ElementsAs(ctx, &callbackURLs, false)

	var logoutCallbackURLs []string
	if !plan.LogoutCallbackURLs.IsNull() {
		_ = plan.LogoutCallbackURLs.ElementsAs(ctx, &logoutCallbackURLs, false)
	}

	var launchPtr *string
	if !plan.LaunchURL.IsNull() && !plan.LaunchURL.IsUnknown() && plan.LaunchURL.ValueString() != "" {
		v := plan.LaunchURL.ValueString()
		launchPtr = &v
	}

	// Determine if group restriction is enabled based on allowed_user_groups
	var isGroupRestricted bool
	if !plan.AllowedUserGroups.IsNull() && !plan.AllowedUserGroups.IsUnknown() {
		var groupIDs []string
		_ = plan.AllowedUserGroups.ElementsAs(ctx, &groupIDs, false)
		isGroupRestricted = len(groupIDs) > 0
	}

	return &client.OIDCClientCreateRequest{
		Name:                                plan.Name.ValueString(),
		CallbackURLs:                        callbackURLs,
		LogoutCallbackURLs:                  logoutCallbackURLs,
		IsPublic:                            plan.IsPublic.ValueBool(),
		RequiresReauthentication:            plan.RequiresReauthentication.ValueBool(),
		RequiresPushedAuthorizationRequests: plan.RequiresPushedAuthorizationRequests.ValueBool(),
		LaunchURL:                           launchPtr,
		PkceEnabled:                         plan.PkceEnabled.ValueBool(),
		IsGroupRestricted:                   isGroupRestricted,
		Credentials:                         buildCredentialsFromPlan(ctx, plan),
	}
}

// buildCredentialsFromPlan converts the federated_identities plan list into API credentials.
func buildCredentialsFromPlan(ctx context.Context, plan *clientResourceModel) client.OIDCClientCredentials {
	if plan.FederatedIdentities.IsNull() || plan.FederatedIdentities.IsUnknown() {
		return client.OIDCClientCredentials{}
	}

	var identities []clientFederatedIdentityModel
	_ = plan.FederatedIdentities.ElementsAs(ctx, &identities, false)

	if len(identities) == 0 {
		return client.OIDCClientCredentials{}
	}

	federated := make([]client.OIDCClientFederatedIdentity, 0, len(identities))
	for _, identity := range identities {
		federated = append(federated, client.OIDCClientFederatedIdentity{
			Issuer:   identity.Issuer.ValueString(),
			Subject:  identity.Subject.ValueString(),
			Audience: identity.Audience.ValueString(),
			JWKS:     identity.JWKS.ValueString(),
		})
	}

	return client.OIDCClientCredentials{FederatedIdentities: federated}
}

// federatedIdentitiesToList converts API federated identities into a Terraform list value.
func federatedIdentitiesToList(ctx context.Context, identities []client.OIDCClientFederatedIdentity) types.List {
	objType := types.ObjectType{AttrTypes: federatedIdentityAttrTypes}
	if len(identities) == 0 {
		return types.ListNull(objType)
	}

	models := make([]clientFederatedIdentityModel, 0, len(identities))
	for _, identity := range identities {
		models = append(models, clientFederatedIdentityModel{
			Issuer:   types.StringValue(identity.Issuer),
			Subject:  optionalString(identity.Subject),
			Audience: optionalString(identity.Audience),
			JWKS:     optionalString(identity.JWKS),
		})
	}

	list, _ := types.ListValueFrom(ctx, objType, models)
	return list
}

// optionalString returns a null string value when the input is empty.
func optionalString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

// mapAPIClientToModel maps an API OIDCClient response into the Terraform resource model.
func mapAPIClientToModel(ctx context.Context, api *client.OIDCClient) clientResourceModel {
	var model clientResourceModel
	model.ID = types.StringValue(api.ID)
	model.Name = types.StringValue(api.Name)
	model.IsPublic = types.BoolValue(api.IsPublic)
	model.PkceEnabled = types.BoolValue(api.PkceEnabled)
	model.HasLogo = types.BoolValue(api.HasLogo)
	model.RequiresReauthentication = types.BoolValue(api.RequiresReauthentication)
	model.FederatedIdentities = federatedIdentitiesToList(ctx, api.Credentials.FederatedIdentities)

	callbackURLs, _ := types.ListValueFrom(ctx, types.StringType, api.CallbackURLs)
	model.CallbackURLs = callbackURLs

	if len(api.LogoutCallbackURLs) > 0 {
		logoutURLs, _ := types.ListValueFrom(ctx, types.StringType, api.LogoutCallbackURLs)
		model.LogoutCallbackURLs = logoutURLs
	} else {
		model.LogoutCallbackURLs = types.ListNull(types.StringType)
	}

	if len(api.AllowedUserGroups) > 0 {
		var groupIDs []string
		for _, g := range api.AllowedUserGroups {
			groupIDs = append(groupIDs, g.ID)
		}
		allowed, _ := types.ListValueFrom(ctx, types.StringType, groupIDs)
		model.AllowedUserGroups = allowed
	} else {
		model.AllowedUserGroups = types.ListNull(types.StringType)
	}

	if api.LaunchURL != "" {
		model.LaunchURL = types.StringValue(api.LaunchURL)
	} else {
		model.LaunchURL = types.StringNull()
	}

	return model
}
