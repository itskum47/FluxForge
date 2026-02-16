package coordination

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/itskum47/FluxForge/control_plane/store"
)

type LockJanitor struct {
	coordinator store.Coordinator
	store       store.Store // For GetDurableEpoch
	interval    time.Duration
}

func NewLockJanitor(c store.Coordinator, s store.Store, interval time.Duration) *LockJanitor {
	return &LockJanitor{
		coordinator: c,
		store:       s,
		interval:    interval,
	}
}

func (j *LockJanitor) Start(ctx context.Context) {
	go j.loop(ctx)
}

func (j *LockJanitor) loop(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.clean(ctx)
		}
	}
}

func (j *LockJanitor) clean(ctx context.Context) {
	// 1. Get Current Durable Epoch
	// We assume "leader_election" resource for now, or we iterate distinct resources?
	// For now, FluxForge uses a single global leader election resource: "leader_election".
	// If we support multiple, we might need to know which resource a lock belongs to.
	// But locks are named "fluxforge:lock:leader".
	// The resource ID in Postgres is "leader_election".
	// We'll hardcode "leader_election" for the global epoch check.
	// If we have localized locks, we'd need mapping.
	currentEpoch, err := j.store.GetDurableEpoch(ctx, "leader_election")
	if err != nil {
		log.Printf("Janitor: Failed to get durable epoch: %v", err)
		return
	}

	keys, err := j.coordinator.ScanLocks(ctx, "fluxforge:lock:*")
	if err != nil {
		log.Printf("Janitor: Scan failed: %v", err)
		return
	}

	for _, key := range keys {
		// keys scan might return epoch keys if pattern matches ":*"
		if len(key) > 6 && key[len(key)-6:] == ":epoch" {
			continue
		}

		val, err := j.coordinator.GetLockOwner(ctx, key)
		if err != nil {
			continue
		}

		if val == "" {
			continue
		}

		var meta LockMetadata
		if err := json.Unmarshal([]byte(val), &meta); err != nil {
			log.Printf("Janitor: Failed to unmarshal lock %s: %v", key, err)
			continue
		}

		// Check 1: Fencing (Epoch Mismatch)
		if meta.Epoch < currentEpoch {
			log.Printf("Janitor: FENCING lock %s (Epoch %d < Current %d). Force releasing.", key, meta.Epoch, currentEpoch)
			if err := j.coordinator.ReleaseLease(ctx, key, val); err != nil {
				log.Printf("Janitor: Failed to release fenced lock: %v", err)
			}
			continue
		}

		// Check 2: Stale (Physical Time)
		if time.Now().After(meta.ExpiresAt.Add(5 * time.Second)) {
			log.Printf("Janitor: Found stale lock %s (expired at %s). Force releasing.", key, meta.ExpiresAt)
			if err := j.coordinator.ReleaseLease(ctx, key, val); err != nil {
				log.Printf("Janitor: Failed to release stale lock: %v", err)
			} else {
				log.Printf("Janitor: Reclaimed lock %s", key)
			}
		}
	}
}
