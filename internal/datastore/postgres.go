package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/RogueAlmond70/code-review-challenge/internal/config"
	"github.com/pushfar/code-review-challenge/database"
	"go.uber.org/zap"
)

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
