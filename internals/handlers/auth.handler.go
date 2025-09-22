package handlers

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/raihaninkam/tickitz/internals/models"
	"github.com/raihaninkam/tickitz/internals/repositories"
	"github.com/raihaninkam/tickitz/internals/utils"
	"github.com/raihaninkam/tickitz/pkg"
)

type AuthHandler struct {
	ar *repositories.AuthRepository
}

func NewAuthHandler(ar *repositories.AuthRepository) *AuthHandler {
	return &AuthHandler{ar: ar}
}

// Register godoc
// @Summary     Register User
// @Description Daftar User baru dengan email dan password. Password akan di-hash sebelum disimpan.
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body models.UserAuth true "Register Request"
// @Success     201 {object} map[string]interface{} "User berhasil didaftarkan"
// @Failure     400 {object} map[string]interface{} "Bad Request - Input tidak valid (email, password)"
// @Failure     409 {object} map[string]interface{} "Conflict - Email sudah terdaftar"
// @Failure     500 {object} map[string]interface{} "Internal Server Error"
// @Router      /auth/register [post]
func (a *AuthHandler) Register(ctx *gin.Context) {
	// menerima body
	var body models.UserAuth
	if err := ctx.ShouldBind(&body); err != nil {
		if strings.Contains(err.Error(), "required") {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Email dan Password harus diisi",
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

	// validasi email
	if err := utils.ValidateEmail(body.Email); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// validasi password
	if err := utils.ValidatePassword(body.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// cek apakah email sudah terdaftar
	exists, err := a.ar.CheckEmailExists(ctx.Request.Context(), body.Email)
	if err != nil {
		log.Println("Internal Server Error.\nCause: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}
	if exists {
		ctx.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Email sudah terdaftar",
		})
		return
	}

	// hash password sebelum disimpan
	hc := pkg.NewHashConfig()
	hc.UseRecommended() // menggunakan konfigurasi yang direkomendasikan
	hashedPassword, err := hc.GenHash(body.Password)
	if err != nil {
		log.Println("Internal Server Error.\nCause: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	// simpan user ke database
	err = a.ar.RegisterUserWithProfile(ctx.Request.Context(), body.Email, hashedPassword)
	if err != nil {
		if strings.Contains(err.Error(), "email already exists") {
			ctx.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Email sudah terdaftar",
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

	// response sukses tanpa data user
	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User berhasil didaftarkan",
	})
}

// Login godoc
// @Summary     Login User
// @Description Login dengan email dan password. Jika sukses, akan mengembalikan JWT Token untuk autentikasi.
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body models.UserAuth true "Login Request"
// @Success     200 {object} map[string]interface{} "Berhasil login, kembalikan token"
// @Failure     400 {object} map[string]interface{} "Bad Request - Email/Password salah atau input tidak valid"
// @Failure     500 {object} map[string]interface{} "Internal Server Error"
// @Router      /auth/login [post]
func (a *AuthHandler) Login(ctx *gin.Context) {
	// menerima body
	var body models.UserAuth
	if err := ctx.ShouldBind(&body); err != nil {
		if strings.Contains(err.Error(), "required") {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Email dan Password harus diisi",
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

	// validasi basic email format (opsional untuk login)
	if body.Email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Email harus diisi",
		})
		return
	}

	if body.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Password harus diisi",
		})
		return
	}

	// ambil data user dari database
	user, err := a.ar.GetEmailUserWithPasswordAndRole(ctx.Request.Context(), body.Email)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Email atau password salah",
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

	// bandingkan password menggunakan hash function Anda
	hc := pkg.NewHashConfig()
	isMatched, err := hc.CompareHashAndPassword(body.Password, user.Password)
	if err != nil {
		log.Println("Internal Server Error.\nCause: ", err.Error())
		// Cek jika error terkait dengan hash format atau crypto
		if strings.Contains(err.Error(), "hash") ||
			strings.Contains(err.Error(), "crypto") ||
			strings.Contains(err.Error(), "argon2id") ||
			strings.Contains(err.Error(), "format") {
			log.Println("Error during password hashing/comparison")
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	// jika password tidak cocok
	if !isMatched {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Email atau password salah",
		})
		return
	}

	// jika match, buat JWT token
	claims := pkg.NewJWTClaims(user.Id, user.Role)
	jwtToken, err := claims.GenToken()
	if err != nil {
		log.Println("Internal Server Error.\nCause: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	// response sukses dengan token
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login berhasil",
		"token":   jwtToken,
		"role":    claims.Role,
	})
}

// SecureLogout godoc
// @Summary      Secure Logout User
// @Description  Logout user dengan menambahkan token ke blacklist. Token tidak dapat digunakan lagi.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.LogoutSuccessResponse  "Logout berhasil dan token di-blacklist"
// @Failure      400  {object}  models.ErrorResponse          "Bad Request - Format token tidak valid"
// @Failure      401  {object}  models.ErrorResponse          "Unauthorized - Token tidak ditemukan"
// @Failure      500  {object}  models.ErrorResponse          "Internal Server Error"
// @Router       /auth/logout [post]
func (a *AuthHandler) SecureLogout(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Token tidak ditemukan",
		})
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Format token tidak valid",
		})
		return
	}

	tokenString := tokenParts[1]

	claims := &pkg.Claims{}
	err := claims.VerifyToken(tokenString)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Token tidak valid",
		})
		return
	}

	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !parsedToken.Valid {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Token tidak valid",
		})
		return
	}

	expiryTime := claims.ExpiresAt.Time
	err = a.ar.AddToBlacklist(ctx.Request.Context(), tokenString, expiryTime)
	if err != nil {
		log.Println("Internal Server Error.\nCause: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout berhasil. Token telah diblacklist.",
		"user": gin.H{
			"id":   claims.UserId,
			"role": claims.Role,
		},
	})
}
