package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

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

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &OneTimeAccessTokenResource{}
var _ resource.ResourceWithImportState = &OneTimeAccessTokenResource{}

func NewOneTimeAccessTokenResource() resource.Resource {
	return &OneTimeAccessTokenResource{}
}

// OneTimeAccessTokenResource defines the resource implementation
type OneTimeAccessTokenResource struct {
	client *client.Client
}

// OneTimeAccessTokenResourceModel describes the resource data model
type OneTimeAccessTokenResourceModel struct {
	ID           types.String `tfsdk:"id"`
	UserID       types.String `tfsdk:"user_id"`
	Token        types.String `tfsdk:"token"`
	ExpiresAt    types.String `tfsdk:"expires_at"`
	CreatedAt    types.String `tfsdk:"created_at"`
	SkipRecreate types.Bool   `tfsdk:"skip_recreate"`
}

func (r *OneTimeAccessTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_one_time_access_token"
}

func (r *OneTimeAccessTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a one-time access token for a user in Pocket-ID. These tokens allow users to authenticate when they don't have access to their passkey.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the one-time access token (same as user_id)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user this token belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The one-time access token value",
				Computed:            true,
				Sensitive:           true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "The expiration time of the token in RFC3339 format",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The creation time of the token in RFC3339 format",
				Computed:            true,
			},
			"skip_recreate": schema.BoolAttribute{
				MarkdownDescription: "If true, the resource will not be recreated when the token is not found (used or expired). This is useful for initial user setup where the token is sent via another provider.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}

func (r *OneTimeAccessTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OneTimeAccessTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OneTimeAccessTokenResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare the request
	tokenReq := &client.OneTimeAccessTokenRequest{}

	// Parse expires_at (now required)
	expiresAt, err := time.Parse(time.RFC3339, data.ExpiresAt.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid expires_at format",
			fmt.Sprintf("The expires_at value must be in RFC3339 format: %s", err),
		)
		return
	}
	tokenReq.ExpiresAt = &expiresAt

	// Create the token
	tflog.Trace(ctx, "creating one-time access token", map[string]interface{}{
		"user_id": data.UserID.ValueString(),
	})

	token, err := r.client.CreateOneTimeAccessToken(data.UserID.ValueString(), tokenReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating one-time access token",
			fmt.Sprintf("Could not create one-time access token for user %s: %s", data.UserID.ValueString(), err),
		)
		return
	}

	// Map response to model
	data.ID = data.UserID
	data.Token = types.StringValue(token.Token)
	data.ExpiresAt = types.StringValue(token.ExpiresAt.Format(time.RFC3339))
	data.CreatedAt = types.StringValue(token.CreatedAt.Format(time.RFC3339))

	tflog.Trace(ctx, "created one-time access token", map[string]interface{}{
		"user_id": data.UserID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OneTimeAccessTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OneTimeAccessTokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current token state
	tflog.Trace(ctx, "reading one-time access token", map[string]interface{}{
		"user_id": data.UserID.ValueString(),
	})

	token, err := r.client.GetOneTimeAccessToken(data.UserID.ValueString())
	if err != nil {
		// If the token is not found
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			// Check if we should skip recreation
			// Default to true if not set
			skipRecreate := true
			if !data.SkipRecreate.IsNull() && !data.SkipRecreate.IsUnknown() {
				skipRecreate = data.SkipRecreate.ValueBool()
			}

			if skipRecreate {
				// Keep the resource in state but clear sensitive values
				// This prevents Terraform from trying to recreate it
				tflog.Debug(ctx, "Token not found but skip_recreate is true, maintaining resource in state")

				// Clear the token value since it's no longer valid
				data.Token = types.StringValue("")

				// Save the updated state
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}

			// Otherwise, remove it from state to trigger recreation
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading one-time access token",
			fmt.Sprintf("Could not read one-time access token for user %s: %s", data.UserID.ValueString(), err),
		)
		return
	}

	// Update model with read data
	data.Token = types.StringValue(token.Token)
	data.ExpiresAt = types.StringValue(token.ExpiresAt.Format(time.RFC3339))
	data.CreatedAt = types.StringValue(token.CreatedAt.Format(time.RFC3339))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OneTimeAccessTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// One-time access tokens cannot be updated
	resp.Diagnostics.AddError(
		"Update not supported",
		"One-time access tokens cannot be updated. To change a token, delete and recreate it.",
	)
}

func (r *OneTimeAccessTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OneTimeAccessTokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the token
	tflog.Trace(ctx, "deleting one-time access token", map[string]interface{}{
		"user_id": data.UserID.ValueString(),
	})

	err := r.client.DeleteOneTimeAccessToken(data.UserID.ValueString())
	if err != nil {
		// If the token is already gone, consider it deleted
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return
		}

		resp.Diagnostics.AddError(
			"Error deleting one-time access token",
			fmt.Sprintf("Could not delete one-time access token for user %s: %s", data.UserID.ValueString(), err),
		)
		return
	}

	tflog.Trace(ctx, "deleted one-time access token", map[string]interface{}{
		"user_id": data.UserID.ValueString(),
	})
}

func (r *OneTimeAccessTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by user ID
	resource.ImportStatePassthroughID(ctx, path.Root("user_id"), req, resp)

	// Also set the ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	// Set skip_recreate to true by default on import
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("skip_recreate"), true)...)
}
