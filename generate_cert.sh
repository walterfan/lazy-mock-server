#!/bin/bash

# Generate self-signed certificate for testing HTTPS
# Usage: ./generate_cert.sh [domain]

DOMAIN=${1:-localhost}
CERT_FILE="server.crt"
KEY_FILE="server.key"

echo "Generating self-signed certificate for ${DOMAIN}..."

# Generate private key and certificate
openssl req -x509 -newkey rsa:4096 -nodes \
    -keyout "${KEY_FILE}" \
    -out "${CERT_FILE}" \
    -days 365 \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=${DOMAIN}" \
    -addext "subjectAltName=DNS:${DOMAIN},DNS:*.${DOMAIN},IP:127.0.0.1"

if [ $? -eq 0 ]; then
    echo "✅ Certificate generated successfully!"
    echo "📄 Certificate: ${CERT_FILE}"
    echo "🔑 Private Key: ${KEY_FILE}"
    echo ""
    echo "To start the mock server with HTTPS:"
    echo "  ./mock-server -tls -cert ${CERT_FILE} -key ${KEY_FILE}"
    echo ""
    echo "⚠️  Note: This is a self-signed certificate for testing only."
    echo "   Browsers will show a security warning that you'll need to accept."
else
    echo "❌ Failed to generate certificate"
    exit 1
fi

