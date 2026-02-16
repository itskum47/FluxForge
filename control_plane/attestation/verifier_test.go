package attestation

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"
)

func TestAttestationVerification(t *testing.T) {
	// Generate test key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Export public key as PEM
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	// Create verifier
	verifier, err := NewVerifier(string(pubKeyPEM), true)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Create signer
	signer := NewSigner(privateKey, "test-agent-1", "v1.0.0", "abc123hash")

	// Create claim
	claim, err := signer.CreateClaim()
	if err != nil {
		t.Fatalf("Failed to create claim: %v", err)
	}

	// Verify claim
	err = verifier.Verify(claim)
	if err != nil {
		t.Errorf("Verification failed: %v", err)
	}

	t.Log("✓ Attestation verification passed")
}

func TestAttestationTampering(t *testing.T) {
	// Generate test key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	verifier, err := NewVerifier(string(pubKeyPEM), true)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	signer := NewSigner(privateKey, "test-agent-1", "v1.0.0", "abc123hash")

	claim, err := signer.CreateClaim()
	if err != nil {
		t.Fatalf("Failed to create claim: %v", err)
	}

	// Tamper with claim
	claim.BinaryHash = "tampered-hash"

	// Verification should fail
	err = verifier.Verify(claim)
	if err == nil {
		t.Error("Expected verification to fail for tampered claim")
	}

	t.Log("✓ Tampering detection passed")
}

func TestAttestationDisabled(t *testing.T) {
	// Create verifier with attestation disabled
	verifier, err := NewVerifier("", false)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Create invalid claim
	claim := &AttestationClaim{
		NodeID:     "test-agent",
		BinaryHash: "invalid",
		Version:    "v1.0.0",
		Signature:  "invalid",
		Timestamp:  time.Now().Unix(),
	}

	// Should not fail when disabled
	err = verifier.Verify(claim)
	if err != nil {
		t.Errorf("Verification should pass when disabled: %v", err)
	}

	t.Log("✓ Disabled attestation passed")
}

func TestBinaryHashVerification(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	verifier, err := NewVerifier(string(pubKeyPEM), true)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	expectedHash := "abc123hash"
	signer := NewSigner(privateKey, "test-agent-1", "v1.0.0", expectedHash)

	claim, err := signer.CreateClaim()
	if err != nil {
		t.Fatalf("Failed to create claim: %v", err)
	}

	// Verify binary hash matches
	err = verifier.VerifyBinaryHash(claim, expectedHash)
	if err != nil {
		t.Errorf("Binary hash verification failed: %v", err)
	}

	// Verify binary hash mismatch detected
	err = verifier.VerifyBinaryHash(claim, "wrong-hash")
	if err == nil {
		t.Error("Expected binary hash mismatch to be detected")
	}

	t.Log("✓ Binary hash verification passed")
}
