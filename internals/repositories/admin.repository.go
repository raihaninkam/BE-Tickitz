package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raihaninkam/tickitz/internals/models"
)

type MovieAdmin struct {
	Db *pgxpool.Pool
}

func NewMovieAdmin(db *pgxpool.Pool) *MovieAdmin {
	return &MovieAdmin{Db: db}
}

// CREATE
func (ma *MovieAdmin) AddMovie(ctx context.Context, req models.MovieAdmin) (*models.MovieEdit, error) {
	sql := `
		INSERT INTO movies (title, synopsis, duration_minutes, release_date, poster_image, directors_id, rating, bg_path, is_deleted)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,false)
		RETURNING id, title, synopsis, duration_minutes, release_date, poster_image, directors_id, rating, bg_path
	`

	var movie models.MovieEdit
	err := ma.Db.QueryRow(ctx, sql,
		req.Title,
		req.Synopsis,
		req.DurationMinutes,
		req.ReleaseDate,
		req.PosterImage,
		req.DirectorsId,
		req.Rating,
		req.BgPath,
	).Scan(
		&movie.Id,
		&movie.Title,
		&movie.Synopsis,
		&movie.DurationMinutes,
		&movie.ReleaseDate,
		&movie.PosterImage,
		&movie.DirectorsId,
		&movie.Rating,
		&movie.BgPath,
	)
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

// READ (hanya ambil yang belum dihapus)
func (ma *MovieAdmin) GetAllMovies(ctx context.Context) ([]models.MovieAdmin, error) {
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
		WHERE m.is_deleted = false
		GROUP BY m.id, d.name
		ORDER BY m.id
	`

	rows, err := ma.Db.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.MovieAdmin
	for rows.Next() {
		var movie models.MovieAdmin
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
	return movies, nil
}

// UPDATE (hanya bisa update kalau belum dihapus)
func (ma *MovieAdmin) UpdateMovie(ctx context.Context, movieId int, req models.MovieUpdateRequest) (*models.MovieAdmin, error) {
	setClauses := []string{}
	args := []interface{}{}
	argID := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title=$%d", argID))
		args = append(args, *req.Title)
		argID++
	}
	if req.Synopsis != nil {
		setClauses = append(setClauses, fmt.Sprintf("synopsis=$%d", argID))
		args = append(args, *req.Synopsis)
		argID++
	}
	if req.DurationMinutes != nil {
		setClauses = append(setClauses, fmt.Sprintf("duration_minutes=$%d", argID))
		args = append(args, *req.DurationMinutes)
		argID++
	}
	if req.ReleaseDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.ReleaseDate)
		if err != nil {
			return nil, errors.New("invalid release_date format, use YYYY-MM-DD")
		}
		setClauses = append(setClauses, fmt.Sprintf("release_date=$%d", argID))
		args = append(args, parsedDate)
		argID++
	}
	if req.PosterImage != nil {
		setClauses = append(setClauses, fmt.Sprintf("poster_image=$%d", argID))
		args = append(args, *req.PosterImage)
		argID++
	}
	if req.DirectorsId != nil {
		setClauses = append(setClauses, fmt.Sprintf("directors_id=$%d", argID))
		args = append(args, *req.DirectorsId)
		argID++
	}
	if req.Rating != nil {
		setClauses = append(setClauses, fmt.Sprintf("rating=$%d", argID))
		args = append(args, *req.Rating)
		argID++
	}
	// Add the missing BgPath handling
	if req.BgPath != nil {
		setClauses = append(setClauses, fmt.Sprintf("bg_path=$%d", argID))
		args = append(args, *req.BgPath)
		argID++
	}

	if len(setClauses) == 0 {
		return nil, errors.New("no fields to update")
	}

	// First, check if movie exists and is not deleted
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM movies WHERE id=$1 AND deleted_at IS NULL)"
	err := ma.Db.QueryRow(ctx, checkQuery, movieId).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	if !exists {
		return nil, errors.New("movie not found or already deleted")
	}

	// Proceed with update
	query := fmt.Sprintf(`
        UPDATE movies 
        SET %s, updated_at = NOW() 
        WHERE id=$%d AND deleted_at IS NULL 
        RETURNING id, title, synopsis, duration_minutes, release_date, poster_image, directors_id, rating, bg_path
    `, strings.Join(setClauses, ", "), argID)

	args = append(args, movieId)

	row := ma.Db.QueryRow(ctx, query, args...)

	var movie models.MovieAdmin
	err = row.Scan(
		&movie.Id,
		&movie.Title,
		&movie.Synopsis,
		&movie.DurationMinutes,
		&movie.ReleaseDate,
		&movie.PosterImage,
		&movie.DirectorsId,
		&movie.Rating,
		&movie.BgPath,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("movie not found or already deleted")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	return &movie, nil
}

// SOFT DELETE
func (ma *MovieAdmin) DeleteMovie(ctx context.Context, movieId int) error {
	sql := `
		UPDATE movies 
		SET is_deleted = true, deleted_at = $1
		WHERE id = $2 AND is_deleted = false
	`
	res, err := ma.Db.Exec(ctx, sql, time.Now(), movieId)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("movie not found or already deleted")
	}
	return nil
}
