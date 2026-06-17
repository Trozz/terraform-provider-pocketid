package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource              = &ldapSyncResource{}
	_ resource.ResourceWithConfigure = &ldapSyncResource{}
)

// NewLdapSyncResource is a helper function to simplify the provider implementation.
func NewLdapSyncResource() resource.Resource {
	return &ldapSyncResource{}
}

// ldapSyncResource defines the resource implementation.
type ldapSyncResource struct {
	client *client.Client
}

// ldapSyncResourceModel maps the resource schema data.
type ldapSyncResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Triggers types.Map    `tfsdk:"triggers"`
	SyncedAt types.String `tfsdk:"synced_at"`
}

func (r *ldapSyncResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ldap_sync"
}

func (r *ldapSyncResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Triggers an LDAP synchronization in Pocket-ID. This is an action resource: applying it " +
			"runs a sync, and changing `triggers` forces a new sync (the resource is recreated). LDAP must be enabled " +
			"in the application configuration, otherwise the sync — and the apply — fails.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the sync resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"triggers": schema.MapAttribute{
				MarkdownDescription: "Arbitrary map of values that forces a new LDAP sync when it changes. Typically " +
					"wired to the LDAP configuration values so a sync runs whenever they change. If omitted, the sync " +
					"runs only once (on create).",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"synced_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp (RFC3339) of the most recent sync triggered by this resource.",
				Computed:            true,
			},
		},
	}
}

func (r *ldapSyncResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *ldapSyncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ldapSyncResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "triggering LDAP sync")
	if err := r.client.SyncLdap(); err != nil {
		resp.Diagnostics.AddError(
			"Error syncing LDAP",
			"Could not trigger LDAP sync: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue("ldap-sync")
	plan.SyncedAt = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ldapSyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// A sync is an action with no readable server-side state; preserve prior state.
	var data ldapSyncResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ldapSyncResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All configurable attributes force replacement, so Update is never expected.
	resp.Diagnostics.AddError(
		"Update not supported",
		"pocketid_ldap_sync cannot be updated in place. Change the triggers map to run a new sync.",
	)
}

func (r *ldapSyncResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// A sync cannot be undone; there is nothing to delete server-side.
	tflog.Trace(ctx, "removing pocketid_ldap_sync from state (no server-side action)")
}
