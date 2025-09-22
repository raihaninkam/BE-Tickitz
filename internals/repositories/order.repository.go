package repositories

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raihaninkam/tickitz/internals/models"
)

// availseat

type SeatsRepository struct {
	db *pgxpool.Pool
}

func NewSeatsRepository(db *pgxpool.Pool) *SeatsRepository {
	return &SeatsRepository{db: db}
}

// GetAvailableSeats mengambil semua kursi yang statusnya 'available' untuk now_showing_id tertentu
func (s *SeatsRepository) GetAvailableSeats(rctx context.Context, nowShowingID int) ([]models.AvailSeat, error) {
	log.Printf("=== DYNAMIC SEATS GENERATION ===")
	log.Printf("Generating seats for now_showing_id: %d", nowShowingID)

	// Step 1: Ambil cinema_id dari now_showing
	var cinemaID int
	getCinemaSQL := "SELECT cinemas_id FROM now_showing WHERE id = $1"
	err := s.db.QueryRow(rctx, getCinemaSQL, nowShowingID).Scan(&cinemaID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("now_showing_id %d not found", nowShowingID)
			return nil, errors.New("showing not found")
		}
		log.Printf("Error getting cinema_id: %v", err)
		return nil, err
	}

	log.Printf("Cinema ID for now_showing %d: %d", nowShowingID, cinemaID)

	// Step 2: Generate semua seats virtual dari master seats table
	// dan check status dari showing_seats jika ada booking
	sql := `
	SELECT 
		CONCAT(s.row, s.seat_number) as seat_id,
		$1 as showing_id,
		COALESCE(ss.status = 'sold', false) as is_sold,
		CASE 
			WHEN s.row = 'F' AND s.seat_number BETWEEN 7 AND 10 THEN true 
			ELSE false 
		END as is_love_nest
	FROM seats s
	LEFT JOIN showing_seats ss ON (ss.seat_id = s.id AND ss.now_showing_id = $1)
	WHERE s.cinemas_id = $2
	ORDER BY s.row, s.seat_number;
	`

	log.Printf("Executing SQL: %s", sql)
	log.Printf("Parameters: nowShowingID=%d, cinemaID=%d", nowShowingID, cinemaID)

	rows, err := s.db.Query(rctx, sql, nowShowingID, cinemaID)
	if err != nil {
		log.Printf("SQL Query Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var availableSeats []models.AvailSeat
	for rows.Next() {
		var seat models.AvailSeat
		if err := rows.Scan(&seat.SeatID, &seat.ShowingId, &seat.IsSold, &seat.IsLoveNest); err != nil {
			log.Printf("Row Scan Error: %v", err)
			return nil, err
		}
		availableSeats = append(availableSeats, seat)
	}

	log.Printf("Total seats generated: %d", len(availableSeats))

	if len(availableSeats) == 0 {
		log.Printf("No seats found for cinema_id: %d", cinemaID)
		return nil, errors.New("no seats found for this cinema")
	}

	// Debug summary
	soldCount := 0
	loveNestCount := 0
	for _, seat := range availableSeats {
		if seat.IsSold {
			soldCount++
		}
		if seat.IsLoveNest {
			loveNestCount++
		}
	}

	log.Printf("Seats summary - Total: %d, Sold: %d, Love Nest: %d, Available: %d",
		len(availableSeats), soldCount, loveNestCount, len(availableSeats)-soldCount)

	log.Printf("Sample seat: %+v", availableSeats[0])
	log.Printf("=== END DYNAMIC SEATS GENERATION ===")

	return availableSeats, nil
}

///////////////////////////////////////////////////////////////////////////

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

// generateQRCode generates a random QR code string
func (o *OrderRepository) generateQRCode() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateOrder creates a new order with all related records in a transaction
func (o *OrderRepository) CreateOrder(rctx context.Context, req models.CreateOrderRequest) (models.CreateOrderResponse, error) {
	log.Printf("=== CREATE ORDER START ===")
	log.Printf("Request: %+v", req)

	// Start transaction
	tx, err := o.db.Begin(rctx)
	if err != nil {
		log.Printf("Transaction begin error: %v", err)
		return models.CreateOrderResponse{}, err
	}
	defer tx.Rollback(rctx)

	var response models.CreateOrderResponse
	var orderID, ticketID int
	qrCode := o.generateQRCode()

	// Validate user exists
	var userExists bool
	userCheckSQL := "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"
	if err := tx.QueryRow(rctx, userCheckSQL, req.UsersID).Scan(&userExists); err != nil {
		log.Printf("User check error: %v", err)
		return models.CreateOrderResponse{}, err
	}
	if !userExists {
		log.Printf("User not found: %d", req.UsersID)
		return models.CreateOrderResponse{}, errors.New("user not found")
	}

	// Validate showing exists and get cinema_id
	var showingExists bool
	var actualCinemaID int
	showingCheckSQL := "SELECT EXISTS(SELECT 1 FROM now_showing WHERE id = $1), cinemas_id FROM now_showing WHERE id = $1"
	if err := tx.QueryRow(rctx, showingCheckSQL, req.NowShowingID).Scan(&showingExists, &actualCinemaID); err != nil {
		if err == pgx.ErrNoRows {
			return models.CreateOrderResponse{}, errors.New("showing not found")
		}
		log.Printf("Showing check error: %v", err)
		return models.CreateOrderResponse{}, err
	}
	if !showingExists {
		return models.CreateOrderResponse{}, errors.New("showing not found")
	}

	// Validate cinema matches
	if actualCinemaID != req.CinemaID {
		log.Printf("Cinema mismatch: request=%d, actual=%d", req.CinemaID, actualCinemaID)
		return models.CreateOrderResponse{}, errors.New("cinema not found")
	}

	// Validate payment method exists
	var paymentExists bool
	paymentCheckSQL := "SELECT EXISTS(SELECT 1 FROM payment WHERE id = $1)"
	if err := tx.QueryRow(rctx, paymentCheckSQL, req.PaymentID).Scan(&paymentExists); err != nil {
		log.Printf("Payment check error: %v", err)
		return models.CreateOrderResponse{}, err
	}
	if !paymentExists {
		return models.CreateOrderResponse{}, errors.New("payment method not found")
	}

	log.Printf("Validations passed. Creating order...")

	// 1. Insert into ORDERS
	orderSQL := `INSERT INTO orders (users_id, price, payment_id, now_showing_id, cinemas_id, created_at, updated_at)
				 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
				 RETURNING id`

	if err := tx.QueryRow(rctx, orderSQL, req.UsersID, req.Price, req.PaymentID, req.NowShowingID, req.CinemaID).Scan(&orderID); err != nil {
		log.Printf("Order insert error: %v", err)
		return models.CreateOrderResponse{}, err
	}
	log.Printf("Order created with ID: %d", orderID)

	// 2. Create ticket
	ticketSQL := `INSERT INTO ticket (qr_code, created_at, updated_at)
				  VALUES ($1, NOW(), NOW())
				  RETURNING id`

	if err := tx.QueryRow(rctx, ticketSQL, qrCode).Scan(&ticketID); err != nil {
		log.Printf("Ticket insert error: %v", err)
		return models.CreateOrderResponse{}, err
	}
	log.Printf("Ticket created with ID: %d, QR: %s", ticketID, qrCode)

	// 3. Link ticket to order
	ordersTicketSQL := `INSERT INTO orders_ticket (orders_id, ticket_id, created_at, updated_at)
						VALUES ($1, $2, NOW(), NOW())`

	if _, err := tx.Exec(rctx, ordersTicketSQL, orderID, ticketID); err != nil {
		log.Printf("Orders-ticket link error: %v", err)
		return models.CreateOrderResponse{}, err
	}
	log.Printf("Order-ticket relationship created")

	// 4. Handle showing_seats - INSERT or UPDATE based on existence
	log.Printf("Processing %d seats: %v", len(req.SeatsMap), req.SeatsMap)

	for i, seatIdentifier := range req.SeatsMap {
		log.Printf("Processing seat %d/%d: %s", i+1, len(req.SeatsMap), seatIdentifier)

		// Get the actual seat ID from seats table based on seat identifier (like "A1")
		var actualSeatID int
		getSeatIDSQL := `SELECT id FROM seats WHERE CONCAT(row, seat_number) = $1 AND cinemas_id = $2`

		if err := tx.QueryRow(rctx, getSeatIDSQL, seatIdentifier, req.CinemaID).Scan(&actualSeatID); err != nil {
			if err == pgx.ErrNoRows {
				log.Printf("Seat %s not found in cinema %d", seatIdentifier, req.CinemaID)
				return models.CreateOrderResponse{}, errors.New("invalid seat selection")
			}
			log.Printf("Error getting seat ID for %s: %v", seatIdentifier, err)
			return models.CreateOrderResponse{}, err
		}
		log.Printf("Seat %s mapped to ID: %d", seatIdentifier, actualSeatID)

		// Check if record already exists in showing_seats
		var existingRecordCount int
		checkSQL := `SELECT COUNT(*) FROM showing_seats WHERE now_showing_id = $1 AND seat_id = $2`

		if err := tx.QueryRow(rctx, checkSQL, req.NowShowingID, actualSeatID).Scan(&existingRecordCount); err != nil {
			log.Printf("Error checking existing showing_seats for seat %s: %v", seatIdentifier, err)
			return models.CreateOrderResponse{}, err
		}

		if existingRecordCount > 0 {
			log.Printf("Seat %s already has record in showing_seats, checking status", seatIdentifier)

			// Record exists, check if already sold
			var currentStatus string
			var currentUserID *int
			statusCheckSQL := `SELECT status, user_id FROM showing_seats WHERE now_showing_id = $1 AND seat_id = $2`

			if err := tx.QueryRow(rctx, statusCheckSQL, req.NowShowingID, actualSeatID).Scan(&currentStatus, &currentUserID); err != nil {
				log.Printf("Error checking seat status for %s: %v", seatIdentifier, err)
				return models.CreateOrderResponse{}, err
			}

			if currentStatus == "sold" {
				log.Printf("Seat %s is already sold to user %v", seatIdentifier, currentUserID)
				return models.CreateOrderResponse{}, errors.New("seat not available")
			}

			// Update existing record
			updateSQL := `UPDATE showing_seats 
						  SET status = 'sold', user_id = $1, updated_at = NOW()
						  WHERE now_showing_id = $2 AND seat_id = $3`

			if _, err := tx.Exec(rctx, updateSQL, req.UsersID, req.NowShowingID, actualSeatID); err != nil {
				log.Printf("Error updating showing_seats for seat %s: %v", seatIdentifier, err)
				return models.CreateOrderResponse{}, err
			}
			log.Printf("Updated seat %s status to sold", seatIdentifier)
		} else {
			log.Printf("Seat %s has no record in showing_seats, inserting new", seatIdentifier)

			// Record doesn't exist, insert new record
			insertSQL := `INSERT INTO showing_seats (now_showing_id, seat_id, status, user_id, created_at, updated_at)
						  VALUES ($1, $2, 'sold', $3, NOW(), NOW())`

			if _, err := tx.Exec(rctx, insertSQL, req.NowShowingID, actualSeatID, req.UsersID); err != nil {
				log.Printf("Error inserting into showing_seats for seat %s: %v", seatIdentifier, err)
				return models.CreateOrderResponse{}, err
			}
			log.Printf("Inserted new record for seat %s", seatIdentifier)
		}

		log.Printf("Successfully processed seat %s (ID: %d) for user %d", seatIdentifier, actualSeatID, req.UsersID)
	}

	// Commit transaction
	if err := tx.Commit(rctx); err != nil {
		log.Printf("Transaction commit error: %v", err)
		return models.CreateOrderResponse{}, err
	}
	log.Printf("Transaction committed successfully")

	// Prepare response
	response = models.CreateOrderResponse{
		ID:        orderID,
		UsersID:   req.UsersID,
		Price:     req.Price,
		QRCode:    qrCode,
		TicketID:  ticketID,
		SeatsMap:  req.SeatsMap,
		CreatedAt: time.Now(),
	}

	log.Printf("=== CREATE ORDER SUCCESS ===")
	log.Printf("Response: %+v", response)

	return response, nil
}
