package attestation

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

// Signer creates attestation signatures for agents
// This would run on the agent side
type Signer struct {
	privateKey *rsa.PrivateKey
	nodeID     string
	version    string
	binaryHash string
}

// NewSigner creates a new attestation signer
func NewSigner(privateKey *rsa.PrivateKey, nodeID, version, binaryHash string) *Signer {
	return &Signer{
		privateKey: privateKey,
		nodeID:     nodeID,
		version:    version,
		binaryHash: binaryHash,
	}
}

// CreateClaim creates a signed attestation claim
func (s *Signer) CreateClaim() (*AttestationClaim, error) {
	timestamp := time.Now().Unix()

	// Construct message
	message := fmt.Sprintf("%s:%s:%s:%d",
		s.nodeID,
		s.binaryHash,
		s.version,
		timestamp,
	)

	// Hash message
	hashed := sha256.Sum256([]byte(message))

	// Sign
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	// Encode signature
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	return &AttestationClaim{
		NodeID:     s.nodeID,
		BinaryHash: s.binaryHash,
		Version:    s.version,
		Signature:  signatureB64,
		Timestamp:  timestamp,
	}, nil
}
