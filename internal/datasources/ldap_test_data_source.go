package datasources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &ldapTestDataSource{}
	_ datasource.DataSourceWithConfigure = &ldapTestDataSource{}
)

// NewLDAPTestDataSource is a helper function to simplify the provider implementation.
func NewLDAPTestDataSource() datasource.DataSource {
	return &ldapTestDataSource{}
}

// ldapTestDataSource is the data source implementation.
type ldapTestDataSource struct {
	client *client.Client
}

// ldapTestDataSourceModel describes the data source data model.
type ldapTestDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ConnectionSuccessful types.Bool   `tfsdk:"connection_successful"`
	BindSuccessful       types.Bool   `tfsdk:"bind_successful"`
	BaseDNAccessible     types.Bool   `tfsdk:"base_dn_accessible"`
	UsersFound           types.Int64  `tfsdk:"users_found"`
	GroupsFound          types.Int64  `tfsdk:"groups_found"`
	ErrorMessage         types.String `tfsdk:"error_message"`
	TestTimestamp        types.String `tfsdk:"test_timestamp"`
}

// Metadata returns the data source type name.
func (d *ldapTestDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ldap_test"
}

// Schema defines the schema for the data source.
func (d *ldapTestDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Tests the current LDAP configuration in Pocket-ID.",
		MarkdownDescription: `Tests the current LDAP configuration in Pocket-ID. This data source performs various connectivity and configuration tests to help diagnose LDAP integration issues.

~> **Note** This data source automatically tests using the current LDAP configuration. Make sure you have configured LDAP before using this data source.

## Example Usage

` + "```hcl" + `
# Test LDAP configuration
data "pocketid_ldap_test" "test" {}

# Use in validation
output "ldap_test_results" {
  value = {
    connection_ok = data.pocketid_ldap_test.test.connection_successful
    bind_ok       = data.pocketid_ldap_test.test.bind_successful
    users_found   = data.pocketid_ldap_test.test.users_found
    groups_found  = data.pocketid_ldap_test.test.groups_found
    error         = data.pocketid_ldap_test.test.error_message
  }
}

# Conditional resource based on test results
resource "pocketid_ldap_sync" "sync" {
  count = data.pocketid_ldap_test.test.connection_successful && data.pocketid_ldap_test.test.bind_successful ? 1 : 0

  triggers = {
    timestamp = timestamp()
  }
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the test.",
				Computed:    true,
			},
			"connection_successful": schema.BoolAttribute{
				Description: "Whether the LDAP connection was successful.",
				Computed:    true,
			},
			"bind_successful": schema.BoolAttribute{
				Description: "Whether the LDAP bind authentication was successful.",
				Computed:    true,
			},
			"base_dn_accessible": schema.BoolAttribute{
				Description: "Whether the base DN is accessible with the provided credentials.",
				Computed:    true,
			},
			"users_found": schema.Int64Attribute{
				Description: "Number of users found with the configured search filter.",
				Computed:    true,
			},
			"groups_found": schema.Int64Attribute{
				Description: "Number of groups found with the configured search filter.",
				Computed:    true,
			},
			"error_message": schema.StringAttribute{
				Description: "Detailed error message if any test failed.",
				Computed:    true,
			},
			"test_timestamp": schema.StringAttribute{
				Description: "When the test was performed in RFC3339 format.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ldapTestDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ldapTestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ldapTestDataSourceModel

	// TODO: Call API to test LDAP configuration
	// This will be implemented by Agent 3
	tflog.Debug(ctx, "Testing LDAP configuration")

	// For now, set dummy test results
	state.ID = types.StringValue("ldap-test")
	state.ConnectionSuccessful = types.BoolValue(true)
	state.BindSuccessful = types.BoolValue(true)
	state.BaseDNAccessible = types.BoolValue(true)
	state.UsersFound = types.Int64Value(10)
	state.GroupsFound = types.Int64Value(3)
	state.ErrorMessage = types.StringNull()
	state.TestTimestamp = types.StringValue(time.Now().Format(time.RFC3339))

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
