package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raihaninkam/tickitz/internals/handlers"
	"github.com/raihaninkam/tickitz/internals/repositories"
	"github.com/redis/go-redis/v9"
)

func InitMovieRouter(router *gin.Engine, db *pgxpool.Pool, rdb *redis.Client) {
	movieRouter := router.Group("/movies")

	// all movie
	allMovieRepository := repositories.NewAllMovies(db, rdb)
	allMovieHandler := handlers.NewAllMovie(allMovieRepository)
	movieRouter.GET("", allMovieHandler.GetAllMovies)

	// upcoming movie
	upcomingMovieRepository := repositories.NewUpcomingMovie(db, rdb)
	upcomingMovieHandler := handlers.NewUpcomingMovieHandler(upcomingMovieRepository)
	movieRouter.GET("/upcoming", upcomingMovieHandler.GetUpcomingMovies)

	// popular movie
	popularMovieRepository := repositories.NewPopularMovie(db, rdb)
	popularMovieHandler := handlers.NewPopularMovieHandler(popularMovieRepository)
	movieRouter.GET("/popular", popularMovieHandler.GetPopularMovies)

	// filter movie
	movieFilterRepository := repositories.NewMovieFilter(db)
	movieFilterHandler := handlers.NewMovieFilterHandler(movieFilterRepository)
	movieRouter.GET("/filter", movieFilterHandler.GetMoviesWithFilter)

	// movie detail (by id)
	movieDetailRepository := repositories.NewMovieDetail(db)
	movieDetailHandler := handlers.NewMovieDetailHandler(movieDetailRepository)
	movieRouter.GET("/:movie_id", movieDetailHandler.GetDetailMovie)

	// movie schedule
	scheduleRepository := repositories.NewSchedule(db)
	scheduleHandler := handlers.NewScheduleHandler(scheduleRepository)

	// movieRouter.GET("/schedule:", scheduleHandler.GetSchedulesByDate)
	movieRouter.GET("/schedule/:movie_id", scheduleHandler.GetSchedulesByMovieID)
}
