#!/bin/bash
# Phase 8: Restore Procedure
# Usage: ./scripts/phase8_restore.sh [backup_file_prefix]
# Example: ./scripts/restore.sh backups/dump_20240218_120000

set -e

BACKUP_PATH=$1
CONTAINER="fluxforge-redis"
LOG_FILE="restore.log"

if [ -z "$BACKUP_PATH" ]; then
    echo "Usage: $0 <path_to_rdb_or_backup_prefix>"
    exit 1
fi

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a $LOG_FILE
}

log "Starting Restore Process from $BACKUP_PATH..."

# Check if file exists
if [ ! -f "$BACKUP_PATH.rdb" ] && [ ! -f "$BACKUP_PATH" ]; then
    log "Error: Backup file not found: $BACKUP_PATH"
    exit 1
fi

# Determine source files
if [[ "$BACKUP_PATH" == *".rdb" ]]; then
    RDB_FILE="$BACKUP_PATH"
    # Try to guess AOF dir name (replace dump_ with appendonlydir_ and remove .rdb)
    # This is a heuristic, in production be more explicit
    DIR_NAME=$(dirname "$BACKUP_PATH")
    BASE_NAME=$(basename "$BACKUP_PATH" .rdb)
    AOF_DIR="$DIR_NAME/${BASE_NAME/dump_/appendonlydir_}"
else
    # Assuming prefix provided
    RDB_FILE="${BACKUP_PATH}.rdb"
    DIR_NAME=$(dirname "$BACKUP_PATH")
    BASE_NAME=$(basename "$BACKUP_PATH")
    AOF_DIR="$DIR_NAME/${BASE_NAME/dump_/appendonlydir_}"
fi

log "Target RDB: $RDB_FILE"
log "Target AOF Dir: $AOF_DIR"

# 1. Stop Redis Container (if running)
log "Stopping Redis Container..."
docker stop $CONTAINER || true

# 2. Cleanup existing data
log "Cleaning up existing Redis data..."
docker run --rm -v fluxforge_redis_data:/data alpine sh -c "rm -rf /data/*"

# 3. Restore Data
# We use a helper container to mount the volume and copy files
log "Restoring Data..."

# Copy RDB
docker run --rm -v fluxforge_redis_data:/data -v "$(pwd)/$RDB_FILE":/backup/dump.rdb alpine cp /backup/dump.rdb /data/dump.rdb

# Copy AOF Dir (if exists)
if [ -d "$AOF_DIR" ]; then
    log "Restoring AOF Directory..."
    docker run --rm -v fluxforge_redis_data:/data -v "$(pwd)/$AOF_DIR":/backup/appendonlydir alpine cp -r /backup/appendonlydir /data/
else
    log "WARNING: AOF Directory not found. Restoring from RDB only."
fi

# 4. Fix Permissions
log "Fixing Permissions..."
docker run --rm -v fluxforge_redis_data:/data alpine chown -R 999:999 /data

# 5. Start Redis
log "Starting Redis Container..."
docker start $CONTAINER

log "Restore Process COMPLETE. Verifying..."
sleep 5
docker exec $CONTAINER redis-cli ping
