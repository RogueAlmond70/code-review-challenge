package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/RogueAlmond70/code-review-challenge/internal/models"
	"github.com/google/uuid"
)

type userStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) UserStore {
	return &userStore{db: db}
}

func (s *userStore) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
	`

	id := uuid.New().String()
	_, err := s.db.ExecContext(ctx, query, id, user.Username, user.PasswordHash, user.CreatedAt)
	if err != nil {
		return err
	}

	user.UserId = id
	return nil
}

func (s *userStore) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, password_hash, created_at FROM users WHERE username = $1`

	row := s.db.QueryRowContext(ctx, query, username)

	var user models.User
	if err := row.Scan(&user.UserId, &user.Username, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
