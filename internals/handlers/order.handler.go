package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/raihaninkam/tickitz/internals/models"
	"github.com/raihaninkam/tickitz/internals/repositories"
	"github.com/raihaninkam/tickitz/pkg"
)

// avail seat

type SeatsHandler struct {
	sr *repositories.SeatsRepository
}

func NewSeatsHandler(sr *repositories.SeatsRepository) *SeatsHandler {
	return &SeatsHandler{sr: sr}
}

// GetAvailableSeats godoc
// @Summary      Mendapatkan semua kursi
// @Description  Mengambil daftar kursi berdasarkan `now_showing_id`.
// @Tags         Seats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        now_showing_id   path      int  true  "ID Now Showing"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /orders/seats/{now_showing_id} [get]
func (s *SeatsHandler) GetAvailableSeats(ctx *gin.Context) {
	nowShowingIDStr := ctx.Param("now_showing_id")
	if nowShowingIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "now_showing_id harus diisi"})
		return
	}

	nowShowingID, err := strconv.Atoi(nowShowingIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "now_showing_id harus berupa angka"})
		return
	}

	seats, err := s.sr.GetAvailableSeats(ctx.Request.Context(), nowShowingID)
	if err != nil {
		if len(seats) == 0 {
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Tidak ada kursi tersedia",
				"data":    []models.AvailSeat{},
			})
			return
		}

		log.Println("Internal Server Error.\nCause: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Berhasil mengambil data kursi",
		"data":    seats,
	})
}

// order

type OrderHandler struct {
	or *repositories.OrderRepository
	sr *repositories.SeatsRepository
}

func NewOrderHandler(orderRepo *repositories.OrderRepository, seatsRepo *repositories.SeatsRepository) *OrderHandler {
	return &OrderHandler{
		or: orderRepo,
		sr: seatsRepo,
	}
}

// CreateOrder membuat pesanan baru
// @Summary Create new order
// @Description Membuat order baru beserta ticket, update showing_seats, dan relasi ke cinema
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body models.CreateOrderRequest true "Order Request"
// @Success 201 {object} models.CreateOrderResponse "Order berhasil dibuat"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /orders [post]
func (o *OrderHandler) CreateOrder(ctx *gin.Context) {
	log.Printf("=== CREATE ORDER HANDLER START ===")

	// Ambil claims dari middleware
	claims, exists := ctx.Get("claims")
	if !exists {
		log.Printf("Claims not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized - claims not found",
		})
		return
	}

	user, ok := claims.(pkg.Claims)
	if !ok {
		log.Printf("Claims type assertion failed: %T", claims)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Internal Server Error (claims invalid)",
		})
		return
	}

	log.Printf("User from JWT: ID=%d", user.UserId)

	// Ambil body request
	var body models.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Printf("Bind error: %v", err)

		if strings.Contains(err.Error(), "required") {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Semua field wajib diisi",
			})
			return
		}

		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Format request tidak valid",
		})
		return
	}

	log.Printf("Request body before user injection: %+v", body)

	// Validasi tambahan
	if body.Price <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Harga harus lebih dari 0",
		})
		return
	}

	if body.PaymentID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Payment ID tidak valid",
		})
		return
	}

	if body.NowShowingID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Now Showing ID tidak valid",
		})
		return
	}

	if body.CinemaID < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Cinema ID tidak valid",
		})
		return
	}

	if len(body.SeatsMap) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Minimal pilih satu kursi",
		})
		return
	}

	// Inject user ID dari JWT ke order
	body.UsersID = user.UserId

	log.Printf("Request body after user injection: %+v", body)

	// Panggil repository CreateOrder
	order, err := o.or.CreateOrder(ctx.Request.Context(), body)
	if err != nil {
		log.Printf("Repository error: %v", err)

		// Error handling lebih spesifik
		switch {
		case strings.Contains(err.Error(), "user not found"):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "User tidak ditemukan",
			})
		case strings.Contains(err.Error(), "showing not found"):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Jadwal tayang tidak ditemukan",
			})
		case strings.Contains(err.Error(), "cinema not found"):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Cinema tidak ditemukan",
			})
		case strings.Contains(err.Error(), "payment method not found"):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Metode pembayaran tidak ditemukan",
			})
		case strings.Contains(err.Error(), "seat not available"):
			ctx.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Salah satu kursi sudah terjual",
			})
		case strings.Contains(err.Error(), "invalid seat selection"):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Pemilihan kursi tidak valid",
			})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Terjadi kesalahan internal server",
			})
		}
		return
	}

	log.Printf("Order created successfully: %+v", order)
	log.Printf("=== CREATE ORDER HANDLER SUCCESS ===")

	// Response sukses
	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Order berhasil dibuat",
		"data":    order,
	})
}

// order history

// order history
type OrderHistoryHandler struct {
	ohr *repositories.OrderHistory
}

func NewOrderHistoryHandler(ohr *repositories.OrderHistory) *OrderHistoryHandler {
	return &OrderHistoryHandler{ohr: ohr}
}

// GetOrderHistory godoc
// @Summary      Get user order history
// @Description  Ambil riwayat pesanan berdasarkan user yang sedang login (dari JWT)
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Success response with order history"
// @Failure      400  {object}  map[string]interface{}  "Invalid request"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized, token tidak valid"
// @Failure      404  {object}  map[string]interface{}  "Tidak ada riwayat pesanan"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /orders/history [get]
func (h *OrderHistoryHandler) GetOrderHistory(ctx *gin.Context) {
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

	orderHistories, err := h.ohr.GetOrderHistory(ctx.Request.Context(), userID)
	if err != nil {
		if err.Error() == "no order history found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tidak ada riwayat pesanan",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orderHistories,
	})
}
