#!/bin/bash
set -e

# Configuration
TEST_TOKEN="test-terraform-provider-token-123456789"
TOKEN_HASH=$(echo -n "$TEST_TOKEN" | sha256sum | cut -d' ' -f1)
DB_PATH="./test-data/data/pocket-id.db"
POCKET_ID_BINARY="./test-data/pocket-id"

# Create test data directory
mkdir -p test-data

# Download pocket-id binary if not present
if [ ! -f "$POCKET_ID_BINARY" ]; then
    echo "Downloading pocket-id binary..."
    # Detect OS and architecture
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    # Map architecture names
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
    esac

    # Map OS names for pocket-id releases
    case "$OS" in
        darwin) OS="macos" ;;
    esac

    # Get latest release version
    LATEST_VERSION=$(curl -s https://api.github.com/repos/pocket-id/pocket-id/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$LATEST_VERSION" ]; then
        echo "Failed to get latest version, using v1.6.2"
        LATEST_VERSION="v1.6.2"
    fi

    DOWNLOAD_URL="https://github.com/pocket-id/pocket-id/releases/download/${LATEST_VERSION}/pocket-id-${OS}-${ARCH}"
    echo "Downloading from: $DOWNLOAD_URL"

    curl -L -o "$POCKET_ID_BINARY" "$DOWNLOAD_URL"
    chmod +x "$POCKET_ID_BINARY"
fi

# Create data directory for pocket-id
mkdir -p test-data/data

# Start pocket-id in background
echo "Starting pocket-id..."
cd test-data && ./pocket-id > pocket-id.log 2>&1 &
POCKET_ID_PID=$!
cd ..

echo "Pocket-ID started with PID: $POCKET_ID_PID"

# Give pocket-id a moment to start
sleep 2

# Check if the process is still running
if ! kill -0 $POCKET_ID_PID 2>/dev/null; then
    echo "ERROR: Pocket-ID process died immediately!"
    echo "Checking log file..."
    if [ -f "./test-data/pocket-id.log" ]; then
        cat ./test-data/pocket-id.log
    else
        echo "No log file found!"
    fi
    exit 1
fi

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
    echo "Expected database path: $DB_PATH"
    echo "Checking directory contents..."
    ls -la ./test-data/ || true
    ls -la ./test-data/data/ || true
    if [ ! -f "$DB_PATH" ]; then
        echo "Database file does not exist!"
        echo "Checking pocket-id log..."
        cat ./test-data/pocket-id.log || true
    else
        echo "Tables in database:"
        sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table';" || true
    fi
    exit 1
fi

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
echo "Creating API key..."
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

# Wait for pocket-id to be ready
echo "Waiting for pocket-id to be ready..."
for i in {1..10}; do
    if curl -s http://localhost:1411/ > /dev/null 2>&1; then
        echo "Pocket-ID is ready!"
        break
    fi
    echo "Waiting for pocket-id to start... (attempt $i/10)"
    sleep 1
done

echo "Pocket-ID test environment ready!"
echo "PID: $POCKET_ID_PID"
