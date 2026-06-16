package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

func TestBuildCreateRequestFromPlan(t *testing.T) {
	ctx := context.Background()

	cb, diags := types.ListValueFrom(ctx, types.StringType, []string{"https://example.com/callback"})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}

	launch := "https://example.com/launch"

	fedList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: federatedIdentityAttrTypes}, []clientFederatedIdentityModel{
		{
			Issuer:   types.StringValue("https://issuer.example.com"),
			Subject:  types.StringValue("subject-1"),
			Audience: types.StringValue("audience-1"),
			JWKS:     types.StringNull(),
		},
	})
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}

	plan := &clientResourceModel{
		Name:                     types.StringValue("Test Client"),
		CallbackURLs:             cb,
		IsPublic:                 types.BoolValue(false),
		PkceEnabled:              types.BoolValue(true),
		RequiresReauthentication: types.BoolValue(true),
		LaunchURL:                types.StringValue(launch),
		FederatedIdentities:      fedList,
	}

	req := buildCreateRequestFromPlan(ctx, plan)

	if req == nil {
		t.Fatal("expected request, got nil")
	}
	assert.Equal(t, "Test Client", req.Name)
	assert.Equal(t, []string{"https://example.com/callback"}, req.CallbackURLs)
	assert.True(t, req.RequiresReauthentication)
	if assert.NotNil(t, req.LaunchURL) {
		assert.Equal(t, launch, *req.LaunchURL)
	}
	if assert.Len(t, req.Credentials.FederatedIdentities, 1) {
		assert.Equal(t, "https://issuer.example.com", req.Credentials.FederatedIdentities[0].Issuer)
		assert.Equal(t, "subject-1", req.Credentials.FederatedIdentities[0].Subject)
		assert.Equal(t, "audience-1", req.Credentials.FederatedIdentities[0].Audience)
		assert.Equal(t, "", req.Credentials.FederatedIdentities[0].JWKS)
	}
}

func TestMapAPIClientToModel(t *testing.T) {
	ctx := context.Background()

	api := &client.OIDCClient{
		ID:                       "client-1",
		Name:                     "API Client",
		CallbackURLs:             []string{"https://example.com/callback"},
		LogoutCallbackURLs:       []string{"https://example.com/logout"},
		IsPublic:                 false,
		PkceEnabled:              true,
		HasLogo:                  false,
		RequiresReauthentication: true,
		LaunchURL:                "https://example.com/launch",
		AllowedUserGroups:        []client.UserGroup{{ID: "g1"}},
		Credentials: client.OIDCClientCredentials{
			FederatedIdentities: []client.OIDCClientFederatedIdentity{
				{Issuer: "https://issuer.example.com", Subject: "subject-1"},
			},
		},
	}

	model := mapAPIClientToModel(ctx, api)

	assert.Equal(t, "client-1", model.ID.ValueString())
	assert.Equal(t, "API Client", model.Name.ValueString())
	assert.Equal(t, true, model.RequiresReauthentication.ValueBool())
	assert.Equal(t, "https://example.com/launch", model.LaunchURL.ValueString())

	// Callback URLs
	var cb []string
	_ = model.CallbackURLs.ElementsAs(ctx, &cb, false)
	assert.Equal(t, []string{"https://example.com/callback"}, cb)

	// Allowed groups
	var gids []string
	_ = model.AllowedUserGroups.ElementsAs(ctx, &gids, false)
	assert.Equal(t, []string{"g1"}, gids)

	// Federated identities
	assert.False(t, model.FederatedIdentities.IsNull())
	var fed []clientFederatedIdentityModel
	_ = model.FederatedIdentities.ElementsAs(ctx, &fed, false)
	if assert.Len(t, fed, 1) {
		assert.Equal(t, "https://issuer.example.com", fed[0].Issuer.ValueString())
		assert.Equal(t, "subject-1", fed[0].Subject.ValueString())
		assert.True(t, fed[0].Audience.IsNull())
		assert.True(t, fed[0].JWKS.IsNull())
	}
}

func TestMapAPIClientToModelNoFederatedIdentities(t *testing.T) {
	ctx := context.Background()

	api := &client.OIDCClient{
		ID:           "client-2",
		Name:         "No Fed",
		CallbackURLs: []string{"https://example.com/callback"},
	}

	model := mapAPIClientToModel(ctx, api)

	assert.True(t, model.FederatedIdentities.IsNull())
}
