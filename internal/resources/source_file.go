package resources

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// sourceFileModel maps an image uploaded from a local file.
type sourceFileModel struct {
	Source types.String `tfsdk:"source"`
	SHA256 types.String `tfsdk:"sha256"`
}

var sourceFileAttrTypes = map[string]attr.Type{
	"source": types.StringType,
	"sha256": types.StringType,
}

// readSourceFile returns the contents of a file and their hex SHA-256.
func readSourceFile(source string) ([]byte, string, error) {
	data, err := os.ReadFile(source) // #nosec G304 - Path comes from the operator's Terraform configuration
	if err != nil {
		return nil, "", fmt.Errorf("could not read file: %w", err)
	}
	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:]), nil
}

// sourceFileSHA256PlanModifier plans the sha256 of the file its sibling
// source points at, so editing the file is a change even though the path in
// the configuration stays the same.
type sourceFileSHA256PlanModifier struct{}

func (m sourceFileSHA256PlanModifier) Description(context.Context) string {
	return "Hashes the source file at plan time."
}

func (m sourceFileSHA256PlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m sourceFileSHA256PlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	sourcePath := req.Path.ParentPath().AtName("source")
	var source types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, sourcePath, &source)...)
	if resp.Diagnostics.HasError() || source.IsNull() {
		return
	}
	if source.IsUnknown() {
		resp.PlanValue = types.StringUnknown()
		return
	}

	_, sum, err := readSourceFile(source.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(sourcePath, "Invalid source file", err.Error())
		return
	}
	resp.PlanValue = types.StringValue(sum)
}
