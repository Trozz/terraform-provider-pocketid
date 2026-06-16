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

// maxOneTimeAccessTokenTTL mirrors the pocket-id API limit (31 days).
const maxOneTimeAccessTokenTTL = 31 * 24 * time.Hour

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
	ID        types.String `tfsdk:"id"`
	UserID    types.String `tfsdk:"user_id"`
	TTL       types.String `tfsdk:"ttl"`
	Token     types.String `tfsdk:"token"`
	ExpiresAt types.String `tfsdk:"expires_at"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *OneTimeAccessTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_one_time_access_token"
}

func (r *OneTimeAccessTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a one-time access token for a user in Pocket-ID. These tokens let a user authenticate when they don't have access to their passkey. " +
			"The token value is returned only once on creation and cannot be read back (pocket-id exposes no read endpoint), so it is stored in Terraform state as a sensitive value.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the one-time access token (same as user_id).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user this token belongs to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ttl": schema.StringAttribute{
				MarkdownDescription: "Lifetime of the token expressed as a Go duration string (e.g. `15m`, `1h`, `24h`). " +
					"Must be greater than 1 second and at most 744h (31 days). Changing this forces a new token to be created.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The one-time access token value. Returned only on creation.",
				Computed:            true,
				Sensitive:           true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "The computed expiration time of the token in RFC3339 format (created_at + ttl).",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The creation time of the token in RFC3339 format.",
				Computed:            true,
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

	// Validate the ttl duration up front to give a clear error before calling the API.
	ttlStr := data.TTL.ValueString()
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("ttl"),
			"Invalid ttl",
			fmt.Sprintf("The ttl value must be a Go duration string such as \"15m\" or \"1h\": %s", err),
		)
		return
	}
	if ttl <= time.Second || ttl > maxOneTimeAccessTokenTTL {
		resp.Diagnostics.AddAttributeError(
			path.Root("ttl"),
			"Invalid ttl",
			"The ttl must be greater than 1 second and at most 744h (31 days).",
		)
		return
	}

	tflog.Debug(ctx, "creating one-time access token", map[string]interface{}{
		"user_id": data.UserID.ValueString(),
		"ttl":     ttlStr,
	})

	token, err := r.client.CreateOneTimeAccessToken(data.UserID.ValueString(), &client.OneTimeAccessTokenRequest{TTL: ttlStr})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating one-time access token",
			fmt.Sprintf("Could not create one-time access token for user %s: %s", data.UserID.ValueString(), err),
		)
		return
	}

	// The API only returns the token value, so derive the remaining attributes locally.
	created := time.Now().UTC()
	data.ID = data.UserID
	data.Token = types.StringValue(token.Token)
	data.CreatedAt = types.StringValue(created.Format(time.RFC3339))
	data.ExpiresAt = types.StringValue(created.Add(ttl).Format(time.RFC3339))

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

	// pocket-id exposes no endpoint to read a one-time access token back, and the
	// token is consumed on use. There is nothing to refresh, so the prior state is
	// preserved as-is.
	tflog.Trace(ctx, "one-time access token is write-only, preserving state", map[string]interface{}{
		"user_id": data.UserID.ValueString(),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OneTimeAccessTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All configurable attributes force replacement, so Update is never expected to run.
	resp.Diagnostics.AddError(
		"Update not supported",
		"One-time access tokens cannot be updated. To change a token, delete and recreate it.",
	)
}

func (r *OneTimeAccessTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// pocket-id exposes no endpoint to revoke a one-time access token; it remains
	// valid until used or expired. Removing it from Terraform state is all we can do.
	tflog.Trace(ctx, "one-time access token cannot be revoked via API, removing from state only")
}

func (r *OneTimeAccessTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by user ID. The token value cannot be recovered from the API.
	resource.ImportStatePassthroughID(ctx, path.Root("user_id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
