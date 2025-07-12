package resources

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                = &clientResource{}
	_ resource.ResourceWithConfigure   = &clientResource{}
	_ resource.ResourceWithImportState = &clientResource{}
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
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	CallbackURLs       types.List   `tfsdk:"callback_urls"`
	LogoutCallbackURLs types.List   `tfsdk:"logout_callback_urls"`
	IsPublic           types.Bool   `tfsdk:"is_public"`
	PkceEnabled        types.Bool   `tfsdk:"pkce_enabled"`
	AllowedUserGroups  types.List   `tfsdk:"allowed_user_groups"`
	HasLogo            types.Bool   `tfsdk:"has_logo"`
	ClientSecret       types.String `tfsdk:"client_secret"`
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

	// Create the client
	createReq := &client.OIDCClientCreateRequest{
		Name:               plan.Name.ValueString(),
		CallbackURLs:       callbackURLs,
		LogoutCallbackURLs: logoutCallbackURLs,
		IsPublic:           plan.IsPublic.ValueBool(),
		PkceEnabled:        plan.PkceEnabled.ValueBool(),
		Credentials:        client.OIDCClientCredentials{}, // Empty for now
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

	// Set state values
	plan.ID = types.StringValue(clientResp.ID)
	plan.HasLogo = types.BoolValue(clientResp.HasLogo)

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

	// Update the client
	updateReq := &client.OIDCClientCreateRequest{
		Name:               plan.Name.ValueString(),
		CallbackURLs:       callbackURLs,
		LogoutCallbackURLs: logoutCallbackURLs,
		IsPublic:           plan.IsPublic.ValueBool(),
		PkceEnabled:        plan.PkceEnabled.ValueBool(),
		Credentials:        client.OIDCClientCredentials{}, // Empty for now
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

	value := req.ConfigValue.ValueString()
	_, err := url.Parse(value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"invalid callback URL",
			fmt.Sprintf("The value %q is not a valid URL: %s", value, err),
		)
		return
	}

	// Additional validation - must have scheme and host
	u, _ := url.Parse(value)
	if u.Scheme == "" || u.Host == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"invalid callback URL",
			fmt.Sprintf("The value %q is not a valid URL: must include scheme and host", value),
		)
	}
}
