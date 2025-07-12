#!/bin/bash
set -e

# Function to check job status
check_status() {
    local lint_result="$1"
    local test_result="$2"
    local acceptance_result="$3"
    local terraform_result="$4"
    local security_result="$5"

    # Check if any required jobs failed
    if [ "$lint_result" != "success" ] || \
       [ "$test_result" != "success" ] || \
       [ "$acceptance_result" != "success" ] || \
       [ "$terraform_result" != "success" ]; then
        echo "::error::One or more CI jobs failed"
        echo "status=failed" >> $GITHUB_OUTPUT
        return 1
    else
        echo "All CI jobs passed successfully!"
        echo "status=passed" >> $GITHUB_OUTPUT
        return 0
    fi
}

# Main execution
echo "Lint: $LINT_RESULT"
echo "Test: $TEST_RESULT"
echo "Acceptance Test: $ACCEPTANCE_RESULT"
echo "Terraform Compatibility: $TERRAFORM_RESULT"

check_status "$LINT_RESULT" "$TEST_RESULT" "$ACCEPTANCE_RESULT" "$TERRAFORM_RESULT" "$SECURITY_RESULT"
