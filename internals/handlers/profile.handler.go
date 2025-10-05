package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/raihaninkam/tickitz/internals/models"
	"github.com/raihaninkam/tickitz/internals/repositories"
	"github.com/raihaninkam/tickitz/pkg"
)

type ProfileHandler struct {
	pr *repositories.ProfileRepository
}

func NewProfileHandler(pr *repositories.ProfileRepository) *ProfileHandler {
	return &ProfileHandler{pr: pr}
}

// GetMyProfile godoc
// @Summary     Get My Profile
// @Description Ambil data profil user dari JWT
// @Tags        Profile
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} models.UserProfileResponse
// @Router      /profile [get]
func (p *ProfileHandler) GetMyProfile(ctx *gin.Context) {
	claimsValue, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Token tidak valid"})
		return
	}

	claims, ok := claimsValue.(pkg.Claims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Token tidak valid"})
		return
	}

	userID := claims.UserId

	profile, err := p.pr.GetProfileResponse(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "data": profile})
}

// UpdateProfileWithImage godoc
// @Summary     Update User Profile
// @Description Update data profil user (dengan upload gambar opsional)
// @Tags        Profile
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       first_name formData string true "First Name"
// @Param       last_name formData string true "Last Name"
// @Param       phone_number formData string true "Phone Number"
// @Param       image formData file false "Profile Picture"
// @Success     200 {object} models.Profile
// @Router      /profile [patch]
func (h *ProfileHandler) UpdateProfileWithImage(ctx *gin.Context) {
	claimsValue, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Token tidak valid"})
		return
	}

	claims, ok := claimsValue.(pkg.Claims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Token tidak valid"})
		return
	}

	userID := claims.UserId

	var body models.StudentBody
	if err := ctx.ShouldBindWith(&body, binding.FormMultipart); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid form data"})
		return
	}

	profileReq := models.ProfileUpdateRequest{
		FirstName:      body.FirstName,
		LastName:       body.LastName,
		PhoneNumber:    body.PhoneNumber,
		ProfilePicture: "",
	}

	// UBAH INI: Simpan ke public/images/profiles
	if body.Images != nil {
		ext := filepath.Ext(body.Images.Filename)
		filename := fmt.Sprintf("%d_profile_%d%s", time.Now().UnixNano(), userID, ext)

		// Ubah path ke public/images/profiles
		uploadDir := filepath.Join("public", "images", "profiles")
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Println("Failed to create upload dir:", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create directory"})
			return
		}

		location := filepath.Join(uploadDir, filename)
		if err := ctx.SaveUploadedFile(body.Images, location); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Upload gagal"})
			return
		}

		// Simpan relative path
		profileReq.ProfilePicture = filepath.Join("images", "profiles", filename)
	}

	profile, err := h.pr.UpdateProfile(ctx.Request.Context(), userID, profileReq)
	if err != nil {
		log.Println("UpdateProfile error:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Profile berhasil diupdate", "data": profile})
}

// ChangePassword godoc
// @Summary     Change Password
// @Description Ubah password user
// @Tags        Profile
// @Accept      json
// @Produce     json
// @Param       body body models.ChangePasswordRequest true "Password Data"
// @Success     200 {string} string "Password berhasil diubah"
// @Router      /profile/change-password [patch]
func (h *ProfileHandler) ChangePassword(ctx *gin.Context) {
	var req models.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body: " + err.Error()})
		return
	}

	// Validasi: email harus disertakan
	if req.Email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Email harus disertakan"})
		return
	}

	// Cari user berdasarkan email
	user, err := h.pr.GetUserByEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "User tidak ditemukan"})
		return
	}

	// Ubah password
	if err := h.pr.ChangePassword(ctx.Request.Context(), user.Id, req.OldPassword, req.NewPassword); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Password berhasil diubah"})
}
