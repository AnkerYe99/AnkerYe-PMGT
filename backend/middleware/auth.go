package middleware

import (
	"ankerye-pmgt/db"
	"ankerye-pmgt/util"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

func GetJWTSecret() []byte {
	return jwtSecret
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			util.Fail(c, http.StatusUnauthorized, "未授权")
			c.Abort()
			return
		}

		// Try API Key first
		if strings.HasPrefix(token, "ak_") {
			user, err := validateAPIKey(token)
			if err != nil {
				util.Fail(c, http.StatusUnauthorized, "API Key 无效")
				c.Abort()
				return
			}
			c.Set("currentUser", user)
			c.Next()
			return
		}

		// JWT validation
		claims := jwt.MapClaims{}
		t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !t.Valid {
			util.Fail(c, http.StatusUnauthorized, "Token 无效或已过期")
			c.Abort()
			return
		}

		uid, _ := claims["uid"].(float64)
		username, _ := claims["username"].(string)
		role, _ := claims["role"].(string)

		c.Set("currentUser", &util.User{
			ID:       int64(uid),
			Username: username,
			Role:     role,
		})
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := util.CurrentUser(c)
		if user == nil {
			util.Fail(c, http.StatusUnauthorized, "未授权")
			c.Abort()
			return
		}
		for _, r := range roles {
			if user.Role == r {
				c.Next()
				return
			}
		}
		util.Fail(c, http.StatusForbidden, "权限不足")
		c.Abort()
	}
}

func RequireDevOrAbove() gin.HandlerFunc {
	return RequireRole("admin", "developer")
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return c.Query("token")
}

func validateAPIKey(key string) (*util.User, error) {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
	row := db.DB.QueryRow(`
		SELECT ak.id, ak.created_by, ak.permissions, u.username, u.role
		FROM api_keys ak JOIN users u ON u.id=ak.created_by
		WHERE ak.key_hash=? AND ak.status='active' AND (ak.expires_at IS NULL OR ak.expires_at > CURRENT_TIMESTAMP)
	`, hash)
	var akID int64
	var uid int64
	var perms, username, role string
	if err := row.Scan(&akID, &uid, &perms, &username, &role); err != nil {
		return nil, err
	}
	db.DB.Exec("UPDATE api_keys SET last_used_at=CURRENT_TIMESTAMP WHERE id=?", akID)
	return &util.User{ID: uid, Username: username, Role: role}, nil
}
