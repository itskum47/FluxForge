package attestation

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"time"
)

// Verifier handles agent binary attestation
type Verifier struct {
	publicKey *rsa.PublicKey
	enabled   bool
}

// NewVerifier creates a new attestation verifier
func NewVerifier(publicKeyPEM string, enabled bool) (*Verifier, error) {
	if !enabled {
		return &Verifier{enabled: false}, nil
	}

	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return &Verifier{
		publicKey: rsaPub,
		enabled:   true,
	}, nil
}

// AttestationClaim represents an agent's attestation claim
type AttestationClaim struct {
	NodeID     string `json:"node_id"`
	BinaryHash string `json:"binary_hash"`
	Version    string `json:"version"`
	Signature  string `json:"signature"`
	Timestamp  int64  `json:"timestamp"`
}

// Verify verifies an agent's attestation claim
// CRITICAL: Includes clock skew tolerance to prevent legitimate agent rejection
func (v *Verifier) Verify(claim *AttestationClaim) error {
	if !v.enabled {
		// Attestation disabled, skip verification
		return nil
	}

	// CRITICAL: Validate timestamp with clock skew tolerance (5 minutes)
	now := time.Now().Unix()
	skew := abs(now - claim.Timestamp)
	const allowedSkew = 5 * 60 // 5 minutes in seconds

	if skew > allowedSkew {
		return fmt.Errorf("timestamp skew too large: %d seconds (max: %d)", skew, allowedSkew)
	}

	// Construct message to verify
	message := fmt.Sprintf("%s:%s:%s:%d",
		claim.NodeID,
		claim.BinaryHash,
		claim.Version,
		claim.Timestamp,
	)

	// Decode signature
	signature, err := base64.StdEncoding.DecodeString(claim.Signature)
	if err != nil {
		return fmt.Errorf("invalid signature encoding: %w", err)
	}

	// Hash message
	hashed := sha256.Sum256([]byte(message))

	// Verify signature
	err = rsa.VerifyPKCS1v15(v.publicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		log.Printf("[ATTESTATION] Verification failed for node %s: %v", claim.NodeID, err)
		return fmt.Errorf("signature verification failed: %w", err)
	}

	log.Printf("[ATTESTATION] ✓ Verified node %s (version %s)", claim.NodeID, claim.Version)
	return nil
}

// abs returns absolute value of int64
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// VerifyBinaryHash verifies the agent binary hash matches expected value
// CRITICAL: Uses constant-time comparison to prevent timing attacks
func (v *Verifier) VerifyBinaryHash(claim *AttestationClaim, expectedHash string) error {
	if !v.enabled {
		return nil
	}

	// CRITICAL: Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(claim.BinaryHash), []byte(expectedHash)) != 1 {
		return fmt.Errorf("binary hash mismatch: got %s, expected %s",
			claim.BinaryHash, expectedHash)
	}

	return nil
}

// IsEnabled returns whether attestation is enabled
func (v *Verifier) IsEnabled() bool {
	return v.enabled
}

// AttestationMetrics tracks attestation statistics
type AttestationMetrics struct {
	TotalVerifications      int64
	SuccessfulVerifications int64
	FailedVerifications     int64
	RejectedAgents          int64
}

// Metrics returns current attestation metrics
func (v *Verifier) Metrics() AttestationMetrics {
	// In production, these would be tracked via Prometheus
	return AttestationMetrics{}
}
