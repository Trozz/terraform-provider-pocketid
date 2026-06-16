package resources_test

import (
	"context"
	"net/http"
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

	// Check required attributes
	assert.Contains(t, schema.Attributes, "id")
	assert.Contains(t, schema.Attributes, "user_id")
	assert.Contains(t, schema.Attributes, "token")
	assert.Contains(t, schema.Attributes, "expires_at")
	assert.Contains(t, schema.Attributes, "created_at")

	// Verify attribute properties
	idAttr := schema.Attributes["id"]
	assert.True(t, idAttr.IsComputed())
	assert.False(t, idAttr.IsRequired())

	userIDAttr := schema.Attributes["user_id"]
	assert.True(t, userIDAttr.IsRequired())
	assert.False(t, userIDAttr.IsComputed())

	tokenAttr := schema.Attributes["token"]
	assert.True(t, tokenAttr.IsComputed())
	assert.True(t, tokenAttr.IsSensitive())

	expiresAtAttr := schema.Attributes["expires_at"]
	assert.True(t, expiresAtAttr.IsRequired())
	assert.False(t, expiresAtAttr.IsComputed())

	createdAtAttr := schema.Attributes["created_at"]
	assert.True(t, createdAtAttr.IsComputed())
	assert.False(t, createdAtAttr.IsRequired())

	// Check skip_recreate attribute
	skipRecreateAttr := schema.Attributes["skip_recreate"]
	assert.True(t, skipRecreateAttr.IsOptional())
	assert.False(t, skipRecreateAttr.IsRequired())
	assert.True(t, skipRecreateAttr.IsComputed())
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

// readWithGetHandler configures a one-time access token resource against a mock
// server, seeds prior state, and invokes Read. It returns the Read response so
// tests can assert on the resulting state.
func readWithGetHandler(t *testing.T, prior resources.OneTimeAccessTokenResourceModel, handler http.HandlerFunc) *resource.ReadResponse {
	t.Helper()
	ctx := context.Background()

	testClient := createMockServer(t, handler)
	r, ok := resources.NewOneTimeAccessTokenResource().(*resources.OneTimeAccessTokenResource)
	require.True(t, ok)

	configResp := &resource.ConfigureResponse{}
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: testClient}, configResp)
	require.False(t, configResp.Diagnostics.HasError())

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	state := tfsdk.State{Schema: schemaResp.Schema}
	require.False(t, state.Set(ctx, &prior).HasError())

	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Read(ctx, resource.ReadRequest{State: state}, resp)
	return resp
}

func samplePriorTokenState(skipRecreate bool) resources.OneTimeAccessTokenResourceModel {
	return resources.OneTimeAccessTokenResourceModel{
		ID:           types.StringValue("user-123"),
		UserID:       types.StringValue("user-123"),
		Token:        types.StringValue("ABC123"),
		ExpiresAt:    types.StringValue("2026-12-01T00:00:00Z"),
		CreatedAt:    types.StringValue("2026-01-01T00:00:00Z"),
		SkipRecreate: types.BoolValue(skipRecreate),
	}
}

// pocket-id v2 removed the GET endpoint and responds with "API endpoint not
// found". Read must preserve the resource (including its token) rather than
// removing it from state.
func TestOneTimeAccessTokenResource_Read_EndpointNotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"API endpoint not found"}`))
	}

	resp := readWithGetHandler(t, samplePriorTokenState(false), handler)
	require.False(t, resp.Diagnostics.HasError())
	require.False(t, resp.State.Raw.IsNull(), "resource should be preserved in state")

	var got resources.OneTimeAccessTokenResourceModel
	require.False(t, resp.State.Get(context.Background(), &got).HasError())
	assert.Equal(t, "user-123", got.UserID.ValueString())
	assert.Equal(t, "ABC123", got.Token.ValueString(), "token should be preserved")
}

// When the token is genuinely gone and skip_recreate is true, Read keeps the
// resource in state but clears the now-invalid token value.
func TestOneTimeAccessTokenResource_Read_TokenNotFound_SkipRecreate(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"One-time access token not found"}`))
	}

	resp := readWithGetHandler(t, samplePriorTokenState(true), handler)
	require.False(t, resp.Diagnostics.HasError())
	require.False(t, resp.State.Raw.IsNull(), "resource should be preserved in state")

	var got resources.OneTimeAccessTokenResourceModel
	require.False(t, resp.State.Get(context.Background(), &got).HasError())
	assert.Equal(t, "user-123", got.UserID.ValueString())
	assert.Equal(t, "", got.Token.ValueString(), "token should be cleared")
}

// When the token is gone and skip_recreate is false, Read removes the resource
// from state so Terraform recreates it.
func TestOneTimeAccessTokenResource_Read_TokenNotFound_NoSkipRecreate(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"One-time access token not found"}`))
	}

	resp := readWithGetHandler(t, samplePriorTokenState(false), handler)
	require.False(t, resp.Diagnostics.HasError())
	assert.True(t, resp.State.Raw.IsNull(), "resource should be removed from state")
}

// When the token still exists, Read succeeds and retains the existing state.
func TestOneTimeAccessTokenResource_Read_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	resp := readWithGetHandler(t, samplePriorTokenState(false), handler)
	require.False(t, resp.Diagnostics.HasError())
	require.False(t, resp.State.Raw.IsNull())

	var got resources.OneTimeAccessTokenResourceModel
	require.False(t, resp.State.Get(context.Background(), &got).HasError())
	assert.Equal(t, "ABC123", got.Token.ValueString())
}
