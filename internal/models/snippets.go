package models

import (
	"context"
	"database/sql"
	"errors"
	"html/template"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Snippet struct {
	Id      int
	Title   string
	Content template.HTML
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *pgxpool.Pool
}

func (m *SnippetModel) Insert(Title string, content string, expires int) (int, error) {
	var id int
	err := m.DB.QueryRow(context.Background(),
		`INSERT INTO snippets (title, content, created, expires)
        VALUES($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '1 day' * $3)
        RETURNING id`,
		Title, content, expires).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *SnippetModel) Get(id int) (Snippet, error) {
	query := `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > NOW() AND id = $1`

	row := m.DB.QueryRow(context.Background(), query, id)

	var s Snippet
	var content string

	err := row.Scan(&s.Id, &s.Title, &content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		}
		return Snippet{}, err
	}

	// Convert newlines to <br> tags
	s.Content = template.HTML(strings.Replace(content, "\n", "<br>", -1))
	return s, nil
}

func (m *SnippetModel) Latest() ([]Snippet, error) {
	query := `
		SELECT id, title, content, created, expires
		FROM snippets
		ORDER BY created DESC
		LIMIT 10`
	rows, err := m.DB.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []Snippet
	for rows.Next() {
		var snippet Snippet
		err := rows.Scan(&snippet.Id, &snippet.Title, &snippet.Content, &snippet.Created, &snippet.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, snippet)
	}
	return snippets, rows.Err()
}
