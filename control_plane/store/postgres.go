package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore implements Store using a PostgreSQL backend.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore initializes a new PostgresStore with a connection pool.
func NewPostgresStore(ctx context.Context, connString string) (*PostgresStore, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	// Optimize pool settings for concurrent Phase 5 load
	config.MaxConns = 50
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &PostgresStore{pool: pool}, nil
}

// Close closes the connection pool.
func (s *PostgresStore) Close() {
	s.pool.Close()
}

// --- Agent Operations ---

// --- Agent Operations ---

func (s *PostgresStore) UpsertAgent(ctx context.Context, tenantID string, agent *Agent) error {
	agent.TenantID = tenantID
	query := `
		INSERT INTO agents (node_id, tenant_id, hostname, ip_address, port, version, status, health_score, last_heartbeat_at, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		ON CONFLICT (node_id) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			ip_address = EXCLUDED.ip_address,
			port = EXCLUDED.port,
			version = EXCLUDED.version,
			status = EXCLUDED.status,
			last_heartbeat_at = EXCLUDED.last_heartbeat_at,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
	`
	_, err := s.pool.Exec(ctx, query,
		agent.NodeID, agent.TenantID, agent.Hostname, agent.IPAddress, agent.Port,
		agent.Version, agent.Status, agent.LastHeartbeat, agent.Metadata,
	)
	return err
}

func (s *PostgresStore) GetAgent(ctx context.Context, tenantID string, nodeID string) (*Agent, error) {
	query := `
		SELECT node_id, tenant_id, hostname, ip_address, port, version, status, last_heartbeat_at, created_at, updated_at, metadata
		FROM agents WHERE node_id = $1 AND tenant_id = $2
	`
	var a Agent
	// Use Scan to map fields. For metadata map[string]string, pgx handles JSONB automatically if compatible.
	// Since struct has map[string]string and DB has JSONB, we might need a wrapper if simple scan fails,
	// but pgx v5 often handles it. If not, we'll fix it in verification.
	err := s.pool.QueryRow(ctx, query, nodeID, tenantID).Scan(
		&a.NodeID, &a.TenantID, &a.Hostname, &a.IPAddress, &a.Port, &a.Version, &a.Status,
		&a.LastHeartbeat, &a.CreatedAt, &a.UpdatedAt, &a.Metadata,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // Return nil if not found, consistent with store.go interface expectation
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *PostgresStore) ListAgents(ctx context.Context, tenantID string) ([]*Agent, error) {
	query := `
		SELECT node_id, tenant_id, hostname, ip_address, port, version, status, last_heartbeat_at, created_at, updated_at, metadata
		FROM agents WHERE tenant_id = $1
	`
	rows, err := s.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		var a Agent
		if err := rows.Scan(
			&a.NodeID, &a.TenantID, &a.Hostname, &a.IPAddress, &a.Port, &a.Version, &a.Status,
			&a.LastHeartbeat, &a.CreatedAt, &a.UpdatedAt, &a.Metadata,
		); err != nil {
			return nil, err
		}
		agents = append(agents, &a)
	}
	return agents, nil
}

func (s *PostgresStore) UpdateAgentHeartbeat(ctx context.Context, tenantID string, nodeID string, t time.Time) error {
	query := `UPDATE agents SET last_heartbeat_at = $1 WHERE node_id = $2 AND tenant_id = $3`
	tag, err := s.pool.Exec(ctx, query, t, nodeID, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("agent not found")
	}
	return nil
}

// --- State Operations ---

func (s *PostgresStore) UpsertState(ctx context.Context, tenantID string, state *DesiredState) error {
	state.TenantID = tenantID
	query := `
		INSERT INTO desired_states (state_id, node_id, tenant_id, check_cmd, apply_cmd, desired_exit_code, version, status, last_checked, last_error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (state_id) DO UPDATE SET
			check_cmd = EXCLUDED.check_cmd,
			apply_cmd = EXCLUDED.apply_cmd,
			desired_exit_code = EXCLUDED.desired_exit_code,
			version = EXCLUDED.version,
			status = EXCLUDED.status,
			last_checked = EXCLUDED.last_checked,
			last_error = EXCLUDED.last_error
	`
	_, err := s.pool.Exec(ctx, query,
		state.StateID, state.NodeID, state.TenantID, state.CheckCmd, state.ApplyCmd,
		state.DesiredExitCode, state.Version, state.Status, state.LastChecked, state.LastError,
	)
	return err
}

func (s *PostgresStore) UpdateStateStatus(ctx context.Context, tenantID string, stateID string, status string, lastError string, lastChecked time.Time, expectedVersion int) error {
	query := `
		UPDATE desired_states
		SET status = $2, last_error = $3, last_checked = $4
		WHERE state_id = $1 AND version = $5 AND tenant_id = $6
	`
	tag, err := s.pool.Exec(ctx, query, stateID, status, lastError, lastChecked, expectedVersion, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("optimistic lock failure: state version changed")
	}
	return nil
}

func (s *PostgresStore) GetState(ctx context.Context, tenantID string, stateID string) (*DesiredState, error) {
	query := `
		SELECT state_id, node_id, tenant_id, check_cmd, apply_cmd, desired_exit_code, version, status, last_checked, last_error, created_at, updated_at
		FROM desired_states WHERE state_id = $1
	`
	var st DesiredState
	err := s.pool.QueryRow(ctx, query, stateID).Scan(
		&st.StateID, &st.NodeID, &st.TenantID, &st.CheckCmd, &st.ApplyCmd,
		&st.DesiredExitCode, &st.Version, &st.Status, &st.LastChecked, &st.LastError, &st.CreatedAt, &st.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *PostgresStore) GetStateByNode(ctx context.Context, tenantID string, nodeID string) (*DesiredState, error) {
	// Assuming one state per node for now, or getting the latest?
	// The schema handles multiple states if IDs differ, but typically Reconciler looks for "The" state.
	// We'll limit 1 ordered by version desc if multiple exist, or just filter by node_id.
	// Assuming state_id is unique, but node_id might have multiple history?
	// For Phase 5, let's assume one active desired state per node for simplicity, or we grab the latest created.
	query := `
		SELECT state_id, node_id, tenant_id, check_cmd, apply_cmd, desired_exit_code, version, created_at, updated_at
		FROM desired_states WHERE node_id = $1
		ORDER BY created_at DESC LIMIT 1
	`
	var st DesiredState
	err := s.pool.QueryRow(ctx, query, nodeID).Scan(
		&st.StateID, &st.NodeID, &st.TenantID, &st.CheckCmd, &st.ApplyCmd,
		&st.DesiredExitCode, &st.Version, &st.CreatedAt, &st.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *PostgresStore) ListStates(ctx context.Context, tenantID string) ([]*DesiredState, error) {
	query := `
		SELECT state_id, node_id, tenant_id, check_cmd, apply_cmd, desired_exit_code, version, created_at, updated_at
		FROM desired_states WHERE tenant_id = $1
	`
	rows, err := s.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []*DesiredState
	for rows.Next() {
		var st DesiredState
		if err := rows.Scan(
			&st.StateID, &st.NodeID, &st.TenantID, &st.CheckCmd, &st.ApplyCmd,
			&st.DesiredExitCode, &st.Version, &st.CreatedAt, &st.UpdatedAt,
		); err != nil {
			return nil, err
		}
		states = append(states, &st)
	}
	return states, nil
}

// --- Job Operations ---

func (s *PostgresStore) CreateJob(ctx context.Context, tenantID string, job *Job) error {
	job.TenantID = tenantID
	query := `
		INSERT INTO jobs (job_id, node_id, tenant_id, state_id, command, status, exit_code, stdout, stderr, trace_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
	`
	_, err := s.pool.Exec(ctx, query,
		job.JobID, job.NodeID, job.TenantID, job.StateID, job.Command, job.Status,
		job.ExitCode, job.Stdout, job.Stderr, job.TraceID,
	)
	return err
}

func (s *PostgresStore) UpdateJobStatus(ctx context.Context, tenantID string, jobID string, status string, exitCode int, stdout, stderr string) error {
	// Determine timestamps based on status
	var query string
	if status == "running" {
		query = `UPDATE jobs SET status = $2, started_at = NOW() WHERE job_id = $1 AND tenant_id = $3`
		_, err := s.pool.Exec(ctx, query, jobID, status, tenantID)
		return err
	} else if status == "completed" || status == "failed" {
		query = `UPDATE jobs SET status = $2, exit_code = $3, stdout = $4, stderr = $5, finished_at = NOW() WHERE job_id = $1 AND tenant_id = $6`
		_, err := s.pool.Exec(ctx, query, jobID, status, exitCode, stdout, stderr, tenantID)
		return err
	}
	// Default update
	query = `UPDATE jobs SET status = $2 WHERE job_id = $1 AND tenant_id = $3`
	_, err := s.pool.Exec(ctx, query, jobID, status, tenantID)
	return err
}

func (s *PostgresStore) GetJob(ctx context.Context, tenantID string, jobID string) (*Job, error) {
	query := `
		SELECT job_id, node_id, tenant_id, state_id, command, status, exit_code, stdout, stderr, trace_id, created_at, started_at, finished_at
		FROM jobs WHERE job_id = $1 AND tenant_id = $2
	`
	var j Job
	err := s.pool.QueryRow(ctx, query, jobID, tenantID).Scan(
		&j.JobID, &j.NodeID, &j.TenantID, &j.StateID, &j.Command, &j.Status,
		&j.ExitCode, &j.Stdout, &j.Stderr, &j.TraceID, &j.CreatedAt, &j.StartedAt, &j.FinishedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (s *PostgresStore) ListStatesByStatus(ctx context.Context, status string, shardIndex int, shardCount int) ([]*DesiredState, error) {
	var query string
	var args []interface{}

	if shardCount > 1 {
		// PostgreSQL hash sharding: hashtext(node_id) % shardCount == shardIndex
		// Handle negative hash results with abs
		query = `
			SELECT state_id, node_id, check_cmd, apply_cmd, desired_exit_code, created_at, updated_at, last_checked, status, last_error
			FROM desired_states
			WHERE status = $1 AND ABS(hashtext(node_id) % $2) = $3
		`
		args = []interface{}{status, shardCount, shardIndex}
	} else {
		query = `
			SELECT state_id, node_id, check_cmd, apply_cmd, desired_exit_code, created_at, updated_at, last_checked, status, last_error
			FROM desired_states WHERE status = $1
		`
		args = []interface{}{status}
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []*DesiredState
	for rows.Next() {
		var state DesiredState
		err := rows.Scan(
			&state.StateID, &state.NodeID, &state.CheckCmd, &state.ApplyCmd,
			&state.DesiredExitCode, &state.CreatedAt, &state.UpdatedAt,
			&state.LastChecked, &state.Status, &state.LastError,
		)
		if err != nil {
			return nil, err
		}
		states = append(states, &state)
	}
	return states, nil
}

func (s *PostgresStore) CountStatesByStatus(ctx context.Context, tenantID string, status string) (int, error) {
	query := `SELECT COUNT(*) FROM desired_states WHERE tenant_id = $1 AND status = $2`
	var count int
	err := s.pool.QueryRow(ctx, query, tenantID, status).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *PostgresStore) ListJobs(ctx context.Context, tenantID string, nodeID string, limit int) ([]*Job, error) {
	query := `
		SELECT job_id, node_id, tenant_id, state_id, command, status, exit_code, stdout, stderr, trace_id, created_at, started_at, finished_at
		FROM jobs WHERE node_id = $1 ORDER BY created_at DESC LIMIT $2
	`
	rows, err := s.pool.Query(ctx, query, nodeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(
			&j.JobID, &j.NodeID, &j.TenantID, &j.StateID, &j.Command, &j.Status,
			&j.ExitCode, &j.Stdout, &j.Stderr, &j.TraceID, &j.CreatedAt, &j.StartedAt, &j.FinishedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, &j)
	}
	return jobs, nil
}

func (s *PostgresStore) ListJobsByTenant(ctx context.Context, tenantID string, limit int) ([]*Job, error) {
	query := `
		SELECT job_id, node_id, tenant_id, state_id, command, status, exit_code, stdout, stderr, trace_id, created_at, started_at, finished_at
		FROM jobs WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2
	`
	rows, err := s.pool.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(
			&j.JobID, &j.NodeID, &j.TenantID, &j.StateID, &j.Command, &j.Status,
			&j.ExitCode, &j.Stdout, &j.Stderr, &j.TraceID, &j.CreatedAt, &j.StartedAt, &j.FinishedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, &j)
	}
	return jobs, nil
}

// --- Coordination Operations ---

func (s *PostgresStore) IncrementDurableEpoch(ctx context.Context, resourceID string) (int64, error) {
	// Atomic UPSERT to increment epoch
	query := `
		INSERT INTO leader_epochs (resource_id, epoch)
		VALUES ($1, 1)
		ON CONFLICT (resource_id) DO UPDATE
		SET epoch = leader_epochs.epoch + 1
		RETURNING epoch
	`
	var newEpoch int64
	err := s.pool.QueryRow(ctx, query, resourceID).Scan(&newEpoch)
	if err != nil {
		return 0, err
	}
	return newEpoch, nil
}

func (s *PostgresStore) GetDurableEpoch(ctx context.Context, resourceID string) (int64, error) {
	query := `SELECT epoch FROM leader_epochs WHERE resource_id = $1`
	var epoch int64
	err := s.pool.QueryRow(ctx, query, resourceID).Scan(&epoch)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil // Default to 0 if not exists
	}
	if err != nil {
		return 0, err
	}
	return epoch, nil
}

// --- Idempotency Operations ---

// GetIdempotencyRecord retrieves a cached idempotency response
// Note: Postgres is not ideal for idempotency caching (use Redis instead)
// This implementation is for completeness
func (s *PostgresStore) GetIdempotencyRecord(key string) (string, error) {
	// Not implemented in Postgres - should use Redis for idempotency
	return "", errors.New("not found")
}

// SetIdempotencyRecord stores an idempotency response
// Note: Postgres is not ideal for idempotency caching (use Redis instead)
// This implementation is for completeness
func (s *PostgresStore) SetIdempotencyRecord(key string, value string, ttl time.Duration) error {
	// Not implemented in Postgres - should use Redis for idempotency
	return nil
}
