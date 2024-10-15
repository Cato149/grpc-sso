package sqlite

import (
	"awesomeProject/internal/domain/model"
	"awesomeProject/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.sqlite.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO users (email, pass_hash) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrAlreadyExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// User returns user by email
func (s *Storage) User(ctx context.Context, email string) (model.User, error) {
	op := "storage.sqlite.User"
	var user model.User

	stmt, err := s.db.Prepare("SELECT id, email, pass_hash FROM users WHERE email = ?")
	if err != nil {
		return model.User{}, fmt.Errorf("%s: %w", op, err)
	}

	res := stmt.QueryRowContext(ctx, email)

	err = res.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, storage.ErrNotFound
		}

		return model.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

// IsAdmin returns if user is admin
func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	op := "storage.sqlite.IsAdmin"
	var isAdmin bool

	stmt, err := s.db.Prepare("SELECT is_admin FROM users WHERE id = ?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	res := stmt.QueryRowContext(ctx, userID)

	err = res.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrAppNotFound
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}
	return isAdmin, nil
}

func (s *Storage) App(ctx context.Context, appID int64) (model.App, error) {
	op := "storage.sqlite.App"
	var app model.App

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
	if err != nil {
		return model.App{}, fmt.Errorf("%s: %w", op, err)
	}

	res := stmt.QueryRowContext(ctx, appID)
	err = res.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.App{}, storage.ErrNotFound
		}

		return model.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}
