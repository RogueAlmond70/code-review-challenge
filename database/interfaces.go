package database

import (
	"context"
	"github.com/pushfar/code-review-challenge/types"
	"time"
)

// Though postgres has been chosen for this implementation, we want the flexibility of using whatever database client
// best fits our needs. The DBClient interface has been designed with this in mind.
type DBClient interface {
	GetSingleNote(ctx context.Context, userId string, noteId string) (types.Note, error)
	GetNotes(ctx context.Context, userId string) ([]types.Note, error)
	//GetAllNotes(ctx context.Context, userId string) ([]types.Note, error)
	//GetArchivedNotes(ctx context.Context, userId string) ([]types.Note, error)
	//GetUnarchivedNotes(ctx context.Context, userId string) ([]types.Note, error)
	CreateNote(ctx context.Context, userId, title, body string) (types.Note, error)
	UpdateNote(ctx context.Context, userId, noteId string, note types.NoteDto) (types.Note, error)
	DeleteNote(ctx context.Context, userId, noteId string) error
}

// Similarly, this gives us the flexibility to use other cache clients.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
