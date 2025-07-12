#!/bin/bash
set -e

# Function to check job status
check_status() {
    local lint_result="$1"
    local test_result="$2"
    local acceptance_result="$3"
    local terraform_result="$4"
    local security_result="$5"

    # Check if any required jobs failed (skipped is acceptable)
    if [ "$lint_result" = "failure" ] || \
       [ "$test_result" = "failure" ] || \
       [ "$acceptance_result" = "failure" ] || \
       [ "$terraform_result" = "failure" ]; then
        echo "::error::One or more CI jobs failed"
        echo "status=failed" >> $GITHUB_OUTPUT
        return 1
    elif [ "$lint_result" = "cancelled" ] || \
         [ "$test_result" = "cancelled" ] || \
         [ "$acceptance_result" = "cancelled" ] || \
         [ "$terraform_result" = "cancelled" ]; then
        echo "::error::One or more CI jobs were cancelled"
        echo "status=cancelled" >> $GITHUB_OUTPUT
        return 1
    else
        # Count successful and skipped jobs
        local success_count=0
        local skip_count=0

        for result in "$lint_result" "$test_result" "$acceptance_result" "$terraform_result"; do
            if [ "$result" = "success" ]; then
                ((success_count++))
            elif [ "$result" = "skipped" ]; then
                ((skip_count++))
            fi
        done

        if [ $success_count -eq 0 ] && [ $skip_count -gt 0 ]; then
            echo "All CI jobs were skipped (no relevant files changed)"
            echo "status=passed" >> $GITHUB_OUTPUT
            return 0
        else
            echo "CI jobs completed successfully (Success: $success_count, Skipped: $skip_count)"
            echo "status=passed" >> $GITHUB_OUTPUT
            return 0
        fi
    fi
}

# Main execution
echo "Lint: $LINT_RESULT"
echo "Test: $TEST_RESULT"
echo "Acceptance Test: $ACCEPTANCE_RESULT"
echo "Terraform Compatibility: $TERRAFORM_RESULT"

check_status "$LINT_RESULT" "$TEST_RESULT" "$ACCEPTANCE_RESULT" "$TERRAFORM_RESULT" "$SECURITY_RESULT"
