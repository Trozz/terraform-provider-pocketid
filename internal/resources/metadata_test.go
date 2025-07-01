package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
	"github.com/Trozz/terraform-provider-pocketid/internal/resources"
)

func TestNewClientResource(t *testing.T) {
	r := resources.NewClientResource()
	assert.NotNil(t, r)
}

func TestNewUserResource(t *testing.T) {
	r := resources.NewUserResource()
	assert.NotNil(t, r)
}

func TestNewGroupResource(t *testing.T) {
	r := resources.NewGroupResource()
	assert.NotNil(t, r)
}

func TestClientResource_Metadata(t *testing.T) {
	ctx := context.Background()
	r := resources.NewClientResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_client", resp.TypeName)
}

func TestUserResource_Metadata(t *testing.T) {
	ctx := context.Background()
	r := resources.NewUserResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_user", resp.TypeName)
}

func TestGroupResource_Metadata(t *testing.T) {
	ctx := context.Background()
	r := resources.NewGroupResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_group", resp.TypeName)
}

func TestClientResource_Configure(t *testing.T) {
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
		{
			name:          "invalid_provider_data_int",
			providerData:  123,
			expectError:   true,
			errorContains: "Expected *client.Client",
		},
		{
			name:          "invalid_provider_data_bool",
			providerData:  true,
			expectError:   true,
			errorContains: "Expected *client.Client",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := resources.NewClientResource()

			configurable, ok := r.(resource.ResourceWithConfigure)
			require.True(t, ok, "Client resource should implement ResourceWithConfigure")

			req := resource.ConfigureRequest{
				ProviderData: tc.providerData,
			}
			resp := &resource.ConfigureResponse{}

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

func TestUserResource_Configure(t *testing.T) {
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
		{
			name:          "invalid_provider_data_map",
			providerData:  map[string]string{"key": "value"},
			expectError:   true,
			errorContains: "Expected *client.Client",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := resources.NewUserResource()

			configurable, ok := r.(resource.ResourceWithConfigure)
			require.True(t, ok, "User resource should implement ResourceWithConfigure")

			req := resource.ConfigureRequest{
				ProviderData: tc.providerData,
			}
			resp := &resource.ConfigureResponse{}

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

func TestGroupResource_Configure(t *testing.T) {
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
			providerData:  struct{}{},
			expectError:   true,
			errorContains: "Expected *client.Client",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := resources.NewGroupResource()

			configurable, ok := r.(resource.ResourceWithConfigure)
			require.True(t, ok, "Group resource should implement ResourceWithConfigure")

			req := resource.ConfigureRequest{
				ProviderData: tc.providerData,
			}
			resp := &resource.ConfigureResponse{}

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

// Test that all resources implement required interfaces
func TestResources_Interfaces(t *testing.T) {
	resourceFactories := []struct {
		name    string
		factory func() resource.Resource
	}{
		{"ClientResource", resources.NewClientResource},
		{"UserResource", resources.NewUserResource},
		{"GroupResource", resources.NewGroupResource},
	}

	for _, rf := range resourceFactories {
		t.Run(rf.name, func(t *testing.T) {
			res := rf.factory()

			// Check implements Resource interface
			_, ok := res.(resource.Resource)
			assert.True(t, ok, "%s should implement resource.Resource", rf.name)

			// Check implements ResourceWithConfigure
			_, ok = res.(resource.ResourceWithConfigure)
			assert.True(t, ok, "%s should implement resource.ResourceWithConfigure", rf.name)

			// Check implements ResourceWithImportState
			_, ok = res.(resource.ResourceWithImportState)
			assert.True(t, ok, "%s should implement resource.ResourceWithImportState", rf.name)
		})
	}
}
