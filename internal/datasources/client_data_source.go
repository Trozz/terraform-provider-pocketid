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
	_ datasource.DataSource              = &clientDataSource{}
	_ datasource.DataSourceWithConfigure = &clientDataSource{}
)

// NewClientDataSource is a helper function to simplify the provider implementation.
func NewClientDataSource() datasource.DataSource {
	return &clientDataSource{}
}

// clientDataSource is the data source implementation.
type clientDataSource struct {
	client *client.Client
}

// clientDataSourceModel maps the data source schema data.
type clientDataSourceModel struct {
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
func (d *clientDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

// Schema defines the schema for the data source.
func (d *clientDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches an OIDC client from Pocket-ID.",
		MarkdownDescription: "Fetches an OIDC client from Pocket-ID by its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The ID of the OIDC client to fetch.",
				MarkdownDescription: "The ID of the OIDC client to fetch.",
				Required:            true,
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
	}
}

// Configure adds the provider configured client to the data source.
func (d *clientDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *clientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current configuration
	var config clientDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading OIDC client data source", map[string]any{
		"id": config.ID.ValueString(),
	})

	// Get client from API
	clientResp, err := d.client.GetClient(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading OIDC client",
			"Could not read OIDC client ID "+config.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to model
	state := clientDataSourceModel{
		ID:          types.StringValue(clientResp.ID),
		Name:        types.StringValue(clientResp.Name),
		IsPublic:    types.BoolValue(clientResp.IsPublic),
		PkceEnabled: types.BoolValue(clientResp.PkceEnabled),
		HasLogo:     types.BoolValue(clientResp.HasLogo),
	}

	// Map callback URLs
	callbackURLs, diags := types.ListValueFrom(ctx, types.StringType, clientResp.CallbackURLs)
	resp.Diagnostics.Append(diags...)
	state.CallbackURLs = callbackURLs

	// Map logout callback URLs
	if len(clientResp.LogoutCallbackURLs) > 0 {
		logoutCallbackURLs, diags := types.ListValueFrom(ctx, types.StringType, clientResp.LogoutCallbackURLs)
		resp.Diagnostics.Append(diags...)
		state.LogoutCallbackURLs = logoutCallbackURLs
	} else {
		state.LogoutCallbackURLs = types.ListNull(types.StringType)
	}

	// Map allowed user groups
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

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
