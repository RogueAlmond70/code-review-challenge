package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/pushfar/code-review-challenge/types"
)

func singleNote(db *sql.DB, userId string, noteId string) (types.Note, error) {
	query := `
        SELECT id, title, content, archived
        FROM notes
        WHERE user_id = $1 AND id = $2`

	rows, err := db.Query(query, userId, noteId) // Handle this error please.
	defer rows.Close()

	if err != nil {
		log.Fatal(err)
		return types.Note{}, err
	}

	for rows.Next() {
		var note types.Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.Archived)
		if err != nil {
			log.Fatal(err)
			return types.Note{}, err
		}

		return note, nil
	}

	return types.Note{}, errors.New("could not find note")
}

func AllNotes(db *sql.DB, userId string) ([]types.Note, error) {
	query := `
        SELECT id, title, content, archived
        FROM notes
        WHERE user_id = $1`

	var notes []types.Note

	rows, err := db.Query(query, userId)
	defer rows.Close()

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	for rows.Next() {
		var note types.Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.Archived)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		notes = append(notes, note)
	}

	return notes, nil
}

func ArchivedNotes(db *sql.DB, userId string) ([]types.Note, error) {
	query := `
        SELECT id, title, content, archived
        FROM notes
        WHERE user_id = $1 AND archived = true`

	var notes []types.Note

	rows, err := db.Query(query, userId)
	defer rows.Close()

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	for rows.Next() {
		var note types.Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.Archived)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		notes = append(notes, note)
	}

	return notes, nil
}

func UnarchivedNotes(db *sql.DB, userId string) ([]types.Note, error) {
	query := `
        SELECT id, title, content, archived
        FROM notes
        WHERE user_id = $1 AND archived=false`

	var notes []types.Note

	rows, err := db.Query(query, userId)
	defer rows.Close()

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	for rows.Next() {
		var note types.Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.Archived)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		notes = append(notes, note)
	}

	return notes, nil
}

func CreateNote(db *sql.DB, userId string, title string, body string) (types.Note, error) {
	query := `
        INSERT INTO notes (user_id, title, content, archived)
        VALUES ($1,$2,$3,$4) RETURNING id, title, content, archived`

	rows, err := db.Query(query, userId, title, body, false)
	defer rows.Close()

	if err != nil {
		log.Fatal(err)
		return types.Note{}, err
	}

	var note types.Note

	for rows.Next() {
		fmt.Print(rows)
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.Archived)
		if err != nil {
			log.Fatal(err)
			return types.Note{}, err
		}
	}

	return note, nil
}

func UpdateNote(db *sql.DB, userId string, noteId string, note types.NoteDto) (types.Note, error) {
	oldNote, err := singleNote(db, userId, noteId)

	if err != nil {
		return types.Note{}, err
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

	rows, err := db.Query(query, oldNote.Title, oldNote.Content, oldNote.Archived, noteId, userId)
	defer rows.Close()

	if err != nil {
		log.Fatal(err)
		return types.Note{}, err
	}

	var newNote types.Note

	for rows.Next() {
		err := rows.Scan(&newNote.ID, &newNote.Title, &newNote.Content, &newNote.Archived)
		if err != nil {
			log.Fatal(err)
			return types.Note{}, err
		}
	}

	return newNote, nil
}

func DeleteNote(db *sql.DB, userId string, id string) error {
	query := `
        DELETE FROM notes 
        WHERE id = $1 AND user_id = $2`

	_, err := db.Exec(query, id, userId)

	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
