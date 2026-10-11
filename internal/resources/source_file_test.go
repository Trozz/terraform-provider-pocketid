package resources

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, data, 0o600))
	return p
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestReadSourceFile(t *testing.T) {
	dir := t.TempDir()
	content := []byte("image bytes")

	tests := []struct {
		name        string
		path        string
		errContains string
	}{
		{"readable file", writeFile(t, dir, "avatar.png", content), ""},
		{"missing file", filepath.Join(dir, "missing.png"), "could not read file"},
		{"directory", dir, "could not read file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, sum, err := readSourceFile(tt.path)
			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, content, data)
			assert.Equal(t, sha256Hex(content), sum)
		})
	}
}

func sourceFileObject(source, sha types.String) types.Object {
	obj, _ := types.ObjectValue(sourceFileAttrTypes, map[string]attr.Value{
		"source": source,
		"sha256": sha,
	})
	return obj
}

func userSchema(t *testing.T) schema.Schema {
	t.Helper()
	resp := &resource.SchemaResponse{}
	NewUserResource().Schema(context.Background(), resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError())
	return resp.Schema
}

// userPlanWithPicture returns a plan for a user whose profile_picture is set
// to picture, with every other attribute typed so it can be written.
func userPlanWithPicture(t *testing.T, s schema.Schema, picture types.Object) tfsdk.Plan {
	t.Helper()
	m := userResourceModel{
		Username:       types.StringValue("test"),
		Email:          types.StringValue("test@example.com"),
		Groups:         types.SetNull(types.StringType),
		CustomClaims:   types.MapNull(types.StringType),
		ProfilePicture: picture,
	}
	plan := tfsdk.Plan{Schema: s}
	diags := plan.Set(context.Background(), &m)
	require.False(t, diags.HasError(), "diags: %v", diags)
	return plan
}

func TestSourceFileSHA256PlanModifier(t *testing.T) {
	s := userSchema(t)
	content := []byte("image bytes")
	existing := writeFile(t, t.TempDir(), "avatar.png", content)

	tests := []struct {
		name        string
		source      types.String
		want        types.String
		errContains string
	}{
		{"hashes the file", types.StringValue(existing), types.StringValue(sha256Hex(content)), ""},
		{"source not known until apply", types.StringUnknown(), types.StringUnknown(), ""},
		{"missing file", types.StringValue(filepath.Join(t.TempDir(), "gone.png")), types.StringUnknown(), "could not read file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				Path:      path.Root("profile_picture").AtName("sha256"),
				Plan:      userPlanWithPicture(t, s, sourceFileObject(tt.source, types.StringUnknown())),
				State:     tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)},
				PlanValue: types.StringUnknown(),
			}
			resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
			sourceFileSHA256PlanModifier{}.PlanModifyString(context.Background(), req, resp)
			if tt.errContains != "" {
				require.True(t, resp.Diagnostics.HasError())
				assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tt.errContains)
				return
			}
			require.False(t, resp.Diagnostics.HasError(), "diags: %v", resp.Diagnostics)
			assert.Equal(t, tt.want, resp.PlanValue)
		})
	}
}

func TestUserResource_ProfilePictureSchema(t *testing.T) {
	attr, ok := userSchema(t).Attributes["profile_picture"].(schema.SingleNestedAttribute)
	require.True(t, ok, "profile_picture should be a single nested attribute")
	assert.True(t, attr.IsOptional())
	assert.False(t, attr.IsComputed())
	assert.True(t, attr.Attributes["source"].IsRequired())
	assert.True(t, attr.Attributes["sha256"].IsComputed())
}
