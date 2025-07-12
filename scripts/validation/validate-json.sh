#!/bin/bash
set -e

echo "Validating JSON files..."
find . -type f -name "*.json" -not -path "./.git/*" -not -path "./vendor/*" | while read -r file; do
    echo "Checking: $file"
    jq empty "$file" || exit 1
done
echo "✅ All JSON files are valid"
