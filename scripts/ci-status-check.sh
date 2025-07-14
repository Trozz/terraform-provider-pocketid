#!/bin/bash
set -e

# Set default GITHUB_OUTPUT if not in GitHub Actions environment
if [ -z "$GITHUB_OUTPUT" ]; then
    GITHUB_OUTPUT="/dev/stdout"
fi

# Function to check job status
check_status() {
    local lint_result="$1"
    local test_result="$2"
    local acceptance_result="$3"
    local build_result="$4"
    local terraform_result="$5"

    # Check if any required jobs failed (skipped is acceptable)
    if [ "$lint_result" = "failure" ] || \
       [ "$test_result" = "failure" ] || \
       [ "$acceptance_result" = "failure" ] || \
       [ "$build_result" = "failure" ] || \
       [ "$terraform_result" = "failure" ]; then
        echo "::error::One or more CI jobs failed"
        echo "status=failed" >> "$GITHUB_OUTPUT"
        return 1
    elif [ "$lint_result" = "cancelled" ] || \
         [ "$test_result" = "cancelled" ] || \
         [ "$acceptance_result" = "cancelled" ] || \
         [ "$build_result" = "cancelled" ] || \
         [ "$terraform_result" = "cancelled" ]; then
        echo "::error::One or more CI jobs were cancelled"
        echo "status=cancelled" >> "$GITHUB_OUTPUT"
        return 1
    else
        # Count successful and skipped jobs
        local success_count=0
        local skip_count=0

        for result in "$lint_result" "$test_result" "$acceptance_result" "$build_result" "$terraform_result"; do
            if [ "$result" = "success" ]; then
                success_count=$((success_count + 1))
            elif [ "$result" = "skipped" ]; then
                skip_count=$((skip_count + 1))
            fi
        done

        if [ $success_count -eq 0 ] && [ $skip_count -gt 0 ]; then
            echo "All CI jobs were skipped (no relevant files changed)"
            echo "status=passed" >> "$GITHUB_OUTPUT"
            return 0
        else
            echo "CI jobs completed successfully (Success: $success_count, Skipped: $skip_count)"
            echo "status=passed" >> "$GITHUB_OUTPUT"
            return 0
        fi
    fi
}

# Main execution
echo "Lint: $LINT_RESULT"
echo "Test: $TEST_RESULT"
echo "Acceptance Test: $ACCEPTANCE_RESULT"
echo "Build Provider: $BUILD_RESULT"
echo "Terraform Compatibility: $TERRAFORM_RESULT"

# Call check_status and capture its exit code
if check_status "$LINT_RESULT" "$TEST_RESULT" "$ACCEPTANCE_RESULT" "$BUILD_RESULT" "$TERRAFORM_RESULT"; then
    exit 0
else
    exit 1
fi
