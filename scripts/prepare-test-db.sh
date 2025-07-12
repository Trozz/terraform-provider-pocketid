#!/bin/bash
set -e

# Configuration
TEST_TOKEN="test-terraform-provider-token-123456789"
TOKEN_HASH=$(echo -n "$TEST_TOKEN" | sha256sum | cut -d' ' -f1)
DB_PATH="./test-data/pocket-id.db"

# Create test data directory
mkdir -p test-data

# Wait for database to exist
echo "Waiting for Pocket-ID database..."
while [ ! -f "$DB_PATH" ]; do
    sleep 1
done
sleep 2

# Initialize test data
sqlite3 "$DB_PATH" <<EOF
INSERT OR REPLACE INTO api_tokens (
    token_hash,
    user_id,
    name,
    created_at
) VALUES (
    '$TOKEN_HASH',
    (SELECT id FROM users WHERE is_admin = 1 LIMIT 1),
    'Terraform Test Token',
    datetime('now')
);
EOF

echo "Test token initialized:"
echo "  POCKETID_API_TOKEN=$TEST_TOKEN"
