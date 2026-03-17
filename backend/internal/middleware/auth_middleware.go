package middleware

import (
	"aituber/pkg/jwtutil"
	"aituber/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT and sets user_id in context
type AuthMiddleware struct {
	jwt *jwtutil.Manager
}

func NewAuthMiddleware(jwt *jwtutil.Manager) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header must be Bearer token")
			c.Abort()
			return
		}

		claims, err := m.jwt.Verify(parts[1])
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.UserEmail)
		c.Next()
	}
}
