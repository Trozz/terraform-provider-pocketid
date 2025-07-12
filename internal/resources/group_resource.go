package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &groupResource{}
	_ resource.ResourceWithConfigure   = &groupResource{}
	_ resource.ResourceWithImportState = &groupResource{}
)

// NewGroupResource is a helper function to simplify the provider implementation.
func NewGroupResource() resource.Resource {
	return &groupResource{}
}

// groupResource is the resource implementation.
type groupResource struct {
	client *client.Client
}

// groupResourceModel maps the resource schema data.
type groupResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	FriendlyName types.String `tfsdk:"friendly_name"`
}

// Metadata returns the resource type name.
func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the schema for the resource.
func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a user group in Pocket-ID.",
		MarkdownDescription: "Manages a user group in Pocket-ID. Groups can be used to organize users and control access to OIDC clients.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the user group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description:         "The unique name identifier of the user group. This is used as the technical identifier.",
				MarkdownDescription: "The unique name identifier of the user group. This is used as the technical identifier and will be included in tokens.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"friendly_name": schema.StringAttribute{
				Description: "The friendly display name of the user group.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 50),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan groupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the group
	createReq := &client.UserGroupCreateRequest{
		Name:         plan.Name.ValueString(),
		FriendlyName: plan.FriendlyName.ValueString(),
	}

	tflog.Debug(ctx, "Creating user group", map[string]any{
		"name":         createReq.Name,
		"friendlyName": createReq.FriendlyName,
	})

	groupResp, err := r.client.CreateUserGroup(createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating user group",
			"Could not create user group, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Created user group", map[string]any{
		"id": groupResp.ID,
	})

	// Set state values
	plan.ID = types.StringValue(groupResp.ID)

	// Set the state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state groupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading user group", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Get group from API
	groupResp, err := r.client.GetUserGroup(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading user group",
			"Could not read user group ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state from API response
	state.Name = types.StringValue(groupResp.Name)
	state.FriendlyName = types.StringValue(groupResp.FriendlyName)

	// Set the state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan groupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update the group
	updateReq := &client.UserGroupCreateRequest{
		Name:         plan.Name.ValueString(),
		FriendlyName: plan.FriendlyName.ValueString(),
	}

	tflog.Debug(ctx, "Updating user group", map[string]any{
		"id":           plan.ID.ValueString(),
		"name":         updateReq.Name,
		"friendlyName": updateReq.FriendlyName,
	})

	_, err := r.client.UpdateUserGroup(plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating user group",
			"Could not update user group, unexpected error: "+err.Error(),
		)
		return
	}

	// Set the state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state groupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting user group", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Delete the group
	err := r.client.DeleteUserGroup(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting user group",
			"Could not delete user group, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted user group", map[string]any{
		"id": state.ID.ValueString(),
	})
}

// ImportState imports an existing resource into Terraform.
func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and set it as the resource ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
