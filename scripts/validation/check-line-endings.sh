#!/bin/bash
set -e

echo "Checking for mixed line endings..."
mixed_files=""
find . -type f \( -name "*.go" -o -name "*.tf" -o -name "*.yml" -o -name "*.yaml" -o -name "*.json" -o -name "*.md" \) -not -path "./.git/*" | while read -r file; do
    if file "$file" | grep -q "CRLF"; then
        echo "CRLF line endings found in: $file"
        mixed_files="$mixed_files $file"
    fi
done

if [ -n "$mixed_files" ]; then
    echo "❌ Found files with CRLF line endings (should be LF)"
    exit 1
else
    echo "✅ All files use consistent LF line endings"
fi
