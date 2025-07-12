#!/bin/bash
set -e

# Function to get status emoji
get_status_emoji() {
    local result="$1"
    case "$result" in
        "success") echo "✅ Passed" ;;
        "failure") echo "❌ Failed" ;;
        "skipped") echo "⏭️ Skipped" ;;
        *) echo "⚠️ Unknown" ;;
    esac
}

# Generate the comment based on overall status
if [ "$STATUS" = "failed" ]; then
    cat << EOF
<!-- CI-STATUS -->
## ❌ CI Pipeline Failed

The CI pipeline has encountered failures. Please review the results below:

| Job | Status |
|-----|--------|
| 🔍 **Lint** | $(get_status_emoji "$LINT_RESULT") |
| 🧪 **Test** | $(get_status_emoji "$TEST_RESULT") |
| 🔬 **Acceptance Test** | $(get_status_emoji "$ACCEPTANCE_RESULT") |
| 🔧 **Terraform Compatibility** | $(get_status_emoji "$TERRAFORM_RESULT") |

Please check the [workflow run]($WORKFLOW_URL) for detailed logs.

<details>
<summary>💡 Troubleshooting Tips</summary>

- **Lint failures**: Run \`make lint\` locally to see formatting issues
- **Test failures**: Run \`make test\` to reproduce test failures
- **Acceptance test failures**: Check if Pocket-ID service is properly configured

</details>
EOF
else
    # If we reach here, everything passed - no comment needed
    echo "NO_COMMENT"
fi
