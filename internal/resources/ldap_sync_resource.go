package resources

import (
	"context"
	"fmt"
	"time"

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
	_ resource.Resource              = &ldapSyncResource{}
	_ resource.ResourceWithConfigure = &ldapSyncResource{}
)

// NewLDAPSyncResource is a helper function to simplify the provider implementation.
func NewLDAPSyncResource() resource.Resource {
	return &ldapSyncResource{}
}

// ldapSyncResource is the resource implementation.
type ldapSyncResource struct {
	client *client.Client
}

// ldapSyncResourceModel describes the resource data model.
type ldapSyncResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Triggers types.Map    `tfsdk:"triggers"`
	LastSync types.String `tfsdk:"last_sync"`
	Status   types.String `tfsdk:"status"`
	Error    types.String `tfsdk:"error"`
}

// Metadata returns the resource type name.
func (r *ldapSyncResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ldap_sync"
}

// Schema defines the schema for the resource.
func (r *ldapSyncResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers an LDAP synchronization in Pocket-ID.",
		MarkdownDescription: `Triggers an LDAP synchronization in Pocket-ID. This resource forces a sync operation whenever it is created or updated.

~> **Note** This resource is useful for ensuring LDAP data is synchronized after configuration changes. Use the triggers attribute to control when syncs occur.

## Example Usage

### Force sync on every apply
` + "```hcl" + `
resource "pocketid_ldap_sync" "sync" {
  triggers = {
    timestamp = timestamp()
  }
}
` + "```" + `

### Sync when configuration changes
` + "```hcl" + `
resource "pocketid_ldap_sync" "sync" {
  triggers = {
    config_hash = sha256(jsonencode(pocketid_ldap_config.main))
  }
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the sync operation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"triggers": schema.MapAttribute{
				Description: "Map of values that, when changed, will trigger a new sync. Common pattern: Use timestamp() to force sync on every apply.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"last_sync": schema.StringAttribute{
				Description: "Timestamp of last successful sync in RFC3339 format.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of last sync: 'success', 'failed', or 'in_progress'.",
				Computed:    true,
			},
			"error": schema.StringAttribute{
				Description: "Error message if sync failed.",
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
	// Retrieve values from plan
	var plan ldapSyncResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate a unique ID for this sync operation
	plan.ID = types.StringValue(fmt.Sprintf("ldap-sync-%d", time.Now().Unix()))

	// Call API to trigger LDAP sync
	tflog.Debug(ctx, "Triggering LDAP sync")

	syncResp, err := r.client.TriggerLDAPSyncWithContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Triggering LDAP Sync",
			fmt.Sprintf("Could not trigger LDAP sync: %s", err),
		)
		return
	}

	// Update state with sync response
	plan.LastSync = types.StringValue(time.Now().Format(time.RFC3339))
	plan.Status = types.StringValue(syncResp.Status)

	// Set error message if sync failed
	if syncResp.Status == "failed" && syncResp.Error != "" {
		plan.Error = types.StringValue(syncResp.Error)
	} else {
		plan.Error = types.StringNull()
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ldapSyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ldapSyncResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// For sync resource, we don't need to fetch status from API
	// The state is maintained locally as the sync is a one-time operation
	tflog.Debug(ctx, "Reading LDAP sync status from state")

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ldapSyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan ldapSyncResourceModel
	var state ldapSyncResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keep the same ID
	plan.ID = state.ID

	// Call API to trigger a new LDAP sync
	tflog.Debug(ctx, "Re-triggering LDAP sync due to update")

	syncResp, err := r.client.TriggerLDAPSyncWithContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Triggering LDAP Sync",
			fmt.Sprintf("Could not trigger LDAP sync: %s", err),
		)
		return
	}

	// Update sync metadata
	plan.LastSync = types.StringValue(time.Now().Format(time.RFC3339))
	plan.Status = types.StringValue(syncResp.Status)

	// Set error message if sync failed
	if syncResp.Status == "failed" && syncResp.Error != "" {
		plan.Error = types.StringValue(syncResp.Error)
	} else {
		plan.Error = types.StringNull()
	}

	// Update the state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ldapSyncResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state ldapSyncResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No-op for delete - syncs cannot be "undone"
	tflog.Debug(ctx, "Deleting LDAP sync resource (no-op)")
}
