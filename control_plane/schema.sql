-- FluxForge Schema

CREATE TABLE IF NOT EXISTS agents (
    node_id VARCHAR(64) PRIMARY KEY,
    hostname VARCHAR(255),
    ip_address VARCHAR(64),
    status VARCHAR(32),
    last_heartbeat TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS jobs (
    job_id VARCHAR(64) PRIMARY KEY,
    node_id VARCHAR(64),
    command TEXT,
    status VARCHAR(32),
    exit_code INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS desired_states (
    state_id VARCHAR(64) PRIMARY KEY,
    node_id VARCHAR(64),
    spec JSONB,
    status VARCHAR(32),
    version BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
