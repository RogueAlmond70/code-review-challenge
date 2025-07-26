package services

import (
	"context"
	"time"

	"github.com/RogueAlmond70/code-review-challenge/internal/models"
	"github.com/RogueAlmond70/code-review-challenge/types"
)

// Though postgres has been chosen for this implementation, we want the flexibility of using whatever database client
// best fits our needs. The DBClient interface has been designed with this in mind.
type DBClient interface {
	GetSingleNote(ctx context.Context, userId string, noteId string) (types.Note, error)
	GetNotes(ctx context.Context, userId string, archiveFilter *bool, limit, offset int) ([]types.Note, int, error)
	CreateNote(ctx context.Context, userId, title, body string) (types.Note, error)
	UpdateNote(ctx context.Context, userId, noteId string, note *types.NoteDto) (types.Note, error)
	DeleteNote(ctx context.Context, userId, noteId string) error
}

// Similarly, this gives us the flexibility to use other cache clients.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type UserStore interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}
