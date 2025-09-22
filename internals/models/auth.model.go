package models

import "time"

type Users struct {
	Id       int    `db:"id" json:"id"`
	Email    string `db:"email" json:"email"`
	Role     string `db:"role" json:"role,omitempty"`
	Password string `db:"password" json:"password,omitempty"`
}

type UserAuth struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=4"`
}

type BlacklistToken struct {
	Token     string    `db:"token" json:"token"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type LogoutUserInfo struct {
	ID   int    `json:"id" example:"1"`
	Role string `json:"role" example:"user"`
}

type LogoutSuccessResponse struct {
	Success bool           `json:"success" example:"true"`
	Message string         `json:"message" example:"Logout berhasil. Token telah diblacklist."`
	User    LogoutUserInfo `json:"user,omitempty"`
}
