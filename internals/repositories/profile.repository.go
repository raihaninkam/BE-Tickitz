package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raihaninkam/tickitz/internals/models"
	"github.com/raihaninkam/tickitz/pkg"
)

type ProfileRepository struct {
	db *pgxpool.Pool
	hc *pkg.HashConfig
}

func NewProfileRepository(db *pgxpool.Pool, hc *pkg.HashConfig) *ProfileRepository {
	return &ProfileRepository{db: db, hc: hc}
}

// Get profile untuk response
func (p *ProfileRepository) GetProfileResponse(ctx context.Context, userID int) (models.UserProfileResponse, error) {
	sql := `SELECT u.id, u.email, u.role, u.poin, p.first_name, p.last_name, p.phone_number, p.profile_picture
			FROM users u
			JOIN profile p ON u.id = p.id
			WHERE u.id = $1`

	var profile models.UserProfileResponse
	if err := p.db.QueryRow(ctx, sql, userID).Scan(
		&profile.ID,
		&profile.Email,
		&profile.Role,
		&profile.Poin,
		&profile.FirstName,
		&profile.LastName,
		&profile.PhoneNumber,
		&profile.ProfilePicture,
	); err != nil {
		if err == pgx.ErrNoRows {
			return models.UserProfileResponse{}, errors.New("user profile not found")
		}
		log.Println("GetProfileResponse error:", err)
		return models.UserProfileResponse{}, err
	}
	return profile, nil
}

// Update profile
func (p *ProfileRepository) UpdateProfile(ctx context.Context, userId int, req models.ProfileUpdateRequest) (*models.Profile, error) {
	sql := `
		UPDATE profile 
		SET 
			first_name = $1,
			last_name = $2,
			phone_number = $3,
			profile_picture = $4,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING id, first_name, last_name, phone_number, profile_picture, created_at, updated_at
	`

	var profile models.Profile
	err := p.db.QueryRow(ctx, sql,
		req.FirstName,
		req.LastName,
		req.PhoneNumber,
		req.ProfilePicture,
		userId,
	).Scan(
		&profile.Id,
		&profile.FirstName,
		&profile.LastName,
		&profile.PhoneNumber,
		&profile.ProfilePicture,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("profile not found")
		}
		return nil, err
	}

	return &profile, nil
}

// Change password
func (p *ProfileRepository) ChangePassword(ctx context.Context, userId int, oldPass, newPass string) error {
	var hashedPassword string
	err := p.db.QueryRow(ctx, "SELECT password FROM users WHERE id=$1", userId).Scan(&hashedPassword)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errors.New("user not found")
		}
		return err
	}

	// cek password lama
	ok, err := p.hc.CompareHashAndPassword(oldPass, hashedPassword)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("old password incorrect")
	}

	// hash password baru dengan argon2
	newHashed, err := p.hc.GenHash(newPass)
	if err != nil {
		return err
	}

	_, err = p.db.Exec(ctx, "UPDATE users SET password=$1 WHERE id=$2", newHashed, userId)
	if err != nil {
		return err
	}

	return nil
}

// order history

type OrderHistory struct {
	Db *pgxpool.Pool
}

func NewOrderHistory(db *pgxpool.Pool) *OrderHistory {
	return &OrderHistory{Db: db}
}

func (o *OrderHistory) GetOrderHistory(ctx context.Context, userId int) ([]models.OrderHistory, error) {
	sql := `
		SELECT 
			o.id,
			o.users_id,
			o.price,
			o.payment_id,
			o."isPaid",
			o.created_at,
			o.now_showing_id,
			m.title AS movie_title,
			ns.date AS show_date,
			ns.time AS show_time,
			c.cinema_name,
			json_agg(json_build_object(
				'row', s.row,
				'seat_number', s.seat_number
			)) AS seats,
			t.qr_code
		FROM orders o
		JOIN now_showing ns ON o.now_showing_id = ns.id
		JOIN movies m ON ns.movie_id = m.id
		JOIN cinemas c ON ns.cinemas_id = c.id
		JOIN showing_seats ss ON o.now_showing_id = ss.now_showing_id AND ss.user_id = o.users_id
		JOIN seats s ON ss.seat_id = s.id
		JOIN orders_ticket ot ON o.id = ot.orders_id
		JOIN ticket t ON ot.ticket_id = t.id
		WHERE o.users_id = $1
		GROUP BY 
			o.id, o.users_id, o.price, o.payment_id, o."isPaid",
			o.created_at, o.now_showing_id,
			m.title, ns.date, ns.time, c.cinema_name, t.qr_code
		ORDER BY o.created_at DESC;
	`

	rows, err := o.Db.Query(ctx, sql, userId)
	if err != nil {
		log.Printf("query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var orderHistories []models.OrderHistory
	for rows.Next() {
		var oh models.OrderHistory
		var seatsJSON []byte

		if err := rows.Scan(
			&oh.Id,
			&oh.UsersId,
			&oh.Price,
			&oh.PaymentId,
			&oh.IsPaid,
			&oh.CreatedAt,
			&oh.NowShowingId,
			&oh.MovieTitle,
			&oh.ShowDate,
			&oh.ShowTime,
			&oh.CinemaName,
			&seatsJSON,
			&oh.QrCode,
		); err != nil {
			log.Printf("scan error: %v", err)
			return nil, err
		}

		// unmarshal kursi
		if err := json.Unmarshal(seatsJSON, &oh.Seats); err != nil {
			log.Printf("unmarshal seats error: %v | data: %s", err, string(seatsJSON))
			return nil, err
		}

		orderHistories = append(orderHistories, oh)
	}

	if len(orderHistories) == 0 {
		return nil, errors.New("no order history found")
	}

	return orderHistories, nil
}

// Tambahkan method GetUserByEmail
func (pr *ProfileRepository) GetUserByEmail(ctx context.Context, email string) (*models.Users, error) {
	query := `
		SELECT id, email, password, role
		FROM users 
		WHERE email = $1
	`

	var user models.Users
	err := pr.db.QueryRow(ctx, query, email).Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.Role,
	)

	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	return &user, nil
}
