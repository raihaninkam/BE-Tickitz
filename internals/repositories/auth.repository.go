package repositories

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raihaninkam/tickitz/internals/models"
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{db: db}
}

func (a *AuthRepository) GetEmailUserWithPasswordAndRole(rctx context.Context, email string) (models.Users, error) {
	sql := "SELECT id, email, password, role FROM users WHERE email = $1"

	var users models.Users
	if err := a.db.QueryRow(rctx, sql, email).Scan(&users.Id, &users.Email, &users.Password, &users.Role); err != nil {
		if err == pgx.ErrNoRows {
			return models.Users{}, errors.New("user not found")
		}
		log.Println("Internal Server Error.\nCz: ", err.Error())
		return models.Users{}, err
	}

	return users, nil
}

func (a *AuthRepository) CheckEmailExists(rctx context.Context, email string) (bool, error) {
	sql := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"

	var exists bool
	if err := a.db.QueryRow(rctx, sql, email).Scan(&exists); err != nil {
		log.Println("Error checking email existence:", err.Error())
		return false, err
	}

	return exists, nil
}

func (a *AuthRepository) RegisterUserWithProfile(rctx context.Context, email, password string) error {
	defaultRole := "user"

	tx, err := a.db.Begin(rctx)
	if err != nil {
		log.Println("Failed to start transaction:", err.Error())
		return err
	}
	defer tx.Rollback(rctx)

	// Insert ke tabel users
	var userId int
	sqlUser := "INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id"
	if err := tx.QueryRow(rctx, sqlUser, email, password, defaultRole).Scan(&userId); err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" ||
			err.Error() == "UNIQUE constraint failed: users.email" {
			return errors.New("email already exists")
		}
		log.Println("Error inserting user:", err.Error())
		return err
	}

	// Insert ke tabel profile
	sqlProfile := `
        INSERT INTO profile (id, first_name, last_name, phone_number, profile_picture, created_at, updated_at)
        VALUES ($1, '', '', '', '', NOW(), NOW())
    `
	_, err = tx.Exec(rctx, sqlProfile, userId)
	if err != nil {
		log.Println("Error inserting profile:", err.Error())
		return err
	}

	// Commit jika semua sukses
	if err := tx.Commit(rctx); err != nil {
		log.Println("Transaction commit failed:", err.Error())
		return err
	}

	return nil
}

// blacklist token

// Method untuk menambahkan token ke blacklist
func (a *AuthRepository) AddToBlacklist(rctx context.Context, token string, expiresAt time.Time) error {
	sql := "INSERT INTO blacklist_tokens (token, expires_at, created_at) VALUES ($1, $2, NOW())"

	_, err := a.db.Exec(rctx, sql, token, expiresAt)
	if err != nil {
		log.Println("Error adding token to blacklist:", err.Error())
		return err
	}
	return nil
}

// Method untuk mengecek apakah token sudah di-blacklist
func (a *AuthRepository) IsTokenBlacklisted(rctx context.Context, token string) (bool, error) {
	sql := "SELECT EXISTS(SELECT 1 FROM blacklist_tokens WHERE token = $1 AND expires_at > NOW())"

	var exists bool
	if err := a.db.QueryRow(rctx, sql, token).Scan(&exists); err != nil {
		log.Println("Error checking token blacklist:", err.Error())
		return false, err
	}

	return exists, nil
}

func (a *AuthRepository) CleanupExpiredTokens(rctx context.Context) error {
	sql := "DELETE FROM blacklist_tokens WHERE expires_at < NOW()"

	_, err := a.db.Exec(rctx, sql)
	if err != nil {
		log.Println("error cleaning up expired tokens:", err.Error())
		return err
	}
	return nil
}
