#!/bin/bash

# Test script for HTTPS mock server
# Usage: ./test_https.sh

PORT=8443
BASE_URL="https://localhost:${PORT}"

echo "🧪 Testing HTTPS Mock Server"
echo "=============================="
echo ""

# Check if server is running
echo "1️⃣  Checking if server is accessible..."
if curl -k -s -o /dev/null -w "%{http_code}" "${BASE_URL}/v1/metadata/sn" | grep -q "200"; then
    echo "✅ Server is running"
else
    echo "❌ Server is not accessible. Please start the server with:"
    echo "   ./start_go_mock_server_https.sh"
    exit 1
fi
echo ""

# Test basic GET endpoint
echo "2️⃣  Testing GET /v1/metadata/sn"
response=$(curl -k -s "${BASE_URL}/v1/metadata/sn")
echo "Response: ${response}"
echo ""

# Test JSON endpoint
echo "3️⃣  Testing GET /api/users"
response=$(curl -k -s "${BASE_URL}/api/users")
echo "Response: ${response}"
echo ""

# Test with custom headers
echo "4️⃣  Testing custom headers"
curl -k -I "${BASE_URL}/api/users" 2>&1 | grep -E "(HTTP|X-Custom-Header|Cache-Control)"
echo ""

# Test Web UI
echo "5️⃣  Testing Web UI"
status=$(curl -k -s -o /dev/null -w "%{http_code}" "${BASE_URL}/_mock/ui")
if [ "$status" = "200" ]; then
    echo "✅ Web UI is accessible at ${BASE_URL}/_mock/ui"
else
    echo "❌ Web UI returned status: $status"
fi
echo ""

# Test with certificate verification (will fail with self-signed)
echo "6️⃣  Testing certificate verification (expected to fail with self-signed cert)"
if curl -s --cacert server.crt "${BASE_URL}/v1/metadata/sn" > /dev/null 2>&1; then
    echo "✅ Certificate verification passed"
else
    echo "⚠️  Certificate verification failed (expected with self-signed cert)"
fi
echo ""

echo "✅ HTTPS testing complete!"
echo ""
echo "💡 Tips:"
echo "   - Use 'curl -k' to skip certificate verification"
echo "   - Use '--cacert server.crt' to verify with the self-signed cert"
echo "   - Access Web UI at: ${BASE_URL}/_mock/ui"

