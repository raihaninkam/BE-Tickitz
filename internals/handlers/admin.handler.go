package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raihaninkam/tickitz/internals/models"
	"github.com/raihaninkam/tickitz/internals/repositories"
	"github.com/raihaninkam/tickitz/internals/utils"
)

type MovieAdminHandler struct {
	mar *repositories.MovieAdmin
}

func NewMovieAdminHandler(mar *repositories.MovieAdmin) *MovieAdminHandler {
	return &MovieAdminHandler{mar: mar}
}

// ======================= CREATE =======================

// AddMovie godoc
// @Summary     Tambah Movie (Admin)
// @Description Tambah data movie baru dengan upload gambar
// @Tags        Admin-Movies
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       title            formData string true  "Judul film"
// @Param       synopsis         formData string true  "Sinopsis film"
// @Param       duration_minutes formData int    true  "Durasi film dalam menit"
// @Param       release_date     formData string true  "Tanggal rilis (YYYY-MM-DD)"
// @Param       directors_id     formData int    true  "ID direktur"
// @Param       rating           formData number false "Rating film"
// @Param       poster_image     formData file   true  "File gambar poster"
// @Param       bg_path          formData file   false "File gambar background"
// @Success     201 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /admin/movies/add [post]
func (h *MovieAdminHandler) AddMovie(ctx *gin.Context) {
	// Parse form data
	title := ctx.PostForm("title")
	synopsis := ctx.PostForm("synopsis")
	durationStr := ctx.PostForm("duration_minutes")
	releaseDateStr := ctx.PostForm("release_date")
	directorsIdStr := ctx.PostForm("directors_id")
	ratingStr := ctx.PostForm("rating")

	// Validate required fields
	if title == "" || synopsis == "" || durationStr == "" || releaseDateStr == "" || directorsIdStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Semua field wajib diisi kecuali rating",
		})
		return
	}

	// Convert string to appropriate types
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Duration harus berupa angka",
		})
		return
	}

	directorsId, err := strconv.Atoi(directorsIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Directors ID harus berupa angka",
		})
		return
	}

	var rating *float64
	if ratingStr != "" {
		ratingVal, err := strconv.ParseFloat(ratingStr, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Rating harus berupa angka",
			})
			return
		}
		rating = &ratingVal
	}

	// Parse release date
	releaseDate, err := time.Parse("2006-01-02", releaseDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Format tanggal harus YYYY-MM-DD",
		})
		return
	}

	// File upload configuration
	uploadConfig := utils.FileUploadConfig{
		MaxSize:     5 * 1024 * 1024, // 5MB
		UploadDir:   "./uploads",
		AllowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
	}

	// Upload poster image (required)
	posterPath, err := utils.UploadImageFile(ctx, "poster_image", "posters", uploadConfig)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Error uploading poster image: " + err.Error(),
		})
		return
	}

	// Upload background image (optional)
	var bgPath *string
	if _, _, err := ctx.Request.FormFile("bg_path"); err == nil {
		bgPathStr, err := utils.UploadImageFile(ctx, "bg_path", "backgrounds", uploadConfig)
		if err != nil {
			utils.DeleteFile(posterPath)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Error uploading background image: " + err.Error(),
			})
			return
		}
		bgPath = &bgPathStr
	}

	// Create movie
	movie, err := h.mar.AddMovie(ctx.Request.Context(), models.MovieAdmin{
		Title:           title,
		Synopsis:        synopsis,
		DurationMinutes: duration,
		ReleaseDate:     releaseDate,
		PosterImage:     &posterPath,
		DirectorsId:     directorsId,
		Rating:          rating,
		BgPath:          bgPath,
	})
	if err != nil {
		// Clean up uploaded files if database operation fails
		utils.DeleteFile(posterPath)
		if bgPath != nil {
			utils.DeleteFile(*bgPath)
		}

		log.Println("AddMovie error:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Film berhasil ditambahkan",
		"data":    movie,
	})
}

// ======================= READ =======================

// GetAllMovies godoc
// @Summary     Get All Movies (Admin)
// @Description Semua data Movie untuk admin
// @Tags        Admin-Movies
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} map[string]interface{}
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /admin/movies [get]
func (h *MovieAdminHandler) GetAllMovies(ctx *gin.Context) {
	movies, err := h.mar.GetAllMovies(ctx.Request.Context())
	if err != nil {
		if err.Error() == "no movies found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tidak ada film ditemukan",
			})
			return
		}
		log.Println("GetAllMovies error:", err.Error())
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

// ======================= UPDATE =======================

// UpdateMovie godoc
// @Summary     Update Movie (Admin)
// @Description Update data film berdasarkan ID dengan upload gambar opsional
// @Tags        Admin-Movies
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       movieId          path     int    true  "Movie ID"
// @Param       title            formData string false "Judul film"
// @Param       synopsis         formData string false "Sinopsis film"
// @Param       duration_minutes formData int    false "Durasi film dalam menit"
// @Param       release_date     formData string false "Tanggal rilis (YYYY-MM-DD)"
// @Param       directors_id     formData int    false "ID direktur"
// @Param       rating           formData number false "Rating film"
// @Param       poster_image     formData file   false "File gambar poster (opsional)"
// @Param       bg_path          formData file   false "File gambar background (opsional)"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /admin/movies/{movieId} [patch]
func (h *MovieAdminHandler) UpdateMovie(ctx *gin.Context) {
	// Get movie ID from path
	movieIdStr := ctx.Param("movieId")
	movieId, err := strconv.Atoi(movieIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid movie ID",
		})
		return
	}

	// Parse form data
	var req models.MovieUpdateRequest

	// Handle text fields
	if title := ctx.PostForm("title"); title != "" {
		req.Title = &title
	}

	if synopsis := ctx.PostForm("synopsis"); synopsis != "" {
		req.Synopsis = &synopsis
	}

	if durationStr := ctx.PostForm("duration_minutes"); durationStr != "" {
		duration, err := strconv.Atoi(durationStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Duration harus berupa angka",
			})
			return
		}
		req.DurationMinutes = &duration
	}

	if releaseDateStr := ctx.PostForm("release_date"); releaseDateStr != "" {
		req.ReleaseDate = &releaseDateStr
	}

	if directorsIdStr := ctx.PostForm("directors_id"); directorsIdStr != "" {
		directorsId, err := strconv.Atoi(directorsIdStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Directors ID harus berupa angka",
			})
			return
		}
		req.DirectorsId = &directorsId
	}

	if ratingStr := ctx.PostForm("rating"); ratingStr != "" {
		rating, err := strconv.ParseFloat(ratingStr, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Rating harus berupa angka",
			})
			return
		}
		req.Rating = &rating
	}

	// File upload configuration
	uploadConfig := utils.FileUploadConfig{
		MaxSize:     5 * 1024 * 1024, // 5MB
		UploadDir:   "./uploads",
		AllowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
	}

	var uploadedFiles []string

	// Handle poster image upload (optional)
	if _, _, err := ctx.Request.FormFile("poster_image"); err == nil {
		posterPath, err := utils.UploadImageFile(ctx, "poster_image", "posters", uploadConfig)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Error uploading poster image: " + err.Error(),
			})
			return
		}
		req.PosterImage = &posterPath
		uploadedFiles = append(uploadedFiles, posterPath)
	}

	// Handle background image upload (optional)
	if _, _, err := ctx.Request.FormFile("bg_path"); err == nil {
		bgPath, err := utils.UploadImageFile(ctx, "bg_path", "backgrounds", uploadConfig)
		if err != nil {
			// Clean up any previously uploaded files
			for _, file := range uploadedFiles {
				utils.DeleteFile(file)
			}
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Error uploading background image: " + err.Error(),
			})
			return
		}
		req.BgPath = &bgPath
		uploadedFiles = append(uploadedFiles, bgPath)
	}

	// Call repository
	movie, err := h.mar.UpdateMovie(ctx.Request.Context(), movieId, req)
	if err != nil {
		// Clean up uploaded files if database operation fails
		for _, file := range uploadedFiles {
			utils.DeleteFile(file)
		}

		switch err.Error() {
		case "movie not found or already deleted":
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Film tidak ditemukan atau sudah dihapus",
			})
		case "no fields to update":
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Tidak ada field yang diupdate",
			})
		default:
			log.Println("UpdateMovie error:", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal server error",
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Film berhasil diupdate",
		"data":    movie,
	})
}

// ======================= HELPER FUNCTIONS =======================

func (h *MovieAdminHandler) parseAddMovieForm(ctx *gin.Context) (*models.MovieCreateFormRequest, error) {
	title := ctx.PostForm("title")
	synopsis := ctx.PostForm("synopsis")
	durationStr := ctx.PostForm("duration_minutes")
	releaseDateStr := ctx.PostForm("release_date")
	directorsIdStr := ctx.PostForm("directors_id")
	ratingStr := ctx.PostForm("rating")

	// Validate required fields
	if title == "" || synopsis == "" || durationStr == "" || releaseDateStr == "" || directorsIdStr == "" || ratingStr == "" {
		return nil, fmt.Errorf("semua field wajib diisi")
	}

	// Convert string to appropriate types
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		return nil, fmt.Errorf("duration harus berupa angka")
	}

	directorsId, err := strconv.Atoi(directorsIdStr)
	if err != nil {
		return nil, fmt.Errorf("directors ID harus berupa angka")
	}

	rating, err := strconv.ParseFloat(ratingStr, 64)
	if err != nil {
		return nil, fmt.Errorf("rating harus berupa angka")
	}

	releaseDate, err := time.Parse("2006-01-02", releaseDateStr)
	if err != nil {
		return nil, fmt.Errorf("format tanggal harus YYYY-MM-DD")
	}

	return &models.MovieCreateFormRequest{
		Title:           title,
		Synopsis:        synopsis,
		DurationMinutes: duration,
		ReleaseDate:     releaseDate,
		DirectorsId:     directorsId,
		Rating:          rating,
	}, nil
}

func (h *MovieAdminHandler) parseUpdateMovieForm(ctx *gin.Context) (models.MovieUpdateRequest, error) {
	var req models.MovieUpdateRequest

	// Handle optional fields
	if title := ctx.PostForm("title"); title != "" {
		req.Title = &title
	}

	if synopsis := ctx.PostForm("synopsis"); synopsis != "" {
		req.Synopsis = &synopsis
	}

	if durationStr := ctx.PostForm("duration_minutes"); durationStr != "" {
		duration, err := strconv.Atoi(durationStr)
		if err != nil {
			return req, fmt.Errorf("duration harus berupa angka")
		}
		req.DurationMinutes = &duration
	}

	if releaseDateStr := ctx.PostForm("release_date"); releaseDateStr != "" {
		req.ReleaseDate = &releaseDateStr
	}

	if directorsIdStr := ctx.PostForm("directors_id"); directorsIdStr != "" {
		directorsId, err := strconv.Atoi(directorsIdStr)
		if err != nil {
			return req, fmt.Errorf("directors ID harus berupa angka")
		}
		req.DirectorsId = &directorsId
	}

	if ratingStr := ctx.PostForm("rating"); ratingStr != "" {
		rating, err := strconv.ParseFloat(ratingStr, 64)
		if err != nil {
			return req, fmt.Errorf("rating harus berupa angka")
		}
		req.Rating = &rating
	}

	return req, nil
}

// ======================= DELETE =======================

// DeleteMovie godoc
// @Summary     Soft delete movie (Admin)
// @Description Tandai movie sebagai deleted (soft delete) dan isi deleted_at timestamp
// @Tags        Admin-Movies
// @Security    BearerAuth
// @Produce     json
// @Param       movieId path int true "Movie ID"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /admin/movies/delete/{movieId} [delete]
func (h *MovieAdminHandler) DeleteMovie(ctx *gin.Context) {
	movieIdStr := ctx.Param("movieId")
	movieId, err := strconv.Atoi(movieIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid movie ID",
		})
		return
	}

	err = h.mar.DeleteMovie(ctx.Request.Context(), movieId)
	if err != nil {
		if err.Error() == "movie not found or already deleted" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Film tidak ditemukan atau sudah dihapus",
			})
			return
		}
		log.Println("DeleteMovie error:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Film berhasil dihapus",
	})
}
