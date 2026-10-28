// middleware/auth_middleware.go
package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil header Authorization (formatnya: "Bearer <token>")
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Token tidak ditemukan"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Format header Authorization salah"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		secret := []byte(os.Getenv("JWT_SECRET_KEY"))
		if len(secret) == 0 {
			secret = []byte("rahasia_perusahaan_bmp_123!") // Samakan dengan di controller
		}

		// Validasi dan parsing token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode signing token tidak valid")
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi telah berakhir atau token tidak valid"})
			c.Abort()
			return
		}

		// Simpan klaim (data user) ke dalam context agar bisa dibaca oleh controller
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
			c.Set("username", claims["username"])
			c.Set("role", claims["role"])
			c.Set("permissions", claims["permissions"])
		}

		c.Next()
	}
}