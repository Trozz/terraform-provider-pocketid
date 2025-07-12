#!/bin/bash
set -e

echo "Checking for missing end-of-file newlines..."
files_without_newline=""
while IFS= read -r -d '' file; do
    if [ -n "$(tail -c 1 "$file")" ]; then
        echo "Missing newline: $file"
        files_without_newline="$files_without_newline $file"
    fi
done < <(find . -type f \( -name "*.go" -o -name "*.tf" -o -name "*.yml" -o -name "*.yaml" -o -name "*.json" -o -name "*.md" \) -print0)

if [ -n "$files_without_newline" ]; then
    echo "❌ Files missing end-of-file newlines found"
    exit 1
else
    echo "✅ All files have proper end-of-file newlines"
fi
