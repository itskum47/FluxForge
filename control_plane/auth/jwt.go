package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Claims extends the standard JWT claims with FluxForge specific fields.
// STRICT: Must include Role and TenantID.
type Claims struct {
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	// Standard Claims
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"` // We treat single string for native simplicity
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	NotBefore int64  `json:"nbf"`
}

var (
	// STRICT: Enforce 32-byte secret length at startup.
	jwtSecret []byte
	issuer    = "fluxforge"
	audience  = "fluxforge-api"
)

func init() {
	secretEnv := os.Getenv("JWT_SECRET")
	if len(secretEnv) < 32 {
		// STRICT: Panic if secret is weak or missing to prevent insecure startup.
		// Note: For local dev without env, we might panic.
		// User must provide JWT_SECRET.
		if secretEnv == "" {
			fmt.Println("WARNING: JWT_SECRET not set. Using insecure default for blocked network dev ONLY.")
			jwtSecret = []byte("insecure_default_secret_for_dev_mode_only_32bytes")
		} else {
			panic("CRITICAL SECURITY ERROR: JWT_SECRET must be at least 32 characters long.")
		}
	} else {
		jwtSecret = []byte(secretEnv)
	}
}

// GenerateToken creates a signed JWT for the given tenant and role.
func GenerateToken(tenantID, role string) (string, error) {
	now := time.Now().Unix()
	claims := Claims{
		TenantID:  tenantID,
		Role:      role,
		Issuer:    issuer,
		Audience:  audience,
		ExpiresAt: now + 86400, // 24h
		IssuedAt:  now,
		NotBefore: now,
	}

	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	tokenPart := base64UrlEncode(headerJSON) + "." + base64UrlEncode(claimsJSON)
	signature := computeHMAC(tokenPart, jwtSecret)

	return tokenPart + "." + signature, nil
}

// ValidateToken parses and validates the JWT string.
func ValidateToken(tokenString string) (*Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	// 1. Verify Signature
	tokenPart := parts[0] + "." + parts[1]
	signature := computeHMAC(tokenPart, jwtSecret)
	if signature != parts[2] {
		return nil, errors.New("invalid signature")
	}

	// 2. Parse Claims
	claimsJSON, err := base64UrlDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %v", err)
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %v", err)
	}

	// 3. Validate Claims
	now := time.Now().Unix()
	if now > claims.ExpiresAt {
		return nil, errors.New("token expired")
	}
	if claims.Issuer != issuer {
		return nil, errors.New("invalid issuer")
	}
	if claims.Audience != audience {
		return nil, errors.New("invalid audience")
	}

	return &claims, nil
}

func computeHMAC(message string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(message))
	return base64UrlEncode(h.Sum(nil))
}

func base64UrlEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func base64UrlDecode(data string) ([]byte, error) {
	if l := len(data) % 4; l > 0 {
		data += strings.Repeat("=", 4-l)
	}
	return base64.URLEncoding.DecodeString(data)
}
