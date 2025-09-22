package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raihaninkam/tickitz/internals/handlers"
	"github.com/raihaninkam/tickitz/internals/middlewares"
	"github.com/raihaninkam/tickitz/internals/repositories"
	"github.com/raihaninkam/tickitz/pkg"
)

func InitProfileRouter(router *gin.Engine, db *pgxpool.Pool, hc *pkg.HashConfig) {
	profileRouter := router.Group("/profile")

	authRepo := repositories.NewAuthRepository(db)
	// profileRouter.Use(middlewares.JWTMiddlewareWithBlacklist(authRepo))

	profileRepository := repositories.NewProfileRepository(db, hc)
	profileHandler := handlers.NewProfileHandler(profileRepository)

	// GET my profile
	profileRouter.GET("", middlewares.JWTMiddlewareWithBlacklist(authRepo),
		middlewares.VerifyToken,
		middlewares.Access("user"),
		profileHandler.GetMyProfile,
	)

	// PATCH update profile (ambil userId dari JWT, bukan param)
	profileRouter.PATCH("", middlewares.JWTMiddlewareWithBlacklist(authRepo),
		middlewares.VerifyToken,
		middlewares.Access("user"),
		profileHandler.UpdateProfileWithImage,
	)

	// PATCH change password (ambil userId dari JWT, bukan param)
	profileRouter.PATCH("/change-password", profileHandler.ChangePassword)
}
