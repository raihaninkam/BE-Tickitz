package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raihaninkam/tickitz/internals/handlers"
	"github.com/raihaninkam/tickitz/internals/middlewares"

	"github.com/raihaninkam/tickitz/internals/repositories"
)

func InitOrderRouter(router *gin.Engine, db *pgxpool.Pool) {
	orderRouter := router.Group("/orders")

	authRepo := repositories.NewAuthRepository(db)
	// router.Use(middlewares.JWTMiddlewareWithBlacklist(authRepo))

	// order
	orderRepo := repositories.NewOrderRepository(db)
	orderHandler := handlers.NewOrderHandler(orderRepo, &repositories.SeatsRepository{})

	orderRouter.POST("", middlewares.JWTMiddlewareWithBlacklist(authRepo), middlewares.VerifyToken, middlewares.Access("user"), orderHandler.CreateOrder)

	// seat avail
	seatsRepository := repositories.NewSeatsRepository(db)
	seatsHandler := handlers.NewSeatsHandler(seatsRepository)

	orderRouter.GET("/seats/:now_showing_id", middlewares.JWTMiddlewareWithBlacklist(authRepo), middlewares.VerifyToken, middlewares.Access("user"), seatsHandler.GetAvailableSeats)

	orderHistoryRepository := repositories.NewOrderHistory(db)
	orderHistoryHandler := handlers.NewOrderHistoryHandler(orderHistoryRepository)

	orderRouter.GET("/history", middlewares.JWTMiddlewareWithBlacklist(authRepo), middlewares.VerifyToken, middlewares.Access("user"), orderHistoryHandler.GetOrderHistory)

}
