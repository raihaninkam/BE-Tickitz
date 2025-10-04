package models

import "time"

type MovieAdmin struct {
	Id              int        `json:"id"`
	Title           string     `json:"title" binding:"required"`
	Synopsis        string     `json:"synopsis"`
	DurationMinutes int        `json:"duration_minutes"`
	ReleaseDate     time.Time  `json:"release_date"`
	PosterImage     *string    `json:"poster_image"`
	DirectorsId     int        `json:"directors_id"`
	Rating          *float64   `json:"rating"`
	BgPath          *string    `json:"bg_path"`
	GenresId        []int      `json:"genres_id"`
	Genres          string     `json:"genres"` // Tambahkan ini untuk nama genre
	CastsId         []int      `json:"casts_id"`
	Showtimes       []Showtime `json:"showtimes"`
}

// Model untuk showtime (optional)
type Showtime struct {
	Date       string `json:"date"` // Ubah dari time.Time ke string
	Time       string `json:"time"` // Ubah dari time.Time ke string
	LocationId int    `json:"location_id"`
	CinemasId  int    `json:"cinemas_id"`
}

// Model untuk response
type MovieEdit struct {
	Id              int       `json:"id"`
	Title           string    `json:"title"`
	Synopsis        string    `json:"synopsis"`
	DurationMinutes int       `json:"duration_minutes"`
	ReleaseDate     time.Time `json:"release_date"`
	PosterImage     string    `json:"poster_image"`
	DirectorsId     int       `json:"directors_id"`
	Rating          float64   `json:"rating"`
	BgPath          *string   `json:"bg_path"`
}

type MovieUpdateComprehensiveRequest struct {
	Title           *string    `json:"title,omitempty"`
	Synopsis        *string    `json:"synopsis,omitempty"`
	DurationMinutes *int       `json:"duration_minutes,omitempty"`
	ReleaseDate     *time.Time `json:"release_date,omitempty"`
	PosterImage     *string    `json:"poster_image,omitempty"`
	DirectorsId     *int       `json:"directors_id,omitempty"`
	Rating          *string    `json:"rating,omitempty"`
	BgPath          *string    `json:"bg_path,omitempty"`
	GenresId        []int      `json:"genres_id,omitempty"`
	CastsId         []int      `json:"casts_id,omitempty"`
	Showtimes       []Showtime `json:"showtimes,omitempty"`
}

type MovieCreateRequest struct {
	Title           string    `json:"title" binding:"required"`
	Synopsis        string    `json:"synopsis" binding:"required"`
	DurationMinutes int       `json:"duration_minutes" binding:"required"`
	ReleaseDate     string    `json:"release_date" binding:"required"`
	PosterImage     string    `json:"poster_image" binding:"required"`
	DirectorsId     int       `json:"directors_id" binding:"required"`
	Rating          float64   `json:"rating" binding:"required"`
	BgPath          *string   `json:"bg_path" binding:"required"`
	Category        []int     `json:"category_id" binding:"required"`
	Location        []int     `json:"location_id" binding:"required"`
	Date            string    `json:"date" binding:"required"`
	Time            time.Time `json:"time" binding:"required"`
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
