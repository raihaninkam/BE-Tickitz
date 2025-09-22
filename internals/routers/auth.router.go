package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raihaninkam/tickitz/internals/handlers"
	"github.com/raihaninkam/tickitz/internals/middlewares"
	"github.com/raihaninkam/tickitz/internals/repositories"
)

func InitAuthRouter(router *gin.Engine, db *pgxpool.Pool) {
	authRouter := router.Group("/auth")

	authRepository := repositories.NewAuthRepository(db)
	authHandler := handlers.NewAuthHandler(authRepository)

	authRouter.POST("/login", authHandler.Login)
	authRouter.POST("/register", authHandler.Register)
	authRouter.POST("/logout", middlewares.VerifyToken, middlewares.Access("user", "admin"), authHandler.SecureLogout)
}
