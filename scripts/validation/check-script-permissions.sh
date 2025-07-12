#!/bin/bash
set -e

echo "Checking executable permissions on scripts..."
scripts_without_exec=""
for file in $(find . -type f -name "*.sh" -not -path "./.git/*"); do
    if [ ! -x "$file" ]; then
        echo "Not executable: $file"
        scripts_without_exec="$scripts_without_exec $file"
    fi
done

if [ -n "$scripts_without_exec" ]; then
    echo "❌ Found scripts without executable permissions"
    echo "Run: chmod +x <script> to fix"
    exit 1
else
    echo "✅ All scripts have executable permissions"
fi
