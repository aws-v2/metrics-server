package middleware

import (
	"context"
	"fmt"
	"metrics-gateway/internal/utils"
	"log"
	"net/http"
	"strings"
)

type contextKey string

const (
	ContextKeyUserID     contextKey = "userId"
	ContextKeyRole       contextKey = "userRole"
	ContextKeyRequestId  contextKey = "requestId"
	ContextKeyAuthMethod contextKey = "authMethod"
)

type AuthMiddleware struct {
	Token string
}

func NewAuthMiddleware(token string) *AuthMiddleware {
	return &AuthMiddleware{
		Token: token,
	}
}

// func AuthContextMiddleware((next http.HandlerFunc))  http.HandlerFunc {
func Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-User-Role")
		authMethod := r.Header.Get("X-Auth-Method")
		requestID := r.Header.Get("X-Request-Id")

		headers := make([]string, 4, 4)
		headers[0] = userID
		headers[1] = role
		headers[2] = authMethod
		headers[3] = requestID

		// Inject claims into request context for downstream handlers.
		ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
		ctx = context.WithValue(ctx, ContextKeyRole, role)
		ctx = context.WithValue(ctx, ContextKeyRequestId, requestID)
		ctx = context.WithValue(ctx, ContextKeyAuthMethod, authMethod)

		// TODO: this solution works but its inelegant, change this to somehtign better

		if strings.Contains(requestID, "public-req") {
			next(w, r.WithContext(ctx))

		} else {
			for iter, header := range headers {

				if header == "" {
					log.Printf("[Middleware] header %v not available, droping request", iter)
					utils.WriteJSONError(w, http.StatusUnauthorized, fmt.Errorf("Missing headers you dumb f*ck"))
					next(w, r.WithContext(ctx))

					return
				}

			}
		}

		next(w, r.WithContext(ctx))
	}
}
