#!/bin/bash
# Helper to generate a long-lived JWT for Agent authentication
# Usage: ./generate_agent_token.sh [node_id]

NODE_ID=${1:-agent-1}
# Default secret matching control plane dev default
JWT_SECRET=${JWT_SECRET:-"insecure_default_secret_for_dev_mode_only_32bytes"}

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

# 1. Header
header='{"alg":"HS256","typ":"JWT"}'
header_b64=$(echo -n "$header" | base64url)

# 2. Payload (Valid for 10 years)
now=$(date +%s)
exp=$((now + 315360000)) 
payload="{\"tenant_id\":\"default\",\"role\":\"agent\",\"sub\":\"$NODE_ID\",\"iss\":\"fluxforge\",\"aud\":\"fluxforge-api\",\"exp\":$exp,\"iat\":$now,\"nbf\":$now}"
payload_b64=$(echo -n "$payload" | base64url)

# 3. Signature
unsigned="$header_b64.$payload_b64"
signature=$(sign "$unsigned" "$JWT_SECRET")

echo "$unsigned.$signature"
