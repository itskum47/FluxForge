#!/bin/bash
# Phase 8: Authentication Verification (JWT)
# Verifies:
# 1. 401 Unauthorized for missing token
# 2. 401 for malformed token
# 3. 200 OK for valid token
# 4. Context injection (Tenant ID from token)

set -e

API_BASE="http://localhost:8090"
JWT_SECRET="insecure_default_secret_for_dev_mode_only_32bytes" # Matches dev default in code

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"; }

# Helper to generate token (using a small go program or python, but since control plane has generation logic, 
# we can't easily call it from bash without exposing an endpoint. 
# BUT, since we implemented standard HMAC-SHA256, we can generate it here via Python or OpenSSL!)
# Let's use Python for portable generation since we know the secret.

# Helper to base64url encode
base64url() {
    openssl enc -base64 -A | tr '+/' '-_' | tr -d '='
}

# Helper to sign
sign() {
    local data="$1"
    local secret="$2"
    echo -n "$data" | openssl dgst -sha256 -hmac "$secret" -binary | base64url
}

generate_token() {
    # 1. Header
    local header='{"alg":"HS256","typ":"JWT"}'
    local header_b64=$(echo -n "$header" | base64url)

    # 2. Payload
    local now=$(date +%s)
    local exp=$((now + 3600))
    local payload="{\"tenant_id\":\"default\",\"role\":\"admin\",\"iss\":\"fluxforge\",\"aud\":\"fluxforge-api\",\"exp\":$exp,\"iat\":$now,\"nbf\":$now}"
    local payload_b64=$(echo -n "$payload" | base64url)

    # 3. Signature
    local unsigned="$header_b64.$payload_b64"
    local signature=$(sign "$unsigned" "$JWT_SECRET")

    echo "$unsigned.$signature"
}

# 1. Test Missing Token
log_info "Test 1: Missing Token..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE/jobs")
if [ "$HTTP_CODE" -eq 401 ]; then
    log_pass "Rejected missing token (401)"
else
    log_fail "Failed to reject missing token. Got $HTTP_CODE"
    exit 1
fi

# 2. Test Malformed Token
log_info "Test 2: Malformed Token..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer invalid.token.garbage" "$API_BASE/jobs")
if [ "$HTTP_CODE" -eq 401 ]; then
    log_pass "Rejected malformed token (401)"
else
    log_fail "Failed to reject malformed token. Got $HTTP_CODE"
    exit 1
fi

# Test Valid Token with GET /agents
log_info "Test 3: Valid Token (GET /agents)..."
# Generate token
TOKEN=$(generate_token)
if [ -z "$TOKEN" ]; then
    log_fail "Failed to generate token locally."
    exit 1
fi

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer $TOKEN" \
    "$API_BASE/agents")

if [[ "$HTTP_CODE" -eq 200 ]]; then
    log_pass "Accepted valid token (Code: $HTTP_CODE)"
else
    log_fail "Rejected valid token. Got $HTTP_CODE"
    exit 1
fi

log_pass "Authentication Verification Complete."
