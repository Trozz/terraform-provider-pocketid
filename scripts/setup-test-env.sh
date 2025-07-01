#!/bin/bash

# Terraform Provider for Pocket-ID - Test Environment Setup Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Terraform Provider for Pocket-ID - Test Environment Setup ===${NC}"
echo

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Error: Docker Compose is not installed. Please install Docker Compose first.${NC}"
    exit 1
fi

# Create test directory
TEST_DIR="pocket-id-test"
echo -e "${GREEN}Creating test directory: ${TEST_DIR}${NC}"
mkdir -p ${TEST_DIR}
cd ${TEST_DIR}

# Download official Pocket-ID files
echo -e "${GREEN}Downloading official Pocket-ID configuration...${NC}"
curl -sO https://raw.githubusercontent.com/pocket-id/pocket-id/main/docker-compose.yml
curl -sO https://raw.githubusercontent.com/pocket-id/pocket-id/main/.env.example
cp .env.example .env

# Create nginx.conf template
echo -e "${GREEN}Creating nginx.conf template...${NC}"
cat > nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    upstream pocket-id {
        server pocket-id:1411;
    }

    server {
        listen 443 ssl;
        server_name YOUR_DOMAIN_HERE;  # TODO: Replace with your domain

        ssl_certificate /etc/nginx/certs/fullchain.pem;      # TODO: Update cert path
        ssl_certificate_key /etc/nginx/certs/privkey.pem;   # TODO: Update key path

        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers on;

        location / {
            proxy_pass http://pocket-id;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # WebSocket support
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
        }
    }

    server {
        listen 80;
        server_name YOUR_DOMAIN_HERE;  # TODO: Replace with your domain
        return 301 https://$server_name$request_uri;
    }
}
EOF

# Create docker-compose override for nginx
echo -e "${GREEN}Creating docker-compose.override.yml for nginx...${NC}"
cat > docker-compose.override.yml << 'EOF'
services:
  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - "./nginx.conf:/etc/nginx/nginx.conf:ro"
      - "./certs:/etc/nginx/certs:ro"  # TODO: Update with your cert path
    depends_on:
      - pocket-id
EOF

# Create directories
mkdir -p data certs

# Print instructions
echo
echo -e "${YELLOW}=== Setup Instructions ===${NC}"
echo
echo -e "${BLUE}1. Configure your domain:${NC}"
echo "   Edit nginx.conf and replace YOUR_DOMAIN_HERE with your actual domain"
echo
echo -e "${BLUE}2. Add TLS certificates:${NC}"
echo "   Copy your certificates to the certs/ directory:"
echo "   - fullchain.pem (certificate chain)"
echo "   - privkey.pem (private key)"
echo
echo "   ${YELLOW}Options for getting certificates:${NC}"
echo "   a) Let's Encrypt (for public domains):"
echo "      certbot certonly --standalone -d your-domain.com"
echo
echo "   b) mkcert (for local development):"
echo "      mkcert -install"
echo "      mkcert -cert-file certs/fullchain.pem -key-file certs/privkey.pem your-domain.local"
echo
echo -e "${BLUE}3. Update environment file:${NC}"
echo "   Edit .env and set:"
echo "   PUBLIC_APP_URL=https://your-domain.com"
echo
echo -e "${BLUE}4. Update docker-compose.override.yml:${NC}"
echo "   If your certificates are in a different location, update the path"
echo
echo -e "${BLUE}5. Add hosts entry (if using local domain):${NC}"
echo "   Add to /etc/hosts:"
echo "   127.0.0.1 your-domain.local"
echo
echo -e "${GREEN}=== Next Steps ===${NC}"
echo
echo "1. Complete the configuration steps above"
echo "2. Start Pocket-ID: docker-compose up -d"
echo "3. Access Pocket-ID at https://your-domain.com"
echo "4. Create an admin user and register a passkey"
echo "5. Generate an API key in Settings → API Keys"
echo "6. Set environment variables for testing:"
echo "   export POCKETID_BASE_URL=\"https://your-domain.com\""
echo "   export POCKETID_API_TOKEN=\"your-api-key\""
echo
echo -e "${YELLOW}Note: Passkeys require HTTPS with valid certificates!${NC}"
echo -e "${YELLOW}Self-signed certificates may not work properly with passkeys.${NC}"
echo
echo -e "${GREEN}Setup files created in: $(pwd)${NC}"
