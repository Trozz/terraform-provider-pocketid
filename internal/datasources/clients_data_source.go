package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/trozz/terraform-provider-pocketid/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &clientsDataSource{}
	_ datasource.DataSourceWithConfigure = &clientsDataSource{}
)

// NewClientsDataSource is a helper function to simplify the provider implementation.
func NewClientsDataSource() datasource.DataSource {
	return &clientsDataSource{}
}

// clientsDataSource is the data source implementation.
type clientsDataSource struct {
	client *client.Client
}

// clientsDataSourceModel maps the data source schema data.
type clientsDataSourceModel struct {
	Clients []clientModel `tfsdk:"clients"`
}

// clientModel represents a single client in the list
type clientModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	CallbackURLs       types.List   `tfsdk:"callback_urls"`
	LogoutCallbackURLs types.List   `tfsdk:"logout_callback_urls"`
	IsPublic           types.Bool   `tfsdk:"is_public"`
	PkceEnabled        types.Bool   `tfsdk:"pkce_enabled"`
	AllowedUserGroups  types.List   `tfsdk:"allowed_user_groups"`
	HasLogo            types.Bool   `tfsdk:"has_logo"`
}

// Metadata returns the data source type name.
func (d *clientsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clients"
}

// Schema defines the schema for the data source.
func (d *clientsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches all OIDC clients from Pocket-ID.",
		MarkdownDescription: "Fetches all OIDC clients from Pocket-ID.",
		Attributes: map[string]schema.Attribute{
			"clients": schema.ListNestedAttribute{
				Description: "List of all OIDC clients.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the OIDC client.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The display name of the OIDC client.",
							Computed:    true,
						},
						"callback_urls": schema.ListAttribute{
							Description: "List of allowed callback URLs for the OIDC client.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"logout_callback_urls": schema.ListAttribute{
							Description: "List of allowed logout callback URLs for the OIDC client.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"is_public": schema.BoolAttribute{
							Description: "Whether this is a public client (no client secret).",
							Computed:    true,
						},
						"pkce_enabled": schema.BoolAttribute{
							Description: "Whether PKCE is enabled for this client.",
							Computed:    true,
						},
						"allowed_user_groups": schema.ListAttribute{
							Description: "List of user group IDs that are allowed to use this client.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"has_logo": schema.BoolAttribute{
							Description: "Whether the client has a logo configured.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *clientsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *clientsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading OIDC clients data source")

	// Get clients from API
	clientsResp, err := d.client.ListClients()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading OIDC clients",
			"Could not read OIDC clients: "+err.Error(),
		)
		return
	}

	// Map response to model
	state := clientsDataSourceModel{
		Clients: make([]clientModel, 0, len(clientsResp.Data)),
	}

	// Convert each client
	for _, clientResp := range clientsResp.Data {
		clientState := clientModel{
			ID:          types.StringValue(clientResp.ID),
			Name:        types.StringValue(clientResp.Name),
			IsPublic:    types.BoolValue(clientResp.IsPublic),
			PkceEnabled: types.BoolValue(clientResp.PkceEnabled),
			HasLogo:     types.BoolValue(clientResp.HasLogo),
		}

		// Map callback URLs
		callbackURLs, diags := types.ListValueFrom(ctx, types.StringType, clientResp.CallbackURLs)
		resp.Diagnostics.Append(diags...)
		clientState.CallbackURLs = callbackURLs

		// Map logout callback URLs
		if len(clientResp.LogoutCallbackURLs) > 0 {
			logoutCallbackURLs, diags := types.ListValueFrom(ctx, types.StringType, clientResp.LogoutCallbackURLs)
			resp.Diagnostics.Append(diags...)
			clientState.LogoutCallbackURLs = logoutCallbackURLs
		} else {
			clientState.LogoutCallbackURLs = types.ListNull(types.StringType)
		}

		// Map allowed user groups
		if len(clientResp.AllowedUserGroups) > 0 {
			var groupIDs []string
			for _, group := range clientResp.AllowedUserGroups {
				groupIDs = append(groupIDs, group.ID)
			}
			allowedGroups, diags := types.ListValueFrom(ctx, types.StringType, groupIDs)
			resp.Diagnostics.Append(diags...)
			clientState.AllowedUserGroups = allowedGroups
		} else {
			clientState.AllowedUserGroups = types.ListNull(types.StringType)
		}

		state.Clients = append(state.Clients, clientState)
	}

	tflog.Debug(ctx, "Found OIDC clients", map[string]any{
		"count": len(state.Clients),
	})

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
