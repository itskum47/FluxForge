package middleware

import (
	"context"
	"fmt"
	"net/http"
)

// TenantContextKey is a strict type for context keys to prevent collisions.
type TenantContextKey string

const (
	// TenantKey is the context key for the TenantID.
	TenantKey TenantContextKey = "tenant_id"
	// TenantHeader is the HTTP header expected to contain the TenantID.
	TenantHeader = "X-Tenant-ID"
)

// TenantMiddleware extracts the TenantID from the request header and injects it into the context.
// It returns a 400 Bad Request if the header is missing.
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get(TenantHeader)

		if tenantID == "" {
			http.Error(w, fmt.Sprintf("Missing required header: %s", TenantHeader), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), TenantKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTenantFromContext safely retrieves the TenantID from the context.
func GetTenantFromContext(ctx context.Context) (string, error) {
	val := ctx.Value(TenantKey)
	if val == nil {
		return "", fmt.Errorf("tenant_id not found in context")
	}

	tenantID, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("tenant_id in context is not a string")
	}

	return tenantID, nil
}
