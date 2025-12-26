package datasources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
	"github.com/Trozz/terraform-provider-pocketid/internal/datasources"
)

// Test Client Data Source
func TestClientDataSource_Metadata(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewClientDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_client", resp.TypeName)
}

func TestClientDataSource_Schema(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewClientDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, resp.Schema)
	assert.NotEmpty(t, resp.Schema.Description)

	// Verify attributes exist
	expectedAttributes := []string{
		"id", "name", "has_logo", "callback_urls", "logout_callback_urls",
		"is_public", "pkce_enabled", "allowed_user_groups", "requires_reauthentication", "launch_url",
	}

	for _, attr := range expectedAttributes {
		_, ok := resp.Schema.Attributes[attr]
		assert.True(t, ok, "Schema should have %s attribute", attr)
	}

	// Check attribute types
	idAttr, ok := resp.Schema.Attributes["id"].(schema.StringAttribute)
	assert.True(t, ok)
	assert.True(t, idAttr.Required)

	callbackUrlsAttr, ok := resp.Schema.Attributes["callback_urls"].(schema.ListAttribute)
	assert.True(t, ok)
	assert.True(t, callbackUrlsAttr.Computed)
}

func TestClientDataSource_Configure(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid_client",
			providerData: &client.Client{},
			expectError:  false,
		},
		{
			name:         "nil_provider_data",
			providerData: nil,
			expectError:  false,
		},
		{
			name:          "invalid_provider_data_type",
			providerData:  "invalid",
			expectError:   true,
			errorContains: "Expected *client.Client",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := datasources.NewClientDataSource()

			configurable, ok := ds.(datasource.DataSourceWithConfigure)
			require.True(t, ok)

			req := datasource.ConfigureRequest{
				ProviderData: tc.providerData,
			}
			resp := &datasource.ConfigureResponse{}

			configurable.Configure(ctx, req, resp)

			if tc.expectError {
				assert.True(t, resp.Diagnostics.HasError())
				assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.errorContains)
			} else {
				assert.False(t, resp.Diagnostics.HasError())
			}
		})
	}
}

// Test Clients Data Source
func TestClientsDataSource_Metadata(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewClientsDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_clients", resp.TypeName)
}

func TestClientsDataSource_Schema(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewClientsDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, resp.Schema)

	// Verify clients attribute exists
	clientsAttr, ok := resp.Schema.Attributes["clients"]
	assert.True(t, ok, "Schema should have clients attribute")

	listAttr, ok := clientsAttr.(schema.ListNestedAttribute)
	assert.True(t, ok, "clients should be a ListNestedAttribute")

	// Verify nested attributes
	expectedNestedAttributes := []string{
		"id", "name", "has_logo", "callback_urls", "logout_callback_urls",
		"is_public", "pkce_enabled", "allowed_user_groups", "requires_reauthentication", "launch_url",
	}

	for _, attr := range expectedNestedAttributes {
		_, ok := listAttr.NestedObject.Attributes[attr]
		assert.True(t, ok, "Nested object should have %s attribute", attr)
	}
}

func TestClientsDataSource_Configure(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid_client",
			providerData: &client.Client{},
			expectError:  false,
		},
		{
			name:         "nil_provider_data",
			providerData: nil,
			expectError:  false,
		},
		{
			name:          "invalid_provider_data_type",
			providerData:  123,
			expectError:   true,
			errorContains: "Expected *client.Client",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := datasources.NewClientsDataSource()

			configurable, ok := ds.(datasource.DataSourceWithConfigure)
			require.True(t, ok)

			req := datasource.ConfigureRequest{
				ProviderData: tc.providerData,
			}
			resp := &datasource.ConfigureResponse{}

			configurable.Configure(ctx, req, resp)

			if tc.expectError {
				assert.True(t, resp.Diagnostics.HasError())
				assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.errorContains)
			} else {
				assert.False(t, resp.Diagnostics.HasError())
			}
		})
	}
}

// Test User Data Source
func TestUserDataSource_Metadata(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewUserDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_user", resp.TypeName)
}

func TestUserDataSource_Schema(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewUserDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, resp.Schema)
	assert.NotEmpty(t, resp.Schema.Description)

	// Verify attributes exist
	expectedAttributes := []string{
		"id", "username", "email", "first_name", "last_name",
		"is_admin", "locale", "disabled", "groups",
	}

	for _, attr := range expectedAttributes {
		_, ok := resp.Schema.Attributes[attr]
		assert.True(t, ok, "Schema should have %s attribute", attr)
	}

	// Check specific attribute types
	idAttr, ok := resp.Schema.Attributes["id"].(schema.StringAttribute)
	assert.True(t, ok)
	assert.True(t, idAttr.Optional)
	assert.True(t, idAttr.Computed)

	usernameAttr, ok := resp.Schema.Attributes["username"].(schema.StringAttribute)
	assert.True(t, ok)
	assert.True(t, usernameAttr.Optional)
	assert.True(t, usernameAttr.Computed)

	groupsAttr, ok := resp.Schema.Attributes["groups"].(schema.SetAttribute)
	assert.True(t, ok, "groups should be a SetAttribute")
	assert.True(t, groupsAttr.Computed)
}

func TestUserDataSource_Configure(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid_client",
			providerData: &client.Client{},
			expectError:  false,
		},
		{
			name:         "nil_provider_data",
			providerData: nil,
			expectError:  false,
		},
		{
			name:          "invalid_provider_data_type",
			providerData:  []string{"invalid"},
			expectError:   true,
			errorContains: "Expected *client.Client",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := datasources.NewUserDataSource()

			configurable, ok := ds.(datasource.DataSourceWithConfigure)
			require.True(t, ok)

			req := datasource.ConfigureRequest{
				ProviderData: tc.providerData,
			}
			resp := &datasource.ConfigureResponse{}

			configurable.Configure(ctx, req, resp)

			if tc.expectError {
				assert.True(t, resp.Diagnostics.HasError())
				assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.errorContains)
			} else {
				assert.False(t, resp.Diagnostics.HasError())
			}
		})
	}
}

// Test Users Data Source
func TestUsersDataSource_Metadata(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewUsersDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_users", resp.TypeName)
}

func TestUsersDataSource_Schema(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewUsersDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, resp.Schema)

	// Verify users attribute
	usersAttr, ok := resp.Schema.Attributes["users"]
	assert.True(t, ok)

	listAttr, ok := usersAttr.(schema.ListNestedAttribute)
	assert.True(t, ok, "users should be a ListNestedAttribute")

	// Verify nested attributes
	expectedNestedAttributes := []string{
		"id", "username", "email", "first_name", "last_name",
		"is_admin", "locale", "disabled", "groups",
	}

	for _, attr := range expectedNestedAttributes {
		_, ok := listAttr.NestedObject.Attributes[attr]
		assert.True(t, ok, "Nested object should have %s attribute", attr)
	}
}

func TestUsersDataSource_Configure(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid_client",
			providerData: &client.Client{},
			expectError:  false,
		},
		{
			name:         "nil_provider_data",
			providerData: nil,
			expectError:  false,
		},
		{
			name:          "invalid_provider_data_type",
			providerData:  map[string]string{"invalid": "type"},
			expectError:   true,
			errorContains: "Expected *client.Client",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := datasources.NewUsersDataSource()

			configurable, ok := ds.(datasource.DataSourceWithConfigure)
			require.True(t, ok)

			req := datasource.ConfigureRequest{
				ProviderData: tc.providerData,
			}
			resp := &datasource.ConfigureResponse{}

			configurable.Configure(ctx, req, resp)

			if tc.expectError {
				assert.True(t, resp.Diagnostics.HasError())
				assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.errorContains)
			} else {
				assert.False(t, resp.Diagnostics.HasError())
			}
		})
	}
}

// Test Group Data Source
func TestGroupDataSource_Metadata(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewGroupDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_group", resp.TypeName)
}

func TestGroupDataSource_Schema(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewGroupDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, resp.Schema)
	assert.NotEmpty(t, resp.Schema.Description)

	// Verify attributes exist
	expectedAttributes := []string{
		"id", "name", "friendly_name",
	}

	for _, attr := range expectedAttributes {
		_, ok := resp.Schema.Attributes[attr]
		assert.True(t, ok, "Schema should have %s attribute", attr)
	}

	// Check specific attribute types and properties
	idAttr, ok := resp.Schema.Attributes["id"].(schema.StringAttribute)
	assert.True(t, ok)
	assert.True(t, idAttr.Optional)
	assert.True(t, idAttr.Computed)

	nameAttr, ok := resp.Schema.Attributes["name"].(schema.StringAttribute)
	assert.True(t, ok)
	assert.True(t, nameAttr.Optional)
	assert.True(t, nameAttr.Computed)

	friendlyNameAttr, ok := resp.Schema.Attributes["friendly_name"].(schema.StringAttribute)
	assert.True(t, ok)
	assert.True(t, friendlyNameAttr.Computed)
}

func TestGroupDataSource_Configure(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid_client",
			providerData: &client.Client{},
			expectError:  false,
		},
		{
			name:         "nil_provider_data",
			providerData: nil,
			expectError:  false,
		},
		{
			name:          "invalid_provider_data_type",
			providerData:  struct{ Name string }{Name: "invalid"},
			expectError:   true,
			errorContains: "Expected *client.Client",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := datasources.NewGroupDataSource()

			configurable, ok := ds.(datasource.DataSourceWithConfigure)
			require.True(t, ok)

			req := datasource.ConfigureRequest{
				ProviderData: tc.providerData,
			}
			resp := &datasource.ConfigureResponse{}

			configurable.Configure(ctx, req, resp)

			if tc.expectError {
				assert.True(t, resp.Diagnostics.HasError())
				assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.errorContains)
			} else {
				assert.False(t, resp.Diagnostics.HasError())
			}
		})
	}
}

// Test Groups Data Source
func TestGroupsDataSource_Metadata(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewGroupsDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_groups", resp.TypeName)
}

func TestGroupsDataSource_Schema(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewGroupsDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, resp.Schema)
	assert.NotEmpty(t, resp.Schema.Description)

	// Verify groups attribute
	groupsAttr, ok := resp.Schema.Attributes["groups"]
	assert.True(t, ok, "Schema should have groups attribute")

	listAttr, ok := groupsAttr.(schema.ListNestedAttribute)
	assert.True(t, ok, "groups should be a ListNestedAttribute")
	assert.True(t, listAttr.Computed)

	// Verify nested attributes
	expectedNestedAttributes := []string{
		"id", "name", "friendly_name",
	}

	for _, attr := range expectedNestedAttributes {
		_, ok := listAttr.NestedObject.Attributes[attr]
		assert.True(t, ok, "Nested object should have %s attribute", attr)
	}

	// Check nested attribute types
	idAttr, ok := listAttr.NestedObject.Attributes["id"].(schema.StringAttribute)
	assert.True(t, ok)
	assert.True(t, idAttr.Computed)

	nameAttr, ok := listAttr.NestedObject.Attributes["name"].(schema.StringAttribute)
	assert.True(t, ok)
	assert.True(t, nameAttr.Computed)

	friendlyNameAttr, ok := listAttr.NestedObject.Attributes["friendly_name"].(schema.StringAttribute)
	assert.True(t, ok)
	assert.True(t, friendlyNameAttr.Computed)
}

func TestGroupsDataSource_Configure(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid_client",
			providerData: &client.Client{},
			expectError:  false,
		},
		{
			name:         "nil_provider_data",
			providerData: nil,
			expectError:  false,
		},
		{
			name:          "invalid_provider_data_type",
			providerData:  42,
			expectError:   true,
			errorContains: "Expected *client.Client",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := datasources.NewGroupsDataSource()

			configurable, ok := ds.(datasource.DataSourceWithConfigure)
			require.True(t, ok)

			req := datasource.ConfigureRequest{
				ProviderData: tc.providerData,
			}
			resp := &datasource.ConfigureResponse{}

			configurable.Configure(ctx, req, resp)

			if tc.expectError {
				assert.True(t, resp.Diagnostics.HasError())
				assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.errorContains)
			} else {
				assert.False(t, resp.Diagnostics.HasError())
			}
		})
	}
}

// Test that all data sources have descriptions
func TestDataSources_HaveDescriptions(t *testing.T) {
	ctx := context.Background()

	dataSources := []datasource.DataSource{
		datasources.NewClientDataSource(),
		datasources.NewClientsDataSource(),
		datasources.NewUserDataSource(),
		datasources.NewUsersDataSource(),
		datasources.NewGroupDataSource(),
		datasources.NewGroupsDataSource(),
	}

	for _, ds := range dataSources {
		t.Run(getDataSourceName(t, ds), func(t *testing.T) {
			req := datasource.SchemaRequest{}
			resp := &datasource.SchemaResponse{}

			ds.Schema(ctx, req, resp)

			require.False(t, resp.Diagnostics.HasError())
			assert.NotEmpty(t, resp.Schema.Description, "Data source should have a description")
		})
	}
}

// Helper function to get data source name
func getDataSourceName(t *testing.T, ds datasource.DataSource) string {
	ctx := context.Background()
	req := datasource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &datasource.MetadataResponse{}
	ds.Metadata(ctx, req, resp)
	return resp.TypeName
}
