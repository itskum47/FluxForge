package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/itskum47/FluxForge/control_plane/auth"
)

// Context keys
const (
	RoleContextKey   TenantContextKey = "role"
	ClaimsContextKey TenantContextKey = "claims"
)

// AuthMiddleware enforces JWT authentication on requests.
// STRICT: Fails fast on missing or malformed headers.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// STRICT: Fail fast if missing
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// STRICT: Validate format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization format. Expected 'Bearer <token>'", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate Token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
			return
		}

		// Inject Claims into Context
		ctx := context.WithValue(r.Context(), TenantKey, claims.TenantID)
		ctx = context.WithValue(ctx, RoleContextKey, claims.Role)
		ctx = context.WithValue(ctx, ClaimsContextKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRoleFromContext retrieves the role from the context.
func GetRoleFromContext(ctx context.Context) (string, error) {
	val := ctx.Value(RoleContextKey)
	if val == nil {
		return "", fmt.Errorf("role not found in context")
	}
	role, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("role in context is not a string")
	}
	return role, nil
}
