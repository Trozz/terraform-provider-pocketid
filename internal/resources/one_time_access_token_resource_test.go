package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Trozz/terraform-provider-pocketid/internal/resources"
)

func TestNewOneTimeAccessTokenResource(t *testing.T) {
	r := resources.NewOneTimeAccessTokenResource()
	assert.NotNil(t, r)
	assert.Implements(t, (*resource.Resource)(nil), r)
}

func TestOneTimeAccessTokenResource_Metadata(t *testing.T) {
	r := resources.NewOneTimeAccessTokenResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.TODO(), req, resp)

	assert.Equal(t, "pocketid_one_time_access_token", resp.TypeName)
}

func TestOneTimeAccessTokenResource_Schema(t *testing.T) {
	r := resources.NewOneTimeAccessTokenResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.TODO(), req, resp)

	// Verify schema attributes
	schema := resp.Schema
	assert.NotNil(t, schema)

	// Check attributes
	assert.Contains(t, schema.Attributes, "id")
	assert.Contains(t, schema.Attributes, "user_id")
	assert.Contains(t, schema.Attributes, "ttl")
	assert.Contains(t, schema.Attributes, "token")
	assert.Contains(t, schema.Attributes, "expires_at")
	assert.Contains(t, schema.Attributes, "created_at")
	assert.NotContains(t, schema.Attributes, "skip_recreate")

	// Verify attribute properties
	idAttr := schema.Attributes["id"]
	assert.True(t, idAttr.IsComputed())
	assert.False(t, idAttr.IsRequired())

	userIDAttr := schema.Attributes["user_id"]
	assert.True(t, userIDAttr.IsRequired())
	assert.False(t, userIDAttr.IsComputed())

	ttlAttr := schema.Attributes["ttl"]
	assert.True(t, ttlAttr.IsRequired())
	assert.False(t, ttlAttr.IsComputed())

	tokenAttr := schema.Attributes["token"]
	assert.True(t, tokenAttr.IsComputed())
	assert.True(t, tokenAttr.IsSensitive())

	expiresAtAttr := schema.Attributes["expires_at"]
	assert.True(t, expiresAtAttr.IsComputed())
	assert.False(t, expiresAtAttr.IsRequired())

	createdAtAttr := schema.Attributes["created_at"]
	assert.True(t, createdAtAttr.IsComputed())
	assert.False(t, createdAtAttr.IsRequired())
}

func TestOneTimeAccessTokenResource_Configure(t *testing.T) {
	tests := []struct {
		name         string
		providerData interface{}
		expectError  bool
	}{
		{
			name:         "nil provider data",
			providerData: nil,
			expectError:  false,
		},
		{
			name:         "invalid provider data type",
			providerData: "invalid",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := resources.NewOneTimeAccessTokenResource()

			// Cast to the concrete type that implements Configure
			tokenResource, ok := r.(*resources.OneTimeAccessTokenResource)
			assert.True(t, ok, "Resource should be of type *OneTimeAccessTokenResource")

			req := resource.ConfigureRequest{
				ProviderData: tt.providerData,
			}
			resp := &resource.ConfigureResponse{}

			tokenResource.Configure(context.TODO(), req, resp)

			if tt.expectError {
				assert.True(t, resp.Diagnostics.HasError())
			} else {
				assert.False(t, resp.Diagnostics.HasError())
			}
		})
	}
}

func samplePriorTokenState() resources.OneTimeAccessTokenResourceModel {
	return resources.OneTimeAccessTokenResourceModel{
		ID:        types.StringValue("user-123"),
		UserID:    types.StringValue("user-123"),
		TTL:       types.StringValue("15m"),
		Token:     types.StringValue("ABC123"),
		ExpiresAt: types.StringValue("2026-01-01T00:15:00Z"),
		CreatedAt: types.StringValue("2026-01-01T00:00:00Z"),
	}
}

// One-time access tokens are write-only (pocket-id exposes no read endpoint),
// so Read must preserve prior state verbatim without contacting the API.
func TestOneTimeAccessTokenResource_Read_PreservesState(t *testing.T) {
	ctx := context.Background()
	r, ok := resources.NewOneTimeAccessTokenResource().(*resources.OneTimeAccessTokenResource)
	require.True(t, ok)

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	prior := samplePriorTokenState()
	state := tfsdk.State{Schema: schemaResp.Schema}
	require.False(t, state.Set(ctx, &prior).HasError())

	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Read(ctx, resource.ReadRequest{State: state}, resp)

	require.False(t, resp.Diagnostics.HasError())
	require.False(t, resp.State.Raw.IsNull(), "resource should be preserved in state")

	var got resources.OneTimeAccessTokenResourceModel
	require.False(t, resp.State.Get(ctx, &got).HasError())
	assert.Equal(t, "user-123", got.UserID.ValueString())
	assert.Equal(t, "15m", got.TTL.ValueString())
	assert.Equal(t, "ABC123", got.Token.ValueString(), "token should be preserved")
	assert.Equal(t, "2026-01-01T00:15:00Z", got.ExpiresAt.ValueString())
}

// Delete has no API endpoint to call, so it must succeed without error.
func TestOneTimeAccessTokenResource_Delete_NoOp(t *testing.T) {
	ctx := context.Background()
	r, ok := resources.NewOneTimeAccessTokenResource().(*resources.OneTimeAccessTokenResource)
	require.True(t, ok)

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	prior := samplePriorTokenState()
	state := tfsdk.State{Schema: schemaResp.Schema}
	require.False(t, state.Set(ctx, &prior).HasError())

	resp := &resource.DeleteResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Delete(ctx, resource.DeleteRequest{State: state}, resp)
	assert.False(t, resp.Diagnostics.HasError())
}
