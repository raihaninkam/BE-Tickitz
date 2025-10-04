package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raihaninkam/tickitz/internals/handlers"
	"github.com/raihaninkam/tickitz/internals/middlewares"
	"github.com/raihaninkam/tickitz/internals/repositories"
	"github.com/redis/go-redis/v9"
)

func InitAdminMovieRouter(router *gin.Engine, db *pgxpool.Pool, rdb *redis.Client) {
	adminMovieRouter := router.Group("/admin")

	authRepo := repositories.NewAuthRepository(db)

	movieRepo := repositories.NewMovieAdmin(db, rdb)
	movieHandler := handlers.NewMovieAdminHandler(movieRepo)

	// CREATE
	// @Summary      Tambah Movie
	// @Description  Admin menambahkan film baru
	// @Tags         Admin-Movies
	// @Security     BearerToken
	// @Accept       json
	// @Produce      json
	// @Param        body  body      models.MovieAdmin  true  "Movie data"
	// @Success      201   {object}  map[string]interface{}
	// @Failure      400   {object}  map[string]interface{}
	// @Failure      401   {object}  map[string]interface{}
	// @Router       /admin/movies/add [post]
	adminMovieRouter.POST("/movies/add", middlewares.JWTMiddlewareWithBlacklist(authRepo),
		middlewares.VerifyToken,
		middlewares.Access("admin"),
		movieHandler.AddMovie,
	)

	// READ
	// @Summary      List Movies
	// @Description  Ambil semua data movie (yang belum dihapus)
	// @Tags         Admin-Movies
	// @Security     BearerToken
	// @Produce      json
	// @Success      200  {object}  map[string]interface{}
	// @Failure      401  {object}  map[string]interface{}
	// @Router       /admin/movies [get]
	adminMovieRouter.GET("/movies", middlewares.JWTMiddlewareWithBlacklist(authRepo),
		middlewares.VerifyToken,
		middlewares.Access("admin"),
		movieHandler.GetAllMovies,
	)

	// UPDATE
	// @Summary      Update Movie
	// @Description  Admin mengupdate data movie
	// @Tags         Admin-Movies
	// @Security     BearerToken
	// @Accept       json
	// @Produce      json
	// @Param        movieId  path      int              true  "Movie ID"
	// @Param        body     body      models.MovieAdmin  true  "Movie update data"
	// @Success      200      {object}  map[string]interface{}
	// @Failure      400      {object}  map[string]interface{}
	// @Failure      404      {object}  map[string]interface{}
	// @Failure      401      {object}  map[string]interface{}
	// @Router       /admin/movies/{movieId} [patch]
	adminMovieRouter.PATCH("/movies/:movieId", middlewares.JWTMiddlewareWithBlacklist(authRepo),
		middlewares.VerifyToken,
		middlewares.Access("admin"),
		movieHandler.UpdateMovie,
	)

	// DELETE (Soft Delete)
	// @Summary      Hapus Movie
	// @Description  Admin melakukan soft delete (is_deleted = true, deleted_at = timestamp)
	// @Tags         Admin-Movies
	// @Security     BearerToken
	// @Produce      json
	// @Param        movieId  path      int  true  "Movie ID"
	// @Success      200      {object}  map[string]interface{}
	// @Failure      404      {object}  map[string]interface{}
	// @Failure      401      {object}  map[string]interface{}
	// @Router       /admin/movies/delete/{movieId} [delete]
	adminMovieRouter.DELETE("/movies/delete/:movieId", middlewares.JWTMiddlewareWithBlacklist(authRepo),
		middlewares.VerifyToken,
		middlewares.Access("admin"),
		movieHandler.DeleteMovie,
	)
}
