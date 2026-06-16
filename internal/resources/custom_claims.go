package resources

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

// customClaimsToAPI converts a Terraform map of custom claims into the API's
// list representation. A null or unknown map yields an empty slice so the API
// performs a full replace that clears any existing claims.
func customClaimsToAPI(ctx context.Context, claims types.Map) ([]client.CustomClaim, diag.Diagnostics) {
	var diags diag.Diagnostics
	if claims.IsNull() || claims.IsUnknown() {
		return []client.CustomClaim{}, diags
	}

	values := make(map[string]string, len(claims.Elements()))
	diags = claims.ElementsAs(ctx, &values, false)
	if diags.HasError() {
		return nil, diags
	}

	// Sort keys for deterministic request ordering.
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]client.CustomClaim, 0, len(keys))
	for _, key := range keys {
		result = append(result, client.CustomClaim{Key: key, Value: values[key]})
	}
	return result, diags
}

// customClaimsToState converts the API's list of custom claims into a Terraform
// map. An empty list is mapped to a null map to avoid perpetual diffs when the
// configuration omits the attribute.
func customClaimsToState(ctx context.Context, claims []client.CustomClaim) (types.Map, diag.Diagnostics) {
	if len(claims) == 0 {
		return types.MapNull(types.StringType), diag.Diagnostics{}
	}

	values := make(map[string]string, len(claims))
	for _, claim := range claims {
		values[claim.Key] = claim.Value
	}
	return types.MapValueFrom(ctx, types.StringType, values)
}
