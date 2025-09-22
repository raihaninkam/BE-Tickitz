package models

import "time"

type MovieAdmin struct {
	Id              int        `json:"id"`
	Title           string     `json:"title"`
	Synopsis        string     `json:"synopsis"`
	DurationMinutes int        `json:"duration_minutes"`
	ReleaseDate     time.Time  `json:"release_date"`
	PosterImage     *string    `json:"poster_image"`
	DirectorsId     int        `json:"directors_id"`
	Rating          *float64   `json:"rating"`
	BgPath          *string    `json:"bg_path"`
	DirectorName    string     `json:"director_name"`
	Genres          *string    `json:"genres"`
	IsDeleted       bool       `json:"is_deleted"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

type MovieEdit struct {
	Id              int        `json:"id"`
	Title           string     `json:"title"`
	Synopsis        string     `json:"synopsis"`
	DurationMinutes int        `json:"duration_minutes"`
	ReleaseDate     time.Time  `json:"release_date"`
	PosterImage     string     `json:"poster_image"`
	DirectorsId     int        `json:"directors_id"`
	Rating          float64    `json:"rating"`
	BgPath          string     `json:"bg_path"`
	IsDeleted       bool       `json:"is_deleted"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

type MovieUpdateRequest struct {
	Title           *string  `json:"title"`
	Synopsis        *string  `json:"synopsis"`
	DurationMinutes *int     `json:"duration_minutes"`
	ReleaseDate     *string  `json:"release_date"`
	PosterImage     *string  `json:"poster_image"`
	DirectorsId     *int     `json:"directors_id"`
	Rating          *float64 `json:"rating"`
	BgPath          *string  `json:"bg_path"`
}

type MovieCreateRequest struct {
	Title           string  `json:"title" binding:"required"`
	Synopsis        string  `json:"synopsis" binding:"required"`
	DurationMinutes int     `json:"duration_minutes" binding:"required"`
	ReleaseDate     string  `json:"release_date" binding:"required"`
	PosterImage     string  `json:"poster_image" binding:"required"`
	DirectorsId     int     `json:"directors_id" binding:"required"`
	Rating          float64 `json:"rating" binding:"required"`
	BgPath          string  `json:"bg_path" binding:"required"`
}

type MovieCreateFormRequest struct {
	Title           string    `form:"title" binding:"required"`
	Synopsis        string    `form:"synopsis" binding:"required"`
	DurationMinutes int       `form:"duration_minutes" binding:"required"`
	ReleaseDate     time.Time `form:"-"`
	DirectorsId     int       `form:"directors_id" binding:"required"`
	Rating          float64   `form:"rating" binding:"required"`
}

// For update form requests
type MovieUpdateFormRequest struct {
	Title           *string  `form:"title"`
	Synopsis        *string  `form:"synopsis"`
	DurationMinutes *int     `form:"duration_minutes"`
	ReleaseDate     *string  `form:"release_date"` // String format, will be validated
	DirectorsId     *int     `form:"directors_id"`
	Rating          *float64 `form:"rating"`
	// Files handled separately
}
