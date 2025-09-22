package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raihaninkam/tickitz/internals/models"
	"github.com/raihaninkam/tickitz/internals/utils"
	"github.com/redis/go-redis/v9"
)

// all movie

type AllMovie struct {
	Db  *pgxpool.Pool
	Rdb *redis.Client
}

func NewAllMovies(db *pgxpool.Pool, rdb *redis.Client) *AllMovie {
	return &AllMovie{Db: db, Rdb: rdb}
}

func (am *AllMovie) GetAllMovies(ctx context.Context) ([]models.AllMovie, error) {
	redisKey := "all_movies"

	// coba ambil dari cache
	if cached, ok := utils.GetFromCache[models.AllMovie](ctx, am.Rdb, redisKey); ok {
		return cached, nil
	}

	sql := `
		SELECT 
			m.id, m.title, m.synopsis, m.duration_minutes, m.release_date, 
			m.poster_image, m.directors_id, m.rating, m.bg_path, 
			d.name as director_name,
			STRING_AGG(DISTINCT g.name, ', ') as genres
		FROM movies m
		JOIN directors d ON m.directors_id = d.id
		LEFT JOIN movies_genre mg ON m.id = mg.movies_id
		LEFT JOIN genres g ON mg.genres_id = g.id
		GROUP BY m.id, d.name
		ORDER BY m.id
	`

	rows, err := am.Db.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.AllMovie
	for rows.Next() {
		var movie models.AllMovie
		if err := rows.Scan(
			&movie.Id,
			&movie.Title,
			&movie.Synopsis,
			&movie.DurationMinutes,
			&movie.ReleaseDate,
			&movie.PosterImage,
			&movie.DirectorsId,
			&movie.Rating,
			&movie.BgPath,
			&movie.DirectorName,
			&movie.Genres,
		); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	if len(movies) == 0 {
		return nil, errors.New("no movies found")
	}

	// simpan ke cache 15 menit
	utils.SetToCache(ctx, am.Rdb, redisKey, movies, 15*time.Minute)

	return movies, nil
}

// upcoming movie

type UpcomingMovie struct {
	Db  *pgxpool.Pool
	Rdb *redis.Client
}

func NewUpcomingMovie(db *pgxpool.Pool, rdb *redis.Client) *UpcomingMovie {
	return &UpcomingMovie{Db: db, Rdb: rdb}
}

func (u *UpcomingMovie) GetUpcomingMovies(ctx context.Context) ([]models.UpcomingMovie, error) {
	redisKey := "upcoming_movies"

	if cached, ok := utils.GetFromCache[models.UpcomingMovie](ctx, u.Rdb, redisKey); ok {
		return cached, nil
	}

	sql := `
		SELECT m.id, m.title, m.directors_id, d.name as director_name,
		       m.rating, m.synopsis, m.duration_minutes, 
		       m.release_date, m.poster_image, m.bg_path
		FROM movies m
		JOIN directors d ON m.directors_id = d.id
		WHERE m.release_date > CURRENT_DATE
		ORDER BY m.release_date ASC
	`

	rows, err := u.Db.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.UpcomingMovie
	for rows.Next() {
		var mv models.UpcomingMovie
		if err := rows.Scan(
			&mv.Id,
			&mv.Title,
			&mv.DirectorsId,
			&mv.DirectorName,
			&mv.Rating,
			&mv.Synopsis,
			&mv.DurationMinutes,
			&mv.ReleaseDate,
			&mv.PosterImage,
			&mv.BgPath,
		); err != nil {
			return nil, err
		}
		movies = append(movies, mv)
	}

	if len(movies) == 0 {
		return nil, errors.New("no upcoming movies found")
	}

	utils.SetToCache(ctx, u.Rdb, redisKey, movies, 24*time.Hour)

	return movies, nil
}

// get popular movie

type PopularMovie struct {
	Db  *pgxpool.Pool
	Rdb *redis.Client
}

func NewPopularMovie(db *pgxpool.Pool, rdb *redis.Client) *PopularMovie {
	return &PopularMovie{Db: db, Rdb: rdb}
}

func (u *PopularMovie) GetPopularMovies(ctx context.Context) ([]models.PopularMovie, error) {
	redisKey := "popular_movies"

	if cached, ok := utils.GetFromCache[models.PopularMovie](ctx, u.Rdb, redisKey); ok {
		return cached, nil
	}

	sql := `
        SELECT 
            m.id,
            m.title,
            m.directors_id,
            d.name as director_name,
            m.rating,
            m.synopsis,
            m.duration_minutes,
            m.release_date,
            m.poster_image,
            m.bg_path,
            m.rating as avg_rating
        FROM movies m
        JOIN directors d ON m.directors_id = d.id
        ORDER BY m.rating DESC NULLS LAST
    `

	rows, err := u.Db.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var popularMovies []models.PopularMovie
	for rows.Next() {
		var pmv models.PopularMovie
		if err := rows.Scan(
			&pmv.Id,
			&pmv.Title,
			&pmv.DirectorId,
			&pmv.DirectorName,
			&pmv.Rating,
			&pmv.Synopsis,
			&pmv.DurationMinutes,
			&pmv.ReleaseDate,
			&pmv.PosterImage,
			&pmv.BgPath,
			&pmv.AvgRating,
		); err != nil {
			return nil, err
		}
		popularMovies = append(popularMovies, pmv)
	}

	if len(popularMovies) == 0 {
		return nil, errors.New("no popular movies found")
	}

	utils.SetToCache(ctx, u.Rdb, redisKey, popularMovies, 10*time.Minute)

	return popularMovies, nil
}

// movie filter

type MovieFilter struct {
	Db *pgxpool.Pool
}

func NewMovieFilter(db *pgxpool.Pool) *MovieFilter {
	return &MovieFilter{Db: db}
}

func (mf *MovieFilter) GetMoviesWithFilter(ctx context.Context, title string, genres []string, offset, limit int) ([]models.MovieFilter, int, error) {
	// Base query untuk data
	sql := `
        SELECT 
            m.id, m.title, m.synopsis, m.duration_minutes, m.release_date, 
            m.poster_image, m.directors_id, m.rating, m.bg_path, 
            d.name as director_name, 
            STRING_AGG(DISTINCT g.name, ', ') as genres
        FROM movies m 
        JOIN directors d ON m.directors_id = d.id 
        LEFT JOIN movies_genre mg ON m.id = mg.movies_id 
        LEFT JOIN genres g ON mg.genres_id = g.id
    `

	// Build WHERE conditions
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	if title != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("m.title ILIKE $%d", argIndex))
		args = append(args, "%"+title+"%")
		argIndex++
	}

	if len(genres) > 0 {
		genrePlaceholders := make([]string, len(genres))
		for i, genre := range genres {
			genrePlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, genre)
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf("g.name IN (%s)", strings.Join(genrePlaceholders, ",")))
	}

	if len(whereConditions) > 0 {
		sql += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	sql += fmt.Sprintf(`
        GROUP BY m.id, d.name 
        ORDER BY m.release_date DESC 
        LIMIT $%d OFFSET $%d
    `, argIndex, argIndex+1)

	args = append(args, limit, offset)

	// Query data
	rows, err := mf.Db.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var movies []models.MovieFilter
	for rows.Next() {
		var movie models.MovieFilter
		if err := rows.Scan(
			&movie.Id,
			&movie.Title,
			&movie.Synopsis,
			&movie.DurationMinutes,
			&movie.ReleaseDate,
			&movie.PosterImage,
			&movie.DirectorsId,
			&movie.Rating,
			&movie.BgPath,
			&movie.DirectorName,
			&movie.Genres,
		); err != nil {
			return nil, 0, err
		}
		movies = append(movies, movie)
	}

	// Query total count (tanpa limit & offset)
	countSql := `
        SELECT COUNT(DISTINCT m.id)
        FROM movies m
        LEFT JOIN movies_genre mg ON m.id = mg.movies_id
        LEFT JOIN genres g ON mg.genres_id = g.id
    `
	var countArgs []interface{}
	countArgIndex := 1
	var countWhere []string

	if title != "" {
		countWhere = append(countWhere, fmt.Sprintf("m.title ILIKE $%d", countArgIndex))
		countArgs = append(countArgs, "%"+title+"%")
		countArgIndex++
	}
	if len(genres) > 0 {
		genrePlaceholders := make([]string, len(genres))
		for i := range genres {
			genrePlaceholders[i] = fmt.Sprintf("$%d", countArgIndex)
			countArgs = append(countArgs, genres[i])
			countArgIndex++
		}
		countWhere = append(countWhere, fmt.Sprintf("g.name IN (%s)", strings.Join(genrePlaceholders, ",")))
	}
	if len(countWhere) > 0 {
		countSql += " WHERE " + strings.Join(countWhere, " AND ")
	}

	var totalCount int
	err = mf.Db.QueryRow(ctx, countSql, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	return movies, totalCount, nil
}

// movie detail

type MovieDetail struct {
	db *pgxpool.Pool
}

func NewMovieDetail(db *pgxpool.Pool) *MovieDetail {
	return &MovieDetail{db: db}
}

// GetDetailMovie mengambil detail movie berdasarkan ID dengan director, genres, dan casts
func (m *MovieDetail) GetDetailMovie(rctx context.Context, movieID int) (models.MovieDetail, error) {
	sql := `SELECT 
    m.id,
    m.title,
    m.synopsis,
    m.release_date,
    m.duration_minutes,
    m.poster_image,
	m.bg_path,
    m.directors_id,
    d.name as director_name,
    STRING_AGG(DISTINCT g.name, ', ') as genres,
    STRING_AGG(DISTINCT c.name, ', ') as casts
FROM movies m
JOIN directors d ON m.directors_id = d.id
LEFT JOIN movies_genre mg ON m.id = mg.movies_id
LEFT JOIN genres g ON mg.genres_id = g.id
LEFT JOIN movies_casts mc ON m.id = mc.movies_id
LEFT JOIN casts c ON mc.casts_id = c.id
WHERE m.id = $1
GROUP BY m.id, d.name
`

	var movie models.MovieDetail
	if err := m.db.QueryRow(rctx, sql, movieID).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Synopsis,
		&movie.ReleaseDate,
		&movie.Duration,
		&movie.PosterImage,
		&movie.BgPath,
		&movie.DirectorsID,
		&movie.DirectorName,
		&movie.Genres,
		&movie.Casts,
	); err != nil {
		if err == pgx.ErrNoRows {
			return models.MovieDetail{}, errors.New("movie not found")
		}
		log.Println("Internal Server Error.\nCz: ", err.Error())
		return models.MovieDetail{}, err
	}
	return movie, nil
}

// movie schedule

type Schedule struct {
	Db *pgxpool.Pool
}

func NewSchedule(db *pgxpool.Pool) *Schedule {
	return &Schedule{Db: db}
}

func (s *Schedule) GetSchedulesByMovieID(ctx context.Context, movieID int) ([]models.MovieSchedule, error) {
	sql := `
        SELECT 
            ns.id,
            ns.movie_id,
            ns.location_id,
            ns.date,
            ns.time,
            m.title as movie_title, 
            l.name as location_name,
            c.cinema_name as cinema_name
        FROM now_showing ns
        JOIN movies m ON ns.movie_id = m.id
        JOIN location l ON ns.location_id = l.id
        JOIN cinemas c ON ns.cinemas_id = c.id
        WHERE ns.movie_id = $1
        ORDER BY ns.date, ns.time
    `

	rows, err := s.Db.Query(ctx, sql, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.MovieSchedule
	for rows.Next() {
		var schedule models.MovieSchedule
		if err := rows.Scan(
			&schedule.Id,
			&schedule.MovieId,
			&schedule.LocationId,
			&schedule.Date,
			&schedule.Time,
			&schedule.MovieTitle,
			&schedule.LocationName,
			&schedule.CinemaName,
		); err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	if len(schedules) == 0 {
		return nil, errors.New("no schedules found for this movie")
	}

	return schedules, nil
}
