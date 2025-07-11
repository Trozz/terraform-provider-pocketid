package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &groupDataSource{}
	_ datasource.DataSourceWithConfigure = &groupDataSource{}
)

// NewGroupDataSource creates a new group data source.
func NewGroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

// groupDataSource is the data source implementation.
type groupDataSource struct {
	client *client.Client
}

// groupDataSourceModel describes the data source data model.
type groupDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	FriendlyName types.String `tfsdk:"friendly_name"`
}

// Metadata returns the data source type name.
func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the schema for the data source.
func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a Pocket-ID group.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the group. Either id or name must be provided.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The unique name identifier of the group. Either id or name must be provided.",
				Optional:    true,
				Computed:    true,
			},
			"friendly_name": schema.StringAttribute{
				Description: "The friendly display name of the group.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *groupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data groupDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either ID or name is provided
	if data.ID.IsNull() && data.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Missing required argument",
			"Either 'id' or 'name' must be provided",
		)
		return
	}

	// Get all groups
	groupsResp, err := d.client.ListUserGroups()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Groups",
			err.Error(),
		)
		return
	}

	// Find the matching group
	var foundGroup *client.UserGroup
	for _, group := range groupsResp.Data {
		if (!data.ID.IsNull() && group.ID == data.ID.ValueString()) ||
			(!data.Name.IsNull() && group.Name == data.Name.ValueString()) {
			foundGroup = &group
			break
		}
	}

	if foundGroup == nil {
		searchField := "id"
		searchValue := data.ID.ValueString()
		if data.ID.IsNull() {
			searchField = "name"
			searchValue = data.Name.ValueString()
		}
		resp.Diagnostics.AddError(
			"Group Not Found",
			fmt.Sprintf("No group found with %s '%s'", searchField, searchValue),
		)
		return
	}

	tflog.Debug(ctx, "Found group", map[string]interface{}{
		"id":            foundGroup.ID,
		"name":          foundGroup.Name,
		"friendly_name": foundGroup.FriendlyName,
	})

	// Map response body to model
	data.ID = types.StringValue(foundGroup.ID)
	data.Name = types.StringValue(foundGroup.Name)
	data.FriendlyName = types.StringValue(foundGroup.FriendlyName)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
