#!/bin/bash
set -e

echo "Checking for case-conflicting filenames..."
# Create a temporary file to store lowercase filenames
tmp_file=$(mktemp)
find . -type f -not -path "./.git/*" | while read -r file; do
    basename_lower=$(basename "$file" | tr '[:upper:]' '[:lower:]')
    dirname_part=$(dirname "$file")
    echo "$dirname_part/$basename_lower" >> "$tmp_file"
done

# Check for duplicates
if [ "$(sort "$tmp_file" | uniq -d | wc -l)" -gt 0 ]; then
    echo "❌ Found potential case conflicts:"
    sort "$tmp_file" | uniq -d
    rm "$tmp_file"
    exit 1
else
    echo "✅ No case-conflicting filenames found"
    rm "$tmp_file"
fi
