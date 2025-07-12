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
    is_admin,
    created_at,
    updated_at
) VALUES (
    '$(uuidgen || cat /proc/sys/kernel/random/uuid)',
    'admin@test.local',
    'admin',
    1,
    datetime('now'),
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

# Wait for Pocket-ID to be healthy
echo "Waiting for Pocket-ID to be healthy..."
MAX_HEALTH_RETRIES=30
HEALTH_RETRY_COUNT=0
while [ $HEALTH_RETRY_COUNT -lt $MAX_HEALTH_RETRIES ]; do
    if docker compose -f docker-compose.test.yml ps pocket-id | grep -q "healthy"; then
        echo "Pocket-ID is healthy!"
        break
    fi
    echo "Waiting for Pocket-ID to become healthy... (attempt $((HEALTH_RETRY_COUNT + 1))/$MAX_HEALTH_RETRIES)"
    sleep 2
    HEALTH_RETRY_COUNT=$((HEALTH_RETRY_COUNT + 1))
done

if [ $HEALTH_RETRY_COUNT -eq $MAX_HEALTH_RETRIES ]; then
    echo "Error: Timeout waiting for Pocket-ID to become healthy"
    docker compose -f docker-compose.test.yml logs pocket-id
    exit 1
fi

echo "Pocket-ID test environment ready!"
