package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/raihaninkam/tickitz/internals/repositories"
)

// all movie

type AllMovie struct {
	am *repositories.AllMovie
}

func NewAllMovie(am *repositories.AllMovie) *AllMovie {
	return &AllMovie{am: am}
}

// GetAllMovies godoc
// @Summary     Get All Movies
// @Description Mengambil semua data movie untuk general user. Endpoint ini tidak memerlukan otorisasi.
// @Tags        Movies
// @Accept      json
// @Produce     json
// @Success     200 {object} map[string]interface{} "Berhasil mengambil semua data movies dengan total count"
// @Failure     404 {object} map[string]interface{} "Not Found - Tidak ada film yang ditemukan"
// @Failure     500 {object} map[string]interface{} "Internal Server Error"
// @Router      /movies [get]
func (h *AllMovie) GetAllMovies(ctx *gin.Context) {
	movies, err := h.am.GetAllMovies(ctx.Request.Context())
	if err != nil {
		if err.Error() == "no movies found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tidak ada film ditemukan",
			})
			return
		}
		log.Println("Internal Server Error.\nCause: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movies,
		"total":   len(movies),
	})
}

// upcoming movie

type UpcomingMovieHandler struct {
	umr *repositories.UpcomingMovie
}

func NewUpcomingMovieHandler(umr *repositories.UpcomingMovie) *UpcomingMovieHandler {
	return &UpcomingMovieHandler{umr: umr}
}

// GetUpcomingMovies godoc
// @Summary     Get Upcoming Movies
// @Description Mengambil daftar semua film yang akan segera rilis atau upcoming movies
// @Tags        Movies
// @Accept      json
// @Produce     json
// @Success     200 {object} map[string]interface{} "Berhasil mengambil daftar upcoming movies"
// @Failure     404 {object} map[string]interface{} "Not Found - Tidak ada film upcoming yang ditemukan"
// @Failure     500 {object} map[string]interface{} "Internal Server Error"
// @Router      /movies/upcoming [get]
func (u *UpcomingMovieHandler) GetUpcomingMovies(ctx *gin.Context) {
	movies, err := u.umr.GetUpcomingMovies(ctx.Request.Context())
	if err != nil {
		if err.Error() == "no upcoming movies found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tidak ada film upcoming",
			})
			return
		}
		log.Println("Internal Server Error.\nCause: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movies,
	})
}

// popular movie

type PopularMovieHandler struct {
	pmr *repositories.PopularMovie
}

func NewPopularMovieHandler(pmr *repositories.PopularMovie) *PopularMovieHandler {
	return &PopularMovieHandler{pmr: pmr}
}

// GetPopularMovies godoc
// @Summary     Get Popular Movies
// @Description Mengambil daftar semua film yang sedang populer berdasarkan rating atau jumlah penonton
// @Tags        Movies
// @Accept      json
// @Produce     json
// @Success     200 {object} map[string]interface{} "Berhasil mengambil daftar popular movies"
// @Failure     404 {object} map[string]interface{} "Not Found - Tidak ada film popular yang ditemukan"
// @Failure     500 {object} map[string]interface{} "Internal Server Error"
// @Router      /movies/popular [get]
func (p *PopularMovieHandler) GetPopularMovies(ctx *gin.Context) {
	movies, err := p.pmr.GetPopularMovies(ctx.Request.Context())
	if err != nil {
		if err.Error() == "no popular movies found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tidak ada film popular",
			})
			return
		}
		log.Println("Internal Server Error.\nCause: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movies,
	})
}

// movie filter

type MovieFilterHandler struct {
	mf *repositories.MovieFilter
}

func NewMovieFilterHandler(mf *repositories.MovieFilter) *MovieFilterHandler {
	return &MovieFilterHandler{mf: mf}
}

// GetMoviesWithFilter godoc
// @Summary     Get Movies With Filter
// @Description Mengambil daftar film dengan filter opsional: judul, genre, dan pagination.
// @Tags        Movies
// @Accept      json
// @Produce     json
// @Param       title     query    string   false  "Filter berdasarkan judul (opsional)"
// @Param       genre     query    string   false  "Filter berdasarkan genre, pisahkan dengan koma jika lebih dari satu (opsional)"
// @Param       page      query    int      false  "Halaman yang ingin diambil (default: 1)"
// @Success     200 {object} map[string]interface{} "Berhasil mengambil daftar film"
// @Failure     404 {object} map[string]interface{} "Not Found - Tidak ada film yang ditemukan"
// @Failure     500 {object} map[string]interface{} "Internal Server Error"
// @Router      /movies/filter [get]
func (h *MovieFilterHandler) GetMoviesWithFilter(ctx *gin.Context) {
	// Ambil query params
	title := ctx.Query("title")
	genreQuery := ctx.Query("genre")
	genres := []string{}
	if genreQuery != "" {
		genres = strings.Split(genreQuery, ",")
	}

	// Pagination
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "12"))
	if err != nil || limit <= 0 {
		limit = 12
	}
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// Panggil repository
	movies, totalCount, err := h.mf.GetMoviesWithFilter(ctx.Request.Context(), title, genres, offset, limit)
	if err != nil {
		if err.Error() == "no movies found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tidak ada film ditemukan",
			})
			return
		}
		log.Println("Internal Server Error.\nCause: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	// Response sukses
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movies,
		"page":    page,
		"limit":   limit,
		"count":   totalCount,
	})
}

// movie detail

type MovieDetailHandler struct {
	mr *repositories.MovieDetail
}

func NewMovieDetailHandler(mr *repositories.MovieDetail) *MovieDetailHandler {
	return &MovieDetailHandler{mr: mr}
}

// GetDetailMovie godoc
// @Summary     Get Movie Detail
// @Description Mengambil detail film berdasarkan ID
// @Tags        Movies
// @Accept      json
// @Produce     json
// @Param       movie_id  path      int  true  "ID Film"
// @Success     200 {object} map[string]interface{} "Berhasil mengambil detail film"
// @Failure     400 {object} map[string]interface{} "Bad Request - movie_id tidak valid"
// @Failure     404 {object} map[string]interface{} "Not Found - Film tidak ditemukan"
// @Failure     500 {object} map[string]interface{} "Internal Server Error"
// @Router      /movies/{movie_id} [get]
func (m *MovieDetailHandler) GetDetailMovie(ctx *gin.Context) {
	// Ambil movie_id dari URL parameter
	movieIDStr := ctx.Param("movie_id")
	if movieIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "movie_id harus diisi",
		})
		return
	}

	// Konversi string ke int
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "movie_id harus berupa angka",
		})
		return
	}

	// Ambil data detail movie dari repository
	movie, err := m.mr.GetDetailMovie(ctx.Request.Context(), movieID)
	if err != nil {
		if strings.Contains(err.Error(), "movie not found") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Film tidak ditemukan",
			})
			return
		}
		log.Println("Internal Server Error.\nCause: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	// Response sukses
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Berhasil mengambil detail film",
		"data":    movie,
	})
}

// schedule

type ScheduleHandler struct {
	sr *repositories.Schedule
}

func NewScheduleHandler(sr *repositories.Schedule) *ScheduleHandler {
	return &ScheduleHandler{sr: sr}
}

// GetSchedulesByMovieID godoc
// @Summary     Get schedules by movie ID
// @Description Mengambil semua data jadwal film berdasarkan movie ID
// @Tags        Movies
// @Produce     json
// @Param       movie_id path int true "ID Film"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{} "ID film tidak valid"
// @Failure     404 {object} map[string]interface{} "Tidak ada jadwal untuk film tersebut"
// @Failure     500 {object} map[string]interface{} "Internal server error"
// @Router      /movies/schedule/{movie_id} [get]
func (s *ScheduleHandler) GetSchedulesByMovieID(ctx *gin.Context) {
	movieIDStr := ctx.Param("movie_id")

	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID film tidak valid",
		})
		return
	}

	schedules, err := s.sr.GetSchedulesByMovieID(ctx.Request.Context(), movieID)
	if err != nil {
		if err.Error() == "no schedules found for this movie" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tidak ada jadwal untuk film tersebut",
			})
			return
		}
		log.Println("Internal Server Error.\nCause: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    schedules,
	})
}
