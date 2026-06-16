#!/bin/bash
set -e

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Configuration
TEST_TOKEN="test-terraform-provider-token-123456789"
TOKEN_HASH=$(echo -n "$TEST_TOKEN" | sha256sum | cut -d' ' -f1)
TEST_DATA_DIR="$PROJECT_ROOT/test-data"
DB_PATH="$TEST_DATA_DIR/data/pocket-id.db"
POCKET_ID_BINARY="$TEST_DATA_DIR/pocket-id"

# Create test data directory
mkdir -p "$TEST_DATA_DIR"

# pocket-id can run either from a downloaded binary (default) or from a
# container image. Set POCKET_ID_IMAGE (e.g. ghcr.io/pocket-id/pocket-id:next)
# to use container mode; otherwise the binary is downloaded and run directly.
POCKET_ID_IMAGE="${POCKET_ID_IMAGE:-}"

# pocket-id v2 requires APP_URL and an ENCRYPTION_KEY of at least 16 bytes.
export APP_URL="${APP_URL:-http://localhost:1411}"
export ENCRYPTION_KEY="${ENCRYPTION_KEY:-test-terraform-provider-encryption-key}"
export TRUST_PROXY="${TRUST_PROXY:-false}"
export MAXMIND_LICENSE_KEY="${MAXMIND_LICENSE_KEY:-}"

# Create data directory for pocket-id (shared with the container volume mount)
mkdir -p "$TEST_DATA_DIR/data"

if [ -n "$POCKET_ID_IMAGE" ]; then
    # Container mode: run pocket-id from a container image. The mounted
    # $TEST_DATA_DIR/data is the same sqlite location the binary would use,
    # so the wait-for-DB-and-seed logic below works identically.
    echo "Starting pocket-id container from image: $POCKET_ID_IMAGE"
    docker rm -f pocket-id-test >/dev/null 2>&1 || true
    docker run -d --name pocket-id-test \
        -e APP_URL \
        -e ENCRYPTION_KEY \
        -e TRUST_PROXY \
        -e MAXMIND_LICENSE_KEY \
        -e PUID="$(id -u)" \
        -e PGID="$(id -g)" \
        -p 1411:1411 \
        -v "$TEST_DATA_DIR/data:/app/data" \
        "$POCKET_ID_IMAGE"
    echo "Pocket-ID container started."
else
    # Binary mode: download pocket-id binary if not present.
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

        # Pin to a known-good pocket-id version for reproducible tests.
        # Override with POCKET_ID_VERSION to test against a different release.
        POCKET_ID_VERSION="${POCKET_ID_VERSION:-v2.9.0}"

        DOWNLOAD_URL="https://github.com/pocket-id/pocket-id/releases/download/${POCKET_ID_VERSION}/pocket-id-${OS}-${ARCH}"
        echo "Downloading from: $DOWNLOAD_URL"

        curl -L -o "$POCKET_ID_BINARY" "$DOWNLOAD_URL"
        chmod +x "$POCKET_ID_BINARY"
    fi

    # Start pocket-id in background
    echo "Starting pocket-id..."
    cd "$TEST_DATA_DIR" && ./pocket-id > pocket-id.log 2>&1 &
    POCKET_ID_PID=$!
    cd "$PROJECT_ROOT"

    echo "Pocket-ID started with PID: $POCKET_ID_PID"

    # Give pocket-id a moment to start
    sleep 2

    # Check if the process is still running
    if ! kill -0 $POCKET_ID_PID 2>/dev/null; then
        echo "ERROR: Pocket-ID process died immediately!"
        echo "Checking log file..."
        if [ -f "$TEST_DATA_DIR/pocket-id.log" ]; then
            cat "$TEST_DATA_DIR/pocket-id.log"
        else
            echo "No log file found!"
        fi
        exit 1
    fi
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
    ls -la "$TEST_DATA_DIR/" || true
    ls -la "$TEST_DATA_DIR/data/" || true
    if [ ! -f "$DB_PATH" ]; then
        echo "Database file does not exist!"
        if [ -n "$POCKET_ID_IMAGE" ]; then
            echo "Checking pocket-id container logs..."
            docker logs pocket-id-test || true
        else
            echo "Checking pocket-id log..."
            cat "$TEST_DATA_DIR/pocket-id.log" || true
        fi
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
    display_name,
    is_admin,
    disabled,
    created_at
) VALUES (
    '$(uuidgen || cat /proc/sys/kernel/random/uuid)',
    'admin@test.local',
    'admin',
    'Test',
    'Admin',
    'Test Admin',
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
if [ -n "$POCKET_ID_IMAGE" ]; then
    echo "Container: pocket-id-test"
else
    echo "PID: $POCKET_ID_PID"
fi
