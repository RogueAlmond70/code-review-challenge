package services

// Firstly, this code should be in a vendor specific implementation file
// We shouldn't directly be using this, but rather a DB interface

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "mysecretpassword"
	dbname   = "postgres"
)

func OpenDB() (*sql.DB, error) {
	connectionInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err // Errors should be wrapped, and informative. Also logged, and captured with metrics.
	}

	fmt.Println("setting limits") // Use logger, not print statements
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
