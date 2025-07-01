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
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

// NewUserDataSource is a helper function to simplify the provider implementation.
func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

// userDataSource is the data source implementation.
type userDataSource struct {
	client *client.Client
}

// userDataSourceModel maps the data source schema data.
type userDataSourceModel struct {
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

// Metadata returns the data source type name.
func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the data source.
func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches a user from Pocket-ID.",
		MarkdownDescription: "Fetches a user from Pocket-ID by their ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The ID of the user to fetch.",
				MarkdownDescription: "The ID of the user to fetch.",
				Required:            true,
			},
			"username": schema.StringAttribute{
				Description: "The username of the user.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email address of the user.",
				Computed:    true,
			},
			"first_name": schema.StringAttribute{
				Description: "The first name of the user.",
				Computed:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "The last name of the user.",
				Computed:    true,
			},
			"is_admin": schema.BoolAttribute{
				Description: "Whether the user has administrator privileges.",
				Computed:    true,
			},
			"locale": schema.StringAttribute{
				Description: "The locale preference for the user.",
				Computed:    true,
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the user account is disabled.",
				Computed:    true,
			},
			"groups": schema.ListAttribute{
				Description: "List of group IDs the user belongs to.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current configuration
	var config userDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading user data source", map[string]any{
		"id": config.ID.ValueString(),
	})

	// Get user from API
	userResp, err := d.client.GetUser(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading user",
			"Could not read user ID "+config.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to model
	state := userDataSourceModel{
		ID:        types.StringValue(userResp.ID),
		Username:  types.StringValue(userResp.Username),
		Email:     types.StringValue(userResp.Email),
		FirstName: types.StringValue(userResp.FirstName),
		LastName:  types.StringValue(userResp.LastName),
		IsAdmin:   types.BoolValue(userResp.IsAdmin),
		Disabled:  types.BoolValue(userResp.Disabled),
	}

	// Handle locale
	if userResp.Locale != nil && *userResp.Locale != "" {
		state.Locale = types.StringValue(*userResp.Locale)
	} else {
		state.Locale = types.StringNull()
	}

	// Map groups
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

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
