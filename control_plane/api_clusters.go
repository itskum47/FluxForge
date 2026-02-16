package main

import (
	"encoding/json"
	"net/http"
	"os"
)

// ClusterInfo represents information about a cluster node.
type ClusterInfo struct {
	ClusterID   string  `json:"cluster_id"`
	Region      string  `json:"region"`
	Role        string  `json:"role"` // leader, follower, standby
	IsLeader    bool    `json:"is_leader"`
	AgentCount  int     `json:"agent_count"`
	HealthScore float64 `json:"health_score"`
	Endpoint    string  `json:"endpoint"`
}

// handleGetClusters returns information about all cluster nodes.
func (a *API) handleGetClusters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Get current node info
	var leaderState bool
	if a.elector != nil {
		leaderState = a.elector.GetState().IsLeader
	}

	agents, _ := a.store.ListAgents(ctx, "")

	// For single-node deployment, return just this node
	// In multi-cluster, this would query a service registry
	clusters := []ClusterInfo{
		{
			ClusterID: getEnvOrDefault("CLUSTER_ID", "cluster-primary"),
			Region:    getEnvOrDefault("REGION", "us-east-1"),
			Role: func() string {
				if leaderState {
					return "leader"
				}
				return "follower"
			}(),
			IsLeader:    leaderState,
			AgentCount:  len(agents),
			HealthScore: 0.95, // TODO: Calculate from metrics
			Endpoint:    "http://localhost:8080",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(clusters)
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
