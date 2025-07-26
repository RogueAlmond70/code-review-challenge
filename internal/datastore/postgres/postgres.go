package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RogueAlmond70/code-review-challenge/internal/config"
	"github.com/RogueAlmond70/code-review-challenge/internal/config/metrics"
	"github.com/RogueAlmond70/code-review-challenge/services"
	"github.com/RogueAlmond70/code-review-challenge/types"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var ErrParameterNotProvided = errors.New("required parameters missing")
var ErrNilNote = errors.New("note is nil")
var ErrNoteNoteFound = errors.New("could not find note")
var _ services.DBClient = &Postgres{}

type Postgres struct {
	logger *zap.Logger
	db     *sql.DB
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
	metrics.CountSingleNoteRequestsTotal.WithLabelValues("count_single_note_requests_total").Inc()

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
		metrics.CountSingleNoteRequestErrorsTotal.WithLabelValues("single_note_request_errors_total").Inc()
		p.logger.Error("userId and noteId must be provided", zap.Error(ErrParameterNotProvided),
			zap.String("userId", userId),
			zap.String("noteId", noteId))

		return types.Note{}, fmt.Errorf("userId and noteId must be provided: %w", ErrParameterNotProvided)
	}

	query := `
        SELECT id, title, content, archived
        FROM notes
        WHERE user_id = $1 AND id = $2`

	var note types.Note
	err := p.db.QueryRowContext(ctx, query, userId, noteId).Scan(
		&note.ID,
		&note.Title,
		&note.Content,
		&note.Archived,
	)

	if err != nil {
		metrics.CountSingleNoteRequestErrorsTotal.WithLabelValues("single_note_request_errors_total").Inc()
		if errors.Is(err, sql.ErrNoRows) {
			p.logger.Error("note not found",
				zap.String("operation_name", "GetSingleNote"),
				zap.String("userId", userId),
				zap.String("noteId", noteId),
				zap.Error(ErrNoteNoteFound),
			)
			return types.Note{}, fmt.Errorf("note not found: %w", ErrNoteNoteFound)
		}
		p.logger.Error("failed to query single note",
			zap.String("operation_name", "GetSingleNote"),
			zap.Error(err),
			zap.String("userId", userId),
		)
		return types.Note{}, fmt.Errorf("failed to get note: %w", err)
	}
	return note, nil
}

func (p *Postgres) GetNotes(ctx context.Context, userId string, archivedFilter *bool, limit, offset int) ([]types.Note, int, error) {
	timer := timerMetricSelection(archivedFilter)
	incrementTotalMetric(archivedFilter)

	defer func() {
		elapsed := timer.ObserveDuration()
		if elapsed > 500*time.Millisecond {
			p.logger.Warn("slow query detected (>500ms)",
				zap.Duration("duration", elapsed),
				zap.String("method", "GetNotes"),
				zap.String("userId", userId),
			)
		}
	}()

	if userId == "" {
		incrementErrorMetric(archivedFilter)
		p.logger.Error("userId must be provided", zap.Error(ErrParameterNotProvided))
		return nil, 0, fmt.Errorf("userId must be provided: %w", ErrParameterNotProvided)
	}

	// Build dynamic WHERE clause
	whereClauses := []string{"user_id = $1"}
	args := []interface{}{userId}
	argIndex := 2

	if archivedFilter != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("archived = $%d", argIndex))
		args = append(args, *archivedFilter)
		argIndex++
	}

	where := strings.Join(whereClauses, " AND ")

	// ----- Total Count Query -----
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notes WHERE %s", where)
	var totalCount int
	if err := p.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		incrementErrorMetric(archivedFilter)
		p.logger.Error("failed to get total count", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// ----- Main Query with Pagination -----
	args = append(args, limit, offset)
	baseQuery := fmt.Sprintf(`
		SELECT id, title, content, archived
		FROM notes
		WHERE %s
		ORDER BY id
		LIMIT $%d OFFSET $%d`, where, argIndex, argIndex+1)

	rows, err := p.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		incrementErrorMetric(archivedFilter)
		p.logger.Error("unable to run sql query", zap.Error(err))
		return nil, 0, fmt.Errorf("unable to run sql query: %w", err)
	}
	defer rows.Close()

	var notes []types.Note
	for rows.Next() {
		var note types.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.Archived); err != nil {
			incrementErrorMetric(archivedFilter)
			p.logger.Error("unable to scan row",
				zap.String("operation_name", "GetNotes"),
				zap.Error(err),
				zap.String("userId", userId))
			return nil, 0, fmt.Errorf("unable to scan row: %w", err)
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		incrementErrorMetric(archivedFilter)
		p.logger.Error("row iteration error", zap.Error(err))
		return nil, 0, fmt.Errorf("row iteration error: %w", err)
	}

	return notes, totalCount, nil
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
	if userId == "" || title == "" || body == "" {
		metrics.CountCreateNoteRequestErrorsTotal.WithLabelValues("create_note_request_errors_total").Inc()
		p.logger.Error("userId, title and body must be provided", zap.Error(ErrParameterNotProvided),
			zap.String("userId", userId),
			zap.String("title", title),
			zap.String("body", body))

		return types.Note{}, fmt.Errorf("userId, title and body must be provided: %w", ErrParameterNotProvided)
	}

	query := `
        INSERT INTO notes (user_id, title, content, archived)
        VALUES ($1,$2,$3,$4) RETURNING id, title, content, archived`

	var newNote types.Note

	err := p.db.QueryRowContext(ctx, query, userId, title, body, false).Scan(
		&newNote.ID,
		&newNote.Title,
		&newNote.Content,
		&newNote.Archived,
	)

	if err != nil {
		metrics.CountCreateNoteRequestErrorsTotal.WithLabelValues("create_note_request_errors_total").Inc()
		p.logger.Error("failed to create note",
			zap.String("operation_name", "CreateNote"),
			zap.Error(err),
			zap.String("userId", userId),
		)
		return types.Note{}, fmt.Errorf("failed to create note: %w", err)
	}

	p.logger.Info("note created",
		zap.String("userId", userId),
		zap.String("noteId", newNote.ID))

	return newNote, nil
}

func (p *Postgres) UpdateNote(ctx context.Context, userId, noteId string, note *types.NoteDto) (types.Note, error) {
	timer := prometheus.NewTimer(metrics.UpdateNoteRequestDurationSeconds)
	metrics.CountUpdateNoteRequestsTotal.WithLabelValues("count_update_note_requests_total").Inc()

	defer func() { // If performance is slow for some reason, it should be logged. An additional option is to define a new metric in prometheus and set up an alert.
		elapsed := timer.ObserveDuration()
		if elapsed > 500*time.Millisecond {
			p.logger.Warn("slow query detected (>500ms)",
				zap.Duration("duration", elapsed),
				zap.String("method", "UpdateNote"),
				zap.String("userId", userId),
				zap.String("noteId", noteId),
			)
		}
	}()

	// Input validation:
	if userId == "" || noteId == "" {
		metrics.CountUpdateNoteRequestErrorsTotal.WithLabelValues("update_note_request_errors_total").Inc()
		p.logger.Error("userId and noteId must be provided", zap.Error(ErrParameterNotProvided),
			zap.String("userId", userId),
			zap.String("noteId", noteId))

		return types.Note{}, fmt.Errorf("userId and noteId must be provided: %w", ErrParameterNotProvided)
	}

	if note == nil { // I can't think of any use case where we would want to allow overwriting with a blank note. We have the Delete method.
		metrics.CountUpdateNoteRequestErrorsTotal.WithLabelValues("update_note_request_errors_total").Inc()
		p.logger.Error("note must not be nil", zap.Error(ErrNilNote))
		return types.Note{}, fmt.Errorf("note must not be nil: %w", ErrNilNote)
	}

	oldNote, err := p.GetSingleNote(ctx, userId, noteId)
	if err != nil {
		metrics.CountUpdateNoteRequestErrorsTotal.WithLabelValues("update_note_request_errors_total").Inc()
		p.logger.Error("unable to update note", zap.Error(err))
		return types.Note{}, fmt.Errorf("unable to update note: %w", err)
	}

	if note.Title != nil {
		oldNote.Title = *note.Title
	}
	if note.Content != nil {
		oldNote.Content = *note.Content
	}
	if note.Archived != nil {
		oldNote.Archived = *note.Archived
	}

	query := `UPDATE notes
		SET title = $1, content = $2, archived = $3
		WHERE id = $4 AND user_id = $5
		RETURNING id, title, content, archived`

	var newNote types.Note

	err = p.db.QueryRowContext(ctx, query, oldNote.Title, oldNote.Content, oldNote.Archived, noteId, userId).Scan(
		&newNote.ID,
		&newNote.Title,
		&newNote.Content,
		&newNote.Archived,
	)

	if err != nil {
		metrics.CountUpdateNoteRequestErrorsTotal.WithLabelValues("update_note_request_errors_total").Inc()
		p.logger.Error("failed to update note",
			zap.String("operation_name", "UpdateNote"),
			zap.Error(err),
			zap.String("userId", userId),
		)
		return types.Note{}, fmt.Errorf("failed to update note: %w", err)
	}

	p.logger.Info("note update",
		zap.String("userId", userId),
		zap.String("noteId", newNote.ID))

	return newNote, nil
}

func (p *Postgres) DeleteNote(ctx context.Context, userId, noteId string) error {
	timer := prometheus.NewTimer(metrics.DeleteNoteRequestDurationSeconds)
	metrics.CountDeleteNoteRequestsTotal.WithLabelValues("count_delete_note_requests_total").Inc()

	defer func() { // If performance is slow for some reason, it should be logged. An additional option is to define a new metric in prometheus and set up an alert.
		elapsed := timer.ObserveDuration()
		if elapsed > 500*time.Millisecond {
			p.logger.Warn("slow query detected (>500ms)",
				zap.Duration("duration", elapsed),
				zap.String("method", "DeleteNote"),
				zap.String("userId", userId),
				zap.String("noteId", noteId),
			)
		}
	}()

	if userId == "" || noteId == "" {
		metrics.CountDeleteNoteRequestErrorsTotal.WithLabelValues("count_delete_note_request_errors_total").Inc()
		p.logger.Error("userId and noteId must be provided", zap.Error(ErrParameterNotProvided),
			zap.String("userId", userId),
			zap.String("noteId", noteId))

		return fmt.Errorf("userId and noteId must be provided: %w", ErrParameterNotProvided)
	}

	query := `
        DELETE FROM notes 
        WHERE id = $1 AND user_id = $2`

	res, err := p.db.ExecContext(ctx, query, noteId, userId)

	if err != nil {
		metrics.CountDeleteNoteRequestErrorsTotal.WithLabelValues("count_delete_note_request_errors_total").Inc()
		p.logger.Error("failed to delete note",
			zap.String("operation_name", "DeleteNote"),
			zap.Error(err),
			zap.String("userId", userId),
		)
		return fmt.Errorf("failed to delete note: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		p.logger.Warn("could not check rows affected after delete", zap.Error(err))
	}

	if rowsAffected == 0 {
		p.logger.Info("no note deleted (not found or not owned by user)",
			zap.String("userId", userId),
			zap.String("noteId", noteId),
		)

	}

	p.logger.Info("note deleted",
		zap.String("userId", userId),
		zap.String("noteId", noteId))

	return nil
}

func timerMetricSelection(archivedFilter *bool) *prometheus.Timer {
	var timer *prometheus.Timer
	switch *archivedFilter {
	case true:
		timer = prometheus.NewTimer(metrics.CountArchivedNotesRequestDurationSeconds)
	case false:
		timer = prometheus.NewTimer(metrics.UnarchivedNotesRequestDurationSeconds)
	default:
		timer = prometheus.NewTimer(metrics.AllNotesRequestDurationSeconds)
	}
	return timer
}

func incrementErrorMetric(archivedFilter *bool) {
	switch {
	case archivedFilter != nil && *archivedFilter:
		metrics.CountArchivedNotesRequestErrorsTotal.WithLabelValues("archived_notes_request_errors_total").Inc()
	case archivedFilter != nil && !*archivedFilter:
		metrics.CountUnarchivedNotesRequestErrorsTotal.WithLabelValues("unarchived_note_request_errors_total").Inc()
	default:
		metrics.CountAllNotesRequestErrorsTotal.WithLabelValues("all_note_request_errors_total").Inc()
	}
}

func incrementTotalMetric(archivedFilter *bool) {
	switch {
	case archivedFilter != nil && *archivedFilter:
		metrics.CountArchivedNotesRequestsTotal.WithLabelValues("archived_notes_requests_total").Inc()
	case archivedFilter != nil && !*archivedFilter:
		metrics.CountUnarchivedNotesRequestsTotal.WithLabelValues("unarchived_note_requests_total").Inc()
	default:
		metrics.CountAllNotesRequestsTotal.WithLabelValues("all_note_requests_total").Inc()
	}
}
