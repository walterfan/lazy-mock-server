#!/bin/bash

# Start the Go mock server with HTTPS support
# This script checks for certificates and generates them if needed

CERT_FILE="server.crt"
KEY_FILE="server.key"
PORT=8443

# Check if certificates exist
if [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
    echo "⚠️  TLS certificates not found. Generating self-signed certificates..."
    ./generate_cert.sh
    echo ""
fi

# Start the server with HTTPS
echo "🚀 Starting Lazy Mock Server with HTTPS..."
echo "🔒 Certificate: $CERT_FILE"
echo "🔑 Private Key: $KEY_FILE"
echo "🌐 Server URL: https://localhost:$PORT"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

./mock-server -tls -cert "$CERT_FILE" -key "$KEY_FILE" -port "$PORT"

