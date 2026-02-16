package store

import (
	"fmt"
)

// Resource type for Redis keys
type Resource string

const (
	ResourceAgent Resource = "agents"
	ResourceJob   Resource = "jobs"
	ResourceState Resource = "states"
)

// TenantKey constructs a fully qualified Redis key for a tenant resource.
// Format: fluxforge:tenants:{tenantID}:{resource}:{id}
func TenantKey(tenantID string, resource Resource, id string) string {
	return fmt.Sprintf("fluxforge:tenants:%s:%s:%s", tenantID, resource, id)
}

// TenantPrefix constructs a search pattern prefix for a tenant resource.
// Format: fluxforge:tenants:{tenantID}:{resource}:
func TenantPrefix(tenantID string, resource Resource) string {
	return fmt.Sprintf("fluxforge:tenants:%s:%s:", tenantID, resource)
}
