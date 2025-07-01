package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

// NewUserResource is a helper function to simplify the provider implementation.
func NewUserResource() resource.Resource {
	return &userResource{}
}

// userResource is the resource implementation.
type userResource struct {
	client *client.Client
}

// userResourceModel maps the resource schema data.
type userResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	Email     types.String `tfsdk:"email"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	IsAdmin   types.Bool   `tfsdk:"is_admin"`
	Locale    types.String `tfsdk:"locale"`
	Disabled  types.Bool   `tfsdk:"disabled"`
	Groups    types.List   `tfsdk:"groups"`
}

// Metadata returns the resource type name.
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the resource.
func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user in Pocket-ID.",
		MarkdownDescription: `Manages a user in Pocket-ID.

~> **Important** Users must complete passkey registration through the Pocket-ID web interface. This resource only creates the user account; authentication setup must be done separately.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "The username for the user. Must be unique.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email address of the user.",
				Required:    true,
			},
			"first_name": schema.StringAttribute{
				Description: "The first name of the user.",
				Optional:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "The last name of the user.",
				Optional:    true,
			},
			"is_admin": schema.BoolAttribute{
				Description: "Whether the user has administrator privileges. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"locale": schema.StringAttribute{
				Description: "The locale preference for the user (e.g., 'en', 'fr').",
				Optional:    true,
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the user account is disabled. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"groups": schema.ListAttribute{
				Description: "List of group IDs the user belongs to.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the user
	createReq := &client.UserCreateRequest{
		Username:  plan.Username.ValueString(),
		Email:     plan.Email.ValueString(),
		FirstName: plan.FirstName.ValueString(),
		LastName:  plan.LastName.ValueString(),
		IsAdmin:   plan.IsAdmin.ValueBool(),
		Disabled:  plan.Disabled.ValueBool(),
	}

	// Handle locale if provided
	if !plan.Locale.IsNull() {
		locale := plan.Locale.ValueString()
		createReq.Locale = &locale
	}

	tflog.Debug(ctx, "Creating user", map[string]any{
		"username": createReq.Username,
		"email":    createReq.Email,
		"isAdmin":  createReq.IsAdmin,
	})

	userResp, err := r.client.CreateUser(createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating user",
			"Could not create user, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Created user", map[string]any{
		"id": userResp.ID,
	})

	// Set state values
	plan.ID = types.StringValue(userResp.ID)

	// Handle user groups
	if !plan.Groups.IsNull() && !plan.Groups.IsUnknown() {
		var groupIDs []string
		diags = plan.Groups.ElementsAs(ctx, &groupIDs, false)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() && len(groupIDs) > 0 {
			tflog.Debug(ctx, "Updating user groups", map[string]any{
				"groups": groupIDs,
			})
			err = r.client.UpdateUserGroups(userResp.ID, groupIDs)
			if err != nil {
				// Try to clean up the created user
				_ = r.client.DeleteUser(userResp.ID)
				resp.Diagnostics.AddError(
					"Error updating user groups",
					"Could not update user groups, the user was deleted. Error: "+err.Error(),
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
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading user", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Get user from API
	userResp, err := r.client.GetUser(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading user",
			"Could not read user ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state from API response
	state.Username = types.StringValue(userResp.Username)
	state.Email = types.StringValue(userResp.Email)
	state.FirstName = types.StringValue(userResp.FirstName)
	state.LastName = types.StringValue(userResp.LastName)
	state.IsAdmin = types.BoolValue(userResp.IsAdmin)
	state.Disabled = types.BoolValue(userResp.Disabled)

	// Handle locale
	if userResp.Locale != nil && *userResp.Locale != "" {
		state.Locale = types.StringValue(*userResp.Locale)
	} else {
		state.Locale = types.StringNull()
	}

	// Update groups
	if len(userResp.UserGroups) > 0 {
		var groupIDs []string
		for _, group := range userResp.UserGroups {
			groupIDs = append(groupIDs, group.ID)
		}
		groups, diags := types.ListValueFrom(ctx, types.StringType, groupIDs)
		resp.Diagnostics.Append(diags...)
		state.Groups = groups
	} else {
		state.Groups = types.ListNull(types.StringType)
	}

	// Set the state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	// Retrieve current state
	var state userResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update the user
	updateReq := &client.UserCreateRequest{
		Username:  plan.Username.ValueString(),
		Email:     plan.Email.ValueString(),
		FirstName: plan.FirstName.ValueString(),
		LastName:  plan.LastName.ValueString(),
		IsAdmin:   plan.IsAdmin.ValueBool(),
		Disabled:  plan.Disabled.ValueBool(),
	}

	// Handle locale if provided
	if !plan.Locale.IsNull() {
		locale := plan.Locale.ValueString()
		updateReq.Locale = &locale
	}

	tflog.Debug(ctx, "Updating user", map[string]any{
		"id":       plan.ID.ValueString(),
		"username": updateReq.Username,
		"email":    updateReq.Email,
	})

	_, err := r.client.UpdateUser(plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating user",
			"Could not update user, unexpected error: "+err.Error(),
		)
		return
	}

	// Handle user groups
	var plannedGroupIDs []string
	if !plan.Groups.IsNull() && !plan.Groups.IsUnknown() {
		diags = plan.Groups.ElementsAs(ctx, &plannedGroupIDs, false)
		resp.Diagnostics.Append(diags...)
	}

	var currentGroupIDs []string
	if !state.Groups.IsNull() {
		diags = state.Groups.ElementsAs(ctx, &currentGroupIDs, false)
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
			tflog.Debug(ctx, "Updating user groups", map[string]any{
				"groups": plannedGroupIDs,
			})
			err = r.client.UpdateUserGroups(plan.ID.ValueString(), plannedGroupIDs)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating user groups",
					"Could not update user groups: "+err.Error(),
				)
				return
			}
		}
	}

	// Set the state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting user", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Delete the user
	err := r.client.DeleteUser(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting user",
			"Could not delete user, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted user", map[string]any{
		"id": state.ID.ValueString(),
	})
}

// ImportState imports an existing resource into Terraform.
func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and set it as the resource ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
