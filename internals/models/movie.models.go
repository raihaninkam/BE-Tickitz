package models

import "time"

type UpcomingMovie struct {
	Id              int
	Title           string
	DirectorsId     int
	DirectorName    string
	Rating          *float64
	Synopsis        string
	DurationMinutes int
	ReleaseDate     time.Time
	PosterImage     string
	BgPath          *string
}

type PopularMovie struct {
	Id              int       `db:"id"`
	Title           string    `db:"title"`
	DirectorId      int       `db:"directors_id"`
	DirectorName    string    `db:"director_name"`
	Rating          *float64  `db:"rating"`
	Synopsis        *string   `db:"synopsis"`
	DurationMinutes int       `db:"duration_minutes"`
	ReleaseDate     time.Time `db:"release_date"`
	PosterImage     string    `db:"poster_image"`
	BgPath          *string   `db:"bg_path"`
	AvgRating       *float64  `db:"avg_rating"`
}

type AllMovie struct {
	Id              int       `json:"id"`
	Title           string    `json:"title"`
	Synopsis        string    `json:"synopsis"`
	DurationMinutes int       `json:"duration_minutes"`
	ReleaseDate     time.Time `json:"release_date"`
	PosterImage     string    `json:"poster_image"`
	DirectorsId     int       `json:"directors_id"`
	Rating          *string   `json:"rating"`
	BgPath          *string   `json:"bg_path"`
	DirectorName    string    `json:"director_name"`
	Genres          *string   `json:"genres"`
}

type MovieFilter struct {
	Id              int       `json:"id"`
	Title           string    `json:"title"`
	Synopsis        string    `json:"synopsis"`
	DurationMinutes int       `json:"duration_minutes"`
	ReleaseDate     time.Time `json:"release_date"`
	PosterImage     string    `json:"poster_image"`
	DirectorsId     int       `json:"directors_id"`
	Rating          *string   `json:"rating"`
	BgPath          *string   `json:"bg_path"`
	DirectorName    string    `json:"director_name"`
	Genres          *string   `json:"genres"`
}

type MovieDetail struct {
	ID           int       `db:"id" json:"id"`
	Title        string    `db:"title" json:"title"`
	Synopsis     string    `db:"synopsis" json:"synopsis"`
	ReleaseDate  time.Time `db:"release_date" json:"release_date"`
	Duration     int       `db:"duration_minutes" json:"duration"`
	PosterImage  string    `db:"poster_image" json:"poster_image"`
	BgPath       string    `db:"bg_path" json:"bg_path"`
	DirectorsID  int       `db:"directors_id" json:"directors_id"`
	DirectorName *string   `db:"director_name" json:"director_name"`
	Genres       *string   `db:"genres" json:"genres"`
	Casts        *string   `db:"casts" json:"casts"`
}

type MovieSchedule struct {
	Id           int       `db:"id" json:"id"`
	Date         time.Time `db:"date" json:"date"`
	Time         string    `db:"time" json:"time"`
	LocationId   int       `db:"location_id" json:"city_id"`
	MovieId      int       `db:"movie_id" json:"movie_id"`
	MovieTitle   string    `db:"movie_title" json:"movie_title"`
	LocationName string    `db:"location_name" json:"location_name"`
	CinemaName   string    `db:"cinema_name" json:"cinema_name"` // Tambahan field cinema
}
