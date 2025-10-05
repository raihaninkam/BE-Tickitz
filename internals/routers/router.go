package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	docs "github.com/raihaninkam/tickitz/docs"
	"github.com/raihaninkam/tickitz/internals/middlewares"
	"github.com/raihaninkam/tickitz/pkg"
	"github.com/redis/go-redis/v9"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter(db *pgxpool.Pool, rdb *redis.Client, hc *pkg.HashConfig) *gin.Engine {
	router := gin.Default()

	router.Static("/images", "./public/images")

	router.Use(middlewares.CORSMiddleware)

	InitAuthRouter(router, db)

	InitMovieRouter(router, db, rdb)

	InitOrderRouter(router, db)

	InitProfileRouter(router, db, hc)

	InitAdminMovieRouter(router, db, rdb)

	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"Message": "Rute Salah",
			"Status":  "Rute Tidak Ditemukan",
		})
	})
	return router

}
