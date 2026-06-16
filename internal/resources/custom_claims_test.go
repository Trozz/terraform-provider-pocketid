package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

func TestCustomClaimsToAPI(t *testing.T) {
	ctx := context.Background()

	t.Run("null map returns empty slice", func(t *testing.T) {
		claims, diags := customClaimsToAPI(ctx, types.MapNull(types.StringType))
		require.False(t, diags.HasError())
		assert.Empty(t, claims)
	})

	t.Run("unknown map returns empty slice", func(t *testing.T) {
		claims, diags := customClaimsToAPI(ctx, types.MapUnknown(types.StringType))
		require.False(t, diags.HasError())
		assert.Empty(t, claims)
	})

	t.Run("populated map returns sorted slice", func(t *testing.T) {
		m, d := types.MapValueFrom(ctx, types.StringType, map[string]string{
			"level":      "senior",
			"department": "engineering",
		})
		require.False(t, d.HasError())

		claims, diags := customClaimsToAPI(ctx, m)
		require.False(t, diags.HasError())
		assert.Equal(t, []client.CustomClaim{
			{Key: "department", Value: "engineering"},
			{Key: "level", Value: "senior"},
		}, claims)
	})
}

func TestCustomClaimsToState(t *testing.T) {
	ctx := context.Background()

	t.Run("empty slice maps to null", func(t *testing.T) {
		m, diags := customClaimsToState(ctx, nil)
		require.False(t, diags.HasError())
		assert.True(t, m.IsNull())
	})

	t.Run("populated slice maps to map", func(t *testing.T) {
		m, diags := customClaimsToState(ctx, []client.CustomClaim{
			{Key: "department", Value: "engineering"},
			{Key: "level", Value: "senior"},
		})
		require.False(t, diags.HasError())
		require.False(t, m.IsNull())

		var values map[string]string
		diags = m.ElementsAs(ctx, &values, false)
		require.False(t, diags.HasError())
		assert.Equal(t, map[string]string{
			"department": "engineering",
			"level":      "senior",
		}, values)
	})

	t.Run("round trip preserves values", func(t *testing.T) {
		original := []client.CustomClaim{
			{Key: "a", Value: "1"},
			{Key: "b", Value: "2"},
		}
		m, diags := customClaimsToState(ctx, original)
		require.False(t, diags.HasError())

		roundTripped, diags := customClaimsToAPI(ctx, m)
		require.False(t, diags.HasError())
		assert.ElementsMatch(t, original, roundTripped)
	})
}
