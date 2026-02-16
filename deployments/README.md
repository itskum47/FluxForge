# FluxForge Phase 7 Production Hardening

This directory contains deployment configurations and automation scripts for Phase 7 production hardening.

## Directory Structure

```
deployments/
├── docker-compose.yml          # Multi-node Docker Compose deployment
├── production.env              # Production configuration values
└── kubernetes/
    └── fluxforge.yaml         # Kubernetes manifests
```

## Quick Start

### Docker Compose Deployment

```bash
# Start 3-node cluster
cd deployments
docker-compose up -d

# Verify cluster health
docker-compose ps

# View logs
docker-compose logs -f control-1

# Stop cluster
docker-compose down
```

### Kubernetes Deployment

```bash
# Deploy to Kubernetes
kubectl apply -f kubernetes/fluxforge.yaml

# Verify deployment
kubectl get pods -n fluxforge

# Get API endpoint
kubectl get svc -n fluxforge fluxforge-api

# View logs
kubectl logs -n fluxforge -l app=fluxforge-control-plane -f
```

## Testing Scripts

Located in `../scripts/`:

- `phase7_multinode_test.sh` - Multi-node deployment validation
- `phase7_stability_test.sh` - Long-duration stability testing
- `phase7_chaos_monkey.sh` - Fault injection testing

### Run Multi-Node Tests

```bash
cd ..
./scripts/phase7_multinode_test.sh
```

### Run Stability Test (24 hours)

```bash
./scripts/phase7_stability_test.sh 24
```

### Run Chaos Testing (1 hour)

```bash
./scripts/phase7_chaos_monkey.sh 60
```

## Production Configuration

See `production.env` for all production configuration values.

Key settings:
- Control plane replicas: 3
- Max WebSocket connections: 200
- Scheduler queue depth: 10,000
- Leader election timeout: 15s
- Heartbeat timeout: 30s

## Monitoring

### Metrics Endpoints

- Prometheus: `http://<node>:9090/metrics`
- Health check: `http://<node>:8080/health`
- Readiness: `http://<node>:8080/ready`

### Key Metrics

```
flux_scheduler_queue_depth
flux_scheduler_active_tasks
flux_agents_total{status="active"}
flux_leader_is_leader
flux_jobs_total{status="completed"}
```

## Troubleshooting

### Check Leader Status

```bash
curl http://localhost:8080/api/dashboard | jq '{is_leader, node_id, current_epoch}'
```

### View Cluster Topology

```bash
curl http://localhost:8080/api/clusters | jq .
```

### Check Database Connection

```bash
docker exec fluxforge-postgres psql -U fluxforge -c "SELECT COUNT(*) FROM agents;"
```

### View Container Logs

```bash
docker logs fluxforge-control-1 --tail 100 -f
```

## Production Checklist

Before deploying to production:

- [ ] Update `DB_PASSWORD` in secrets
- [ ] Enable TLS (`TLS_ENABLED=true`)
- [ ] Configure backup retention
- [ ] Set up monitoring alerts
- [ ] Configure log aggregation
- [ ] Review resource limits
- [ ] Test disaster recovery
- [ ] Run full test suite
- [ ] Document runbooks

## Support

For issues or questions, see the main project README.
