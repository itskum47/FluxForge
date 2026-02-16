# Agent Binary Attestation

## Overview

Agent binary attestation ensures that only authorized, unmodified agent binaries can connect to the FluxForge control plane. This prevents rogue or compromised agents from joining the cluster.

## How It Works

### 1. Key Generation

Generate an RSA key pair for signing agent binaries:

```bash
# Generate private key (keep secure!)
openssl genrsa -out agent-private.pem 2048

# Extract public key (distribute to control plane)
openssl rsa -in agent-private.pem -pubout -out agent-public.pem
```

### 2. Agent Registration

When an agent starts, it creates an attestation claim:

```go
// Agent side
signer := attestation.NewSigner(
    privateKey,
    "agent-123",
    "v1.2.3",
    "sha256:abc123...", // Binary hash
)

claim, err := signer.CreateClaim()
// Send claim to control plane during registration
```

### 3. Control Plane Verification

Control plane verifies the attestation:

```go
// Control plane side
verifier, err := attestation.NewVerifier(publicKeyPEM, true)

err = verifier.Verify(claim)
if err != nil {
    // Reject agent registration
    return fmt.Errorf("attestation failed: %w", err)
}

// Also verify binary hash matches expected version
err = verifier.VerifyBinaryHash(claim, expectedHash)
```

## Attestation Claim Structure

```json
{
  "node_id": "agent-123",
  "binary_hash": "sha256:abc123def456...",
  "version": "v1.2.3",
  "signature": "base64-encoded-signature",
  "timestamp": 1708012345
}
```

## Security Properties

### What Attestation Prevents

✅ **Rogue Agents**: Unsigned binaries cannot connect  
✅ **Modified Binaries**: Hash verification detects tampering  
✅ **Version Enforcement**: Only approved versions allowed  
✅ **Replay Attacks**: Timestamp validation prevents reuse

### What Attestation Does NOT Prevent

❌ **Compromised Keys**: If private key is stolen, attacker can sign  
❌ **Runtime Tampering**: Only verifies binary at startup  
❌ **Network Attacks**: Does not protect communication (use mTLS)

## Configuration

### Control Plane

```env
# Enable attestation
ATTESTATION_ENABLED=true

# Public key for verification
ATTESTATION_PUBLIC_KEY_PATH=/etc/fluxforge/agent-public.pem

# Expected binary hash (optional, for version pinning)
ATTESTATION_EXPECTED_HASH=sha256:abc123def456...
```

### Agent

```env
# Private key for signing
ATTESTATION_PRIVATE_KEY_PATH=/etc/fluxforge/agent-private.pem

# Agent metadata
AGENT_VERSION=v1.2.3
AGENT_BINARY_PATH=/usr/local/bin/fluxforge-agent
```

## Integration with Registration API

```go
func (a *API) handleRegisterAgent(w http.ResponseWriter, r *http.Request) {
    var req struct {
        NodeID      string                      `json:"node_id"`
        Attestation *attestation.AttestationClaim `json:"attestation"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Verify attestation
    if a.attestationVerifier.IsEnabled() {
        if err := a.attestationVerifier.Verify(req.Attestation); err != nil {
            http.Error(w, "Attestation failed", http.StatusUnauthorized)
            return
        }
    }
    
    // Proceed with registration
    // ...
}
```

## Metrics

```prometheus
# Total attestation verifications
flux_attestation_verifications_total{result="success|failure"}

# Rejected agents due to attestation failure
flux_attestation_rejections_total

# Attestation verification latency
flux_attestation_verification_duration_seconds
```

## Operational Procedures

### Key Rotation

1. Generate new key pair
2. Sign new agent binaries with new private key
3. Deploy new agents gradually
4. Update control plane with new public key
5. Retire old key after all agents upgraded

### Incident Response

If private key is compromised:

1. **Immediate**: Disable attestation temporarily
2. **Generate**: New key pair
3. **Rebuild**: All agent binaries with new key
4. **Deploy**: New agents to all nodes
5. **Re-enable**: Attestation with new public key

## Testing

```bash
# Run attestation tests
cd control_plane
go test ./attestation -v

# Test with real keys
go test ./attestation -run TestAttestationVerification -v

# Test tampering detection
go test ./attestation -run TestAttestationTampering -v
```

## Production Deployment

### Phase 1: Soft Launch (Week 1)

- Enable attestation in audit-only mode
- Log verification results without rejecting
- Monitor false positive rate

### Phase 2: Enforcement (Week 2)

- Enable rejection of failed attestations
- Monitor rejection rate
- Have rollback plan ready

### Phase 3: Hardening (Week 3+)

- Enable binary hash pinning
- Enforce version requirements
- Implement key rotation schedule

## Best Practices

1. **Secure Key Storage**: Store private keys in HSM or secure vault
2. **Automated Signing**: Integrate signing into CI/CD pipeline
3. **Version Pinning**: Pin expected binary hash in production
4. **Monitoring**: Alert on high attestation failure rates
5. **Key Rotation**: Rotate keys every 90 days

## Troubleshooting

### Agent Registration Fails

```
Error: attestation failed: signature verification failed
```

**Causes**:
- Binary was modified after signing
- Wrong private key used for signing
- Clock skew between agent and control plane

**Solutions**:
- Rebuild agent binary
- Verify correct key is being used
- Sync clocks via NTP

### High Rejection Rate

**Investigate**:
- Check if agents are using old keys
- Verify binary hash matches expected
- Check for clock skew issues

## Summary

Agent binary attestation provides:

- ✅ **Authentication**: Verify agent identity
- ✅ **Integrity**: Detect binary tampering
- ✅ **Authorization**: Enforce version requirements
- ✅ **Auditability**: Log all verification attempts

**Status**: ✅ Complete and tested
