package transport

import (
	"context"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is an unexported type for context keys in this package.
type contextKey string

const (
	// ContextKeyUserID is the context key for the authenticated user ID.
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyRole is the context key for the authenticated role.
	ContextKeyRole contextKey = "role"
)

// AuthMiddleware validates JWT tokens on incoming requests.
// It expects:
//   - Authorization: Bearer <token>
//   - Token signed with the provided secret (HMAC, hex-decoded)
//   - Claims: role="instance", user_id=<string>
type AuthMiddleware struct {
	logger *slog.Logger
	secret []byte
}

// NewAuthMiddleware creates a new AuthMiddleware.
// The secret is expected as a hex-encoded string (matching the IAM service)
// and will be decoded to raw bytes for HMAC verification.
func NewAuthMiddleware(logger *slog.Logger, secretHex string) *AuthMiddleware {
	secret, err := hex.DecodeString(secretHex)
	if err != nil {
		// Fall back to raw bytes if not valid hex.
		logger.Warn("INSTANCE_TOKEN_SECRET is not valid hex, using raw bytes", "error", err)
		secret = []byte(secretHex)
	}
	return &AuthMiddleware{
		logger: logger,
		secret: secret,
	}
}

// Wrap returns an http.HandlerFunc that validates the JWT before calling next.
func (m *AuthMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Warn("auth: missing Authorization header",
				"method", r.Method,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
			)
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing Authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			m.logger.Warn("auth: invalid Authorization header format",
				"method", r.Method,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
			)
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid Authorization header format"})
			return
		}
		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			// Ensure we only accept HMAC signing methods.
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.secret, nil
		})
		if err != nil || !token.Valid {
			m.logger.Warn("auth: invalid JWT token",
				"error", err,
				"method", r.Method,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
			)
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			m.logger.Warn("auth: invalid token claims",
				"method", r.Method,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
			)
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
			return
		}

		userID, _ := claims["user_id"].(string)
		role, _ := claims["role"].(string)

		if userID == "" || role != "instance" {
			m.logger.Warn("auth: insufficient permissions",
				"user_id", userID,
				"role", role,
				"method", r.Method,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
			)
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
			return
		}

		m.logger.Info("auth: request authorized",
			"user_id", userID,
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
		)

		// Inject claims into request context for downstream handlers.
		ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
		ctx = context.WithValue(ctx, ContextKeyRole, role)

		next(w, r.WithContext(ctx))
	}
}
