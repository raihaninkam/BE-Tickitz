package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/raihaninkam/tickitz/internals/repositories"
	"github.com/raihaninkam/tickitz/pkg"
)

// Middleware sederhana untuk JWT (tanpa blacklist check)
// func JWTMiddleware() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		// Ambil token dari header
// 		authHeader := ctx.GetHeader("Authorization")
// 		if authHeader == "" {
// 			ctx.JSON(http.StatusUnauthorized, gin.H{
// 				"success": false,
// 				"error":   "Token tidak ditemukan",
// 			})
// 			ctx.Abort()
// 			return
// 		}

// 		// Extract token
// 		tokenParts := strings.Split(authHeader, " ")
// 		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
// 			ctx.JSON(http.StatusUnauthorized, gin.H{
// 				"success": false,
// 				"error":   "Format token tidak valid",
// 			})
// 			ctx.Abort()
// 			return
// 		}

// 		tokenString := tokenParts[1]

// 		// Validasi JWT token menggunakan Claims struct Anda
// 		claims := &pkg.Claims{}
// 		err := claims.VerifyToken(tokenString)
// 		if err != nil {
// 			ctx.JSON(http.StatusUnauthorized, gin.H{
// 				"success": false,
// 				"error":   "Token tidak valid atau sudah expired",
// 			})
// 			ctx.Abort()
// 			return
// 		}

// 		// Set user info ke context
// 		ctx.Set("user_id", claims.UserId)
// 		ctx.Set("user_role", claims.Role)
// 		ctx.Set("token", tokenString)

// 		ctx.Next()
// 	}
// }

// Middleware untuk JWT dengan blacklist check
func JWTMiddlewareWithBlacklist(ar *repositories.AuthRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Ambil token dari header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Token tidak ditemukan",
			})
			ctx.Abort()
			return
		}

		// Extract token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Format token tidak valid",
			})
			ctx.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Cek apakah token sudah di-blacklist
		isBlacklisted, err := ar.IsTokenBlacklisted(ctx.Request.Context(), tokenString)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "internal server error",
			})
			ctx.Abort()
			return
		}

		if isBlacklisted {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Silahkan login kembali",
			})
			ctx.Abort()
			return
		}

		// Validasi JWT token menggunakan Claims struct Anda
		claims := &pkg.Claims{}
		err = claims.VerifyToken(tokenString)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Token tidak valid atau sudah expired",
			})
			ctx.Abort()
			return
		}

		// Set user info ke context
		ctx.Set("user_id", claims.UserId)
		ctx.Set("user_role", claims.Role)
		ctx.Set("token", tokenString)

		ctx.Next()
	}
}
