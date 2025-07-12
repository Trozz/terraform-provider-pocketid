#!/bin/bash
set -e

# Configuration
TEST_TOKEN="test-terraform-provider-token-123456789"
TOKEN_HASH=$(echo -n "$TEST_TOKEN" | sha256sum | cut -d' ' -f1)
DB_PATH="./test-data/pocket-id.db"

# Create test data directory
mkdir -p test-data

# Wait for database to exist and migrations to complete
echo "Waiting for Pocket-ID database and migrations..."
MAX_RETRIES=30
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if [ -f "$DB_PATH" ] && sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='api_keys';" 2>/dev/null | grep -q "api_keys"; then
        echo "Database exists and migrations complete!"
        break
    fi
    echo "Waiting for database and migrations... (attempt $((RETRY_COUNT + 1))/$MAX_RETRIES)"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "Error: Timeout waiting for database migrations."
    if [ ! -f "$DB_PATH" ]; then
        echo "Database file does not exist!"
    else
        echo "Tables in database:"
        sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table';" || true
    fi
    exit 1
fi

# Stop Pocket-ID to avoid database read-only mode
echo "Stopping Pocket-ID service..."
docker compose -f docker-compose.test.yml stop pocket-id

# Initialize test data
# First check if an admin user exists
ADMIN_EXISTS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users WHERE is_admin = 1;")

if [ "$ADMIN_EXISTS" -eq 0 ]; then
    echo "No admin user found. Creating test admin user..."
    sqlite3 "$DB_PATH" <<EOF
INSERT INTO users (
    id,
    email,
    username,
    first_name,
    last_name,
    is_admin,
    disabled,
    created_at
) VALUES (
    '$(uuidgen || cat /proc/sys/kernel/random/uuid)',
    'admin@test.local',
    'admin',
    'Test',
    'Admin',
    1,
    0,
    datetime('now')
);
EOF
fi

# Now create the API key
sqlite3 "$DB_PATH" <<EOF
INSERT OR REPLACE INTO api_keys (
    id,
    key,
    user_id,
    name,
    created_at,
    expires_at
) VALUES (
    '$(uuidgen || cat /proc/sys/kernel/random/uuid)',
    '$TOKEN_HASH',
    (SELECT id FROM users WHERE is_admin = 1 LIMIT 1),
    'Terraform Test Token',
    datetime('now'),
    datetime('now', '+1 year')
);
EOF

echo "Test token initialized:"
echo "  POCKETID_API_TOKEN=$TEST_TOKEN"

# Restart Pocket-ID service
echo "Restarting Pocket-ID service..."
docker compose -f docker-compose.test.yml start pocket-id

echo "Pocket-ID test environment ready!"
