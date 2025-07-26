package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/RogueAlmond70/code-review-challenge/database"
	"github.com/RogueAlmond70/code-review-challenge/internal/config"
	"github.com/RogueAlmond70/code-review-challenge/internal/config/metrics"
	"github.com/RogueAlmond70/code-review-challenge/types"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var ErrParameterNotProvided = errors.New("required parameters missing")
var ErrNoteNoteFound = errors.New("could not find note")
var _ database.DBClient = &Postgres{}

type Postgres struct {
	logger *zap.Logger
	db     *slq.DB
	cfg    config.Config
}

func NewPostgres(logger *zap.Logger, sqldb *sql.DB, cfg config.Config) *Postgres {
	return &Postgres{
		logger: logger,
		db:     sqldb,
		cfg:    cfg,
	}
}

func getPostgresConnStr(cfg *config.Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDB,
	)
}

func ConnectDB(cfg config.Config) (*sql.DB, error) {
	connStr := getPostgresConnStr(&cfg)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error checking database connection: %w", err)
	}
	return db, nil
}

func ConnectDBWithRetry(cfg config.Config, zlog zap.Logger) (*sql.DB, error) {
	var db *sql.DB
	var err error
	for i := 0; i < cfg.PostgresRetry; i++ {
		db, err = ConnectDB(cfg)
		if err == nil {
			return db, nil
		}
		zlog.Warn("Failed to connect to DB",
			zap.Int("attempt", i+1),
			zap.Int("max_attempts", cfg.PostgresRetry),
			zap.Error(err),
		)
		time.Sleep(cfg.PostgresDelay)
	}
	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", cfg.PostgresRetry, err)
}

func (p *Postgres) GetSingleNote(ctx context.Context, userId string, noteId string) (types.Note, error) {
	timer := prometheus.NewTimer(metrics.SingleNoteRequestDurationSeconds)
	metrics.CountSingleNoteRequestsTotal.WithLabelalues("count_single_note_requests_total").Inc()

	defer func() { // If performance is slow for some reason, it should be logged. An additional option is to define a new metric in prometheus and set up an alert.
		elapsed := timer.ObserveDuration()
		if elapsed > 500*time.Millisecond {
			p.logger.Warn("slow query detected (>500ms)",
				zap.Duration("duration", elapsed),
				zap.String("method", "GetSingleNote"),
				zap.String("userId", userId),
			)
		}
	}()

	// Input validation:
	if userId == "" || noteId == "" {
		metrics.SingleNoteRequestErrorsTotal.WithLabelalues("single_note_request_errors_total").Inc()
		p.logger.Error("userId and noteId must be provided", zap.Error(ErrParameterNotProvided))
		return types.Note{}, fmt.Errorf("userId and noteId must be provided: %w", ErrParameterNotProvided)
	}

	query := `
        SELECT id, title, content, archived
        FROM notes
        WHERE user_id = $1 AND id = $2`

	rows, err := p.db.QueryContext(ctx, query, userId, noteId)

	if err != nil {
		metrics.SingleNoteRequestErrorsTotal.WithLabelalues("single_note_request_errors_total").Inc()
		p.logger.Error("unable to run sql query", zap.Error(err))
		return types.Note{}, fmt.Errorf("unable to run sql query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var note types.Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.Archived)
		if err != nil {
			metrics.SingleNoteRequestErrorsTotal.WithLabelalues("single_note_request_errors_total").Inc()
			p.logger.Error("unable to scan rows",
				zap.String("operation_name", "GetSingleNote"),
				zap.Error(err),
				zap.String("userId", userId))

			return types.Note{}, fmt.Errorf("unable to scan rows: %w", err)
		}
		return note, nil
	}

	metrics.SingleNoteRequestErrorsTotal.WithLabelalues("single_note_request_errors_total").Inc()
	p.logger.Error("unable to obtain note",
		zap.String("operation_name", "GetSingleNote"),
		zap.Error(ErrNoteNoteFound),
		zap.String("userId", userId))

	return types.Note{}, fmt.Errorf("unable to obtain note: %w", ErrNoteNoteFound)
}

// Get Notes covers functionality for GetAllNotes, GetArchivedNotes, and GetUnarchivedNotes.
func (p *Postgres) GetNotes(ctx context.Context, userId string, archivedFilter *bool) ([]types.Note, error) {
	timer := timerMetricSelection(archivedFilter)
	p.incrementTotalMetric(archivedFilter)

	defer func() { // If performance is slow for some reason, it should be logged. An additional option is to define a new metric in prometheus and set up an alert.
		elapsed := timer.ObserveDuration()
		if elapsed > 500*time.Millisecond {
			p.logger.Warn("slow query detected (>500ms)",
				zap.Duration("duration", elapsed),
				zap.String("method", "GetNotes"),
				zap.String("userId", userId),
			)
		}
	}()

	// Input validation:
	if userId == "" {
		p.incrementErrorMetric(archivedFilter)
		p.logger.Error("userId must be provided", zap.Error(ErrParameterNotProvided))
		return []types.Note{}, fmt.Errorf("userId must be provided: %w", ErrParameterNotProvided)
	}

	baseQuery := `
	SELECT id, title, content, archived
	FROM notes
	WHERE user_id = $1`

	args := []interface{}{userId}

	if archivedFilter != nil {
		baseQuery += " AND archived = $2"
		args = append(args, *archivedFilter)
	}

	rows, err := p.db.QueryContext(ctx, baseQuery, args...)

	if err != nil {
		p.incrementErrorMetric(archivedFilter)
		p.logger.Error("unable to run sql query", zap.Error(err))
		return []types.Note{}, fmt.Errorf("unable to run sql query: %w", err)
	}
	defer rows.Close()

	var notes []types.Note

	for rows.Next() {
		var note types.Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.Archived)
		if err != nil {
			p.incrementErrorMetric(archivedFilter)
			p.logger.Error("unable to scan rows",
				zap.String("operation_name", "GetNotes"),
				zap.Error(err),
				zap.String("userId", userId))

			return []types.Note{}, fmt.Errorf("unable to scan rows: %w", err)
		}

		notes = append(notes, note)

	}

	if err := rows.Err(); err != nil {
		p.incrementErrorMetric(archivedFilter)
		p.logger.Error("row iteration error", zap.Error(err))
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	if len(notes) == 0 {
		return []types.Note{}, nil
	}
	return notes, nil
}

func (p *Postgres) CreateNote(ctx context.Context, userId, title, body string) (types.Note, error) {
	timer := prometheus.NewTimer(metrics.CreateNoteRequestDurationSeconds)
	metrics.CountCreateNoteRequestsTotal.WithLabelValues("count_create_note_requests_total").Inc()

	defer func() { // If performance is slow for some reason, it should be logged. An additional option is to define a new metric in prometheus and set up an alert.
		elapsed := timer.ObserveDuration()
		if elapsed > 500*time.Millisecond {
			p.logger.Warn("slow query detected (>500ms)",
				zap.Duration("duration", elapsed),
				zap.String("method", "CreateNote"),
				zap.String("userId", userId),
			)
		}
	}()

	// Input validation:
	if userId == "" || noteId == "" {
		metrics.CountCreateNoteRequestErrorsTotal.WithLabelalues("create_note_request_errors_total").Inc()
		p.logger.Error("userId and noteId must be provided", zap.Error(ErrParameterNotProvided))
		return types.Note{}, fmt.Errorf("userId and noteId must be provided: %w", ErrParameterNotProvided)
	}

	query := `
        INSERT INTO notes (user_id, title, content, archived)
        VALUES ($1,$2,$3,$4) RETURNING id, title, content, archived`

	rows, err := p.db.QueryContext(ctx, query, userId, title, body, false)

	if err != nil {
		metrics.CountCreateNoteRequestErrorsTotal.WithLabelValues("create_note_requests_total").Inc()
		p.logger.Error("failed to create note",
			zap.String("operation_name", "CreateNote"),
			zap.Error(err),
			zap.String("userId", userId),
			zap.String("noteId", noteId),
		)
		return types.Note{}, fmt.Errorf("failed to create note: %w", err)
	}

	p.logger.Info("note created",
		zap.String("userId", userId),
		zapString("noteId", noteId))

	return nil
}

func (p *Postgres) UpdateNote(ctx context.Context, userId, noteId string, note types.NoteDto) (types.Note, error) {
	return nil
}

func (p *Postgres) DeleteNote(ctx context.Context, userId, noteId string) error {
	return nil
}

func (p *Postgres) timerMetricSelection(acrhivedFilter *bool) {
	var timer *prometheus.Timer
	switch operation {
	case "archived":
		timer = prometheus.NewTimer(metrics.CountArchivedNotesRequestDurationSeconds)
	case "unarchived":
		timer = prometheus.NewTimer(metrics.UnarchivedNotesRequestDurationSeconds)
	default:
		timer = prometheus.NewTimer(metrics.AllNotesRequestDurationSeconds)
	}
	return timer
}

func (p *Postgres) incrementErrorMetric(archivedFilter *bool) {
	switch {
	case archivedFilter != nil && *archivedFilter:
		metrics.CountArchivedNotesRequestErrorsTotal.WithLabelValues("archived_notes_request_errors_total").Inc()
	case archivedFilter != nil && !*archivedFilter:
		metrics.CountUnarchivedNotesRequestErrorsTotal.WithLabelValues("unarchived_note_request_errors_total").Inc()
	default:
		metrics.CountAllNotesRequestErrorsTotal.WithLabelValues("all_note_request_errors_total").Inc()
	}
}

func (p *Postgres) incremenTotalMetric(archivedFilter *bool) {
	switch {
	case archivedFilter != nil && *archivedFilter:
		metrics.CountArchivedNotesRequestsTotal.WithLabelValues("archived_notes_requests_total").Inc()
	case archivedFilter != nil && !*archivedFilter:
		metrics.CountUnarchivedNotesRequestsTotal.WithLabelValues("unarchived_note_requests_total").Inc()
	default:
		metrics.CountAllNotesRequestsTotal.WithLabelValues("all_note_requests_total").Inc()
	}
}
