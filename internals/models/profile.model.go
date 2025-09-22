package models

import (
	"mime/multipart"
	"time"
)

// BaseResponse untuk membungkus semua response API
type BaseResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Request successful"`
	Data    interface{} `json:"data,omitempty"`
}

// Profile di DB
type Profile struct {
	Id             int       `json:"id" example:"1"`
	FirstName      string    `json:"first_name" example:"John"`
	LastName       string    `json:"last_name" example:"Doe"`
	PhoneNumber    string    `json:"phone_number" example:"+628123456789"`
	ProfilePicture string    `json:"profile_picture" example:"profile_123.png"`
	CreatedAt      time.Time `json:"created_at" example:"2025-09-14T12:00:00Z"`
	UpdatedAt      time.Time `json:"updated_at" example:"2025-09-14T12:30:00Z"`
}

// Request untuk update profile (JSON atau form tanpa file)
type ProfileUpdateRequest struct {
	FirstName      string `json:"first_name" form:"first_name" example:"John"`
	LastName       string `json:"last_name" form:"last_name" example:"Doe"`
	PhoneNumber    string `json:"phone_number" form:"phone_number" example:"+628123456789"`
	ProfilePicture string `json:"profile_picture,omitempty" example:"profile_123.png"`
}

// Request body untuk upload form-data
type StudentBody struct {
	FirstName   string                `form:"first_name" example:"John"`
	LastName    string                `form:"last_name" example:"Doe"`
	PhoneNumber string                `form:"phone_number" example:"+628123456789"`
	Images      *multipart.FileHeader `form:"image"`
}

// Untuk response lengkap (tanpa password)
type UserProfileResponse struct {
	ID             int    `json:"id" example:"1"`
	Email          string `json:"email" example:"john@example.com"`
	Role           string `json:"role" example:"user"`
	Poin           int    `json:"poin" example:"120"`
	FirstName      string `json:"first_name" example:"John"`
	LastName       string `json:"last_name" example:"Doe"`
	PhoneNumber    string `json:"phone_number" example:"+628123456789"`
	ProfilePicture string `json:"profile_picture" example:"profile_123.png"`
}

type ChangePasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
	// Hapus ConfirmPassword jika tidak digunakan di backend
}
