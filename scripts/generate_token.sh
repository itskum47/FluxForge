#!/bin/bash
# Generate a JWT for testing
# Usage: ./generate_token.sh [tenant_id] [role] [subject]

TENANT_ID=${1:-default}
ROLE=${2:-admin}
SUBJECT=${3:-user-1}

# Default secret matching control plane dev default
JWT_SECRET=${JWT_SECRET:-"insecure_default_secret_for_dev_mode_only_32bytes"}

base64url() {
    openssl enc -base64 -A | tr '+/' '-_' | tr -d '='
}

sign() {
    local data="$1"
    local secret="$2"
    echo -n "$data" | openssl dgst -sha256 -hmac "$secret" -binary | base64url
}

header='{"alg":"HS256","typ":"JWT"}'
header_b64=$(echo -n "$header" | base64url)

now=$(date +%s)
exp=$((now + 3600))
payload="{\"tenant_id\":\"$TENANT_ID\",\"role\":\"$ROLE\",\"sub\":\"$SUBJECT\",\"iss\":\"fluxforge\",\"aud\":\"fluxforge-api\",\"exp\":$exp,\"iat\":$now,\"nbf\":$now}"
payload_b64=$(echo -n "$payload" | base64url)

unsigned="$header_b64.$payload_b64"
signature=$(sign "$unsigned" "$JWT_SECRET")

echo "$unsigned.$signature"
