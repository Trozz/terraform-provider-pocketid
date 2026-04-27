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
