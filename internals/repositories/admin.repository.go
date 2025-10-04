package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raihaninkam/tickitz/internals/models"
	"github.com/raihaninkam/tickitz/internals/utils"
	"github.com/redis/go-redis/v9"
)

type MovieAdmin struct {
	Db  *pgxpool.Pool
	Rdb *redis.Client
}

func NewMovieAdmin(db *pgxpool.Pool, rdb *redis.Client) *MovieAdmin {
	return &MovieAdmin{Db: db,
		Rdb: rdb}
}

// CREATE
func (ma *MovieAdmin) AddMovie(ctx context.Context, req models.MovieAdmin) (*models.MovieEdit, error) {
	// Start transaction
	tx, err := ma.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert movie
	sqlMovie := `
		INSERT INTO movies (title, synopsis, duration_minutes, release_date, poster_image, directors_id, rating, bg_path, is_deleted)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,false)
		RETURNING id, title, synopsis, duration_minutes, release_date, poster_image, directors_id, rating, bg_path
	`

	var movie models.MovieEdit
	err = tx.QueryRow(ctx, sqlMovie,
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
		return nil, fmt.Errorf("failed to insert movie: %w", err)
	}

	// Insert genres
	if len(req.GenresId) > 0 {
		sqlGenre := `INSERT INTO movies_genre (movies_id, genres_id) VALUES ($1, $2)`
		for _, genreId := range req.GenresId {
			_, err = tx.Exec(ctx, sqlGenre, movie.Id, genreId)
			if err != nil {
				return nil, fmt.Errorf("failed to insert movie genre: %w", err)
			}
		}
	}

	// Insert casts
	if len(req.CastsId) > 0 {
		sqlCast := `INSERT INTO movies_casts (movies_id, casts_id) VALUES ($1, $2)`
		for _, castId := range req.CastsId {
			_, err = tx.Exec(ctx, sqlCast, movie.Id, castId)
			if err != nil {
				return nil, fmt.Errorf("failed to insert movie cast: %w", err)
			}
		}
	}

	// Insert showtimes ke tabel now_showing
	if len(req.Showtimes) > 0 {
		sqlShowtime := `
			INSERT INTO now_showing (date, time, location_id, movie_id, cinemas_id) 
			VALUES ($1, $2, $3, $4, $5)
		`
		for _, showtime := range req.Showtimes {
			_, err = tx.Exec(ctx, sqlShowtime,
				showtime.Date,
				showtime.Time,
				showtime.LocationId,
				movie.Id, // movie_id dari hasil insert movie
				showtime.CinemasId,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to insert showtime: %w", err)
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	_ = utils.InvalidateCache(ctx, ma.Rdb, "all_movies", "upcoming_movies", "popular_movies")

	return &movie, nil
}

// READ (hanya ambil yang belum dihapus)
func (ma *MovieAdmin) GetAllMovies(ctx context.Context) ([]models.MovieAdmin, error) {
	sql := `
		SELECT 
            m.id, 
            m.title, 
            m.synopsis, 
            m.duration_minutes, 
            m.release_date, 
            m.poster_image, 
            m.bg_path,
            m.directors_id,
            m.rating,
            d.name as director_name
        FROM movies m
        LEFT JOIN directors d ON m.directors_id = d.id
        ORDER BY m.id ASC
	`

	rows, err := ma.Db.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.MovieAdmin
	for rows.Next() {
		var movie models.MovieAdmin
		var directorName string

		if err := rows.Scan(
			&movie.Id,
			&movie.Title,
			&movie.Synopsis,
			&movie.DurationMinutes,
			&movie.ReleaseDate,
			&movie.PosterImage,
			&movie.BgPath,
			&movie.DirectorsId,
			&movie.Rating,
			&directorName,
		); err != nil {
			return nil, err
		}

		// Query genre IDs AND names for each movie
		genreSQL := `
			SELECT g.id, g.name 
			FROM movies_genre mg
			JOIN genres g ON mg.genres_id = g.id
			WHERE mg.movies_id = $1
		`

		genreRows, err := ma.Db.Query(ctx, genreSQL, movie.Id)
		if err != nil {
			return nil, err
		}

		var genreIds []int
		var genreNames []string
		for genreRows.Next() {
			var genreId int
			var genreName string
			if err := genreRows.Scan(&genreId, &genreName); err != nil {
				genreRows.Close()
				return nil, err
			}
			genreIds = append(genreIds, genreId)
			genreNames = append(genreNames, genreName)
		}
		genreRows.Close()

		// Set genres_id (array of IDs)
		if len(genreIds) == 0 {
			movie.GenresId = []int{}
			movie.Genres = "N/A"
		} else {
			movie.GenresId = genreIds
			movie.Genres = strings.Join(genreNames, ", ") // Join names with comma
		}

		movies = append(movies, movie)
	}

	if len(movies) == 0 {
		return nil, errors.New("no movies found")
	}

	return movies, nil
}

// UPDATE (hanya bisa update kalau belum dihapus)
// UPDATE MOVIE COMPREHENSIVE - seperti AddMovie tapi support partial update
func (ma *MovieAdmin) UpdateMovieComprehensive(ctx context.Context, movieId int, req models.MovieUpdateComprehensiveRequest) (*models.MovieEdit, error) {
	// Start transaction
	tx, err := ma.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if movie exists and not deleted
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM movies WHERE id = $1 AND is_deleted = false)", movieId).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check movie existence: %w", err)
	}
	if !exists {
		return nil, errors.New("movie not found or already deleted")
	}

	// Update movie basic info (partial update)
	var movie models.MovieEdit
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
		setClauses = append(setClauses, fmt.Sprintf("release_date=$%d", argID))
		args = append(args, *req.ReleaseDate)
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
	if req.BgPath != nil {
		setClauses = append(setClauses, fmt.Sprintf("bg_path=$%d", argID))
		args = append(args, *req.BgPath)
		argID++
	}

	// Jika ada field basic info yang diupdate
	if len(setClauses) > 0 {
		query := fmt.Sprintf(`
			UPDATE movies 
			SET %s, update_at = NOW() 
			WHERE id = $%d AND is_deleted = false
			RETURNING id, title, synopsis, duration_minutes, release_date, poster_image, directors_id, rating, bg_path
		`, strings.Join(setClauses, ", "), argID)
		args = append(args, movieId)

		err = tx.QueryRow(ctx, query, args...).Scan(
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
			return nil, fmt.Errorf("failed to update movie: %w", err)
		}
	} else {
		// Jika tidak ada update basic info, ambil data movie yang existing
		err = tx.QueryRow(ctx, `
			SELECT id, title, synopsis, duration_minutes, release_date, poster_image, directors_id, rating, bg_path
			FROM movies WHERE id = $1 AND is_deleted = false
		`, movieId).Scan(
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
			return nil, fmt.Errorf("failed to get movie data: %w", err)
		}
	}

	// Update genres jika provided (tidak nil)
	if req.GenresId != nil {
		// Delete existing genres
		_, err = tx.Exec(ctx, "DELETE FROM movies_genre WHERE movies_id = $1", movieId)
		if err != nil {
			return nil, fmt.Errorf("failed to delete existing genres: %w", err)
		}

		// Insert new genres jika array tidak kosong
		if len(req.GenresId) > 0 {
			sqlGenre := `INSERT INTO movies_genre (movies_id, genres_id) VALUES ($1, $2)`
			for _, genreId := range req.GenresId {
				_, err = tx.Exec(ctx, sqlGenre, movieId, genreId)
				if err != nil {
					return nil, fmt.Errorf("failed to insert movie genre: %w", err)
				}
			}
		}
	}

	// Update casts jika provided (tidak nil)
	if req.CastsId != nil {
		// Delete existing casts
		_, err = tx.Exec(ctx, "DELETE FROM movies_casts WHERE movies_id = $1", movieId)
		if err != nil {
			return nil, fmt.Errorf("failed to delete existing casts: %w", err)
		}

		// Insert new casts jika array tidak kosong
		if len(req.CastsId) > 0 {
			sqlCast := `INSERT INTO movies_casts (movies_id, casts_id) VALUES ($1, $2)`
			for _, castId := range req.CastsId {
				_, err = tx.Exec(ctx, sqlCast, movieId, castId)
				if err != nil {
					return nil, fmt.Errorf("failed to insert movie cast: %w", err)
				}
			}
		}
	}

	// Update showtimes jika provided (tidak nil)
	if req.Showtimes != nil {
		// Delete existing showtimes
		_, err = tx.Exec(ctx, "DELETE FROM now_showing WHERE movie_id = $1", movieId)
		if err != nil {
			return nil, fmt.Errorf("failed to delete existing showtimes: %w", err)
		}

		// Insert new showtimes jika array tidak kosong
		if len(req.Showtimes) > 0 {
			sqlShowtime := `
				INSERT INTO now_showing (date, time, location_id, movie_id, cinemas_id) 
				VALUES ($1, $2, $3, $4, $5)
			`
			for _, showtime := range req.Showtimes {
				_, err = tx.Exec(ctx, sqlShowtime,
					showtime.Date,
					showtime.Time,
					showtime.LocationId,
					movieId,
					showtime.CinemasId,
				)
				if err != nil {
					return nil, fmt.Errorf("failed to insert showtime: %w", err)
				}
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache
	_ = utils.InvalidateCache(ctx, ma.Rdb,
		"all_movies",
		"upcoming_movies",
		"popular_movies",
		fmt.Sprintf("movie_detail:%d", movieId),
	)

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

	_ = utils.InvalidateCache(ctx, ma.Rdb,
		"all_movies",
		"upcoming_movies",
		"popular_movies",
		fmt.Sprintf("movie_detail:%d", movieId),
	)

	return nil
}
