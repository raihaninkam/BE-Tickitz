package middlewares

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/raihaninkam/tickitz/pkg"
)

func VerifyToken(ctx *gin.Context) {
	// Ambil token dari header Authorization
	bearerToken := ctx.GetHeader("Authorization")
	if bearerToken == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Authorization header tidak ditemukan",
		})
		return
	}

	// Pastikan formatnya "Bearer <token>"
	parts := strings.SplitN(bearerToken, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Format Authorization harus 'Bearer <token>'",
		})
		return
	}

	token := parts[1]
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Token kosong",
		})
		return
	}

	// Verify token
	var claims pkg.Claims
	if err := claims.VerifyToken(token); err != nil {
		switch {
		case strings.Contains(err.Error(), jwt.ErrTokenInvalidIssuer.Error()):
			log.Println("JWT Error. Cause:", err.Error())
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Issuer tidak valid, silahkan login kembali",
			})
			return
		case strings.Contains(err.Error(), jwt.ErrTokenExpired.Error()):
			log.Println("JWT Error. Cause:", err.Error())
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Token expired, silahkan login kembali",
			})
			return
		default:
			log.Println("Internal Server Error. Cause:", err.Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			return
		}
	}

	// Simpan claims ke context supaya bisa dipakai di handler
	ctx.Set("claims", claims)
	ctx.Next()
}
