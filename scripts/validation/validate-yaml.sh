#!/bin/bash
set -e

echo "Validating YAML files..."
find . -type f \( -name "*.yml" -o -name "*.yaml" \) -not -path "./.git/*" | while read -r file; do
    echo "Checking: $file"
    python3 -c "import yaml; yaml.safe_load(open('$file'))" || exit 1
done
echo "✅ All YAML files are valid"
