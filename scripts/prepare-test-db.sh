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

# Wait for migrations to complete by checking if api_keys table exists
echo "Waiting for database migrations to complete..."
MAX_RETRIES=30
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='api_keys';" 2>/dev/null | grep -q "api_keys"; then
        echo "Database migrations complete!"
        break
    fi
    echo "Waiting for api_keys table to be created... (attempt $((RETRY_COUNT + 1))/$MAX_RETRIES)"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "Error: Timeout waiting for database migrations. Tables in database:"
    sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table';" || true
    exit 1
fi

# Additional safety delay
sleep 2

# Initialize test data
sqlite3 "$DB_PATH" <<EOF
INSERT OR REPLACE INTO api_keys (
    key,
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
