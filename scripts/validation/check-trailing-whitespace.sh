#!/bin/bash
set -e

echo "Checking for trailing whitespace..."
if grep -r '[[:space:]]$' --include="*.go" --include="*.tf" --include="*.yml" --include="*.yaml" --include="*.json" --include="*.md" . ; then
    echo "❌ Found files with trailing whitespace"
    exit 1
else
    echo "✅ No trailing whitespace found"
fi
