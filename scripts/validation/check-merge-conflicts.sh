#!/bin/bash
set -e

echo "Checking for merge conflict markers..."
if grep -r '^<<<<<<<\|^=======$\|^>>>>>>>' --include="*.go" --include="*.tf" --include="*.yml" --include="*.yaml" --include="*.json" --include="*.md" . ; then
    echo "❌ Found merge conflict markers"
    exit 1
else
    echo "✅ No merge conflict markers found"
fi
