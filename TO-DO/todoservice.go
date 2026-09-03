package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
	_ "modernc.org/sqlite"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type TodoStats struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Active    int `json:"active"`
}

type TodoService struct {
	db *sql.DB
}

func NewTodoService() *TodoService {
	return &TodoService{}
}

// ServiceStartup opens (creating on first run) the SQLite database.
// Lowercase 'up' — same lifecycle hook spelling as the QR project.
func (t *TodoService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(configDir, "TO-DO")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	db, err := sql.Open("sqlite", filepath.Join(dir, "todos.db"))
	if err != nil {
		return err
	}
	t.db = db

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			title     TEXT NOT NULL,
			completed INTEGER NOT NULL DEFAULT 0
		)`)
	return err
}

// ServiceShutdown closes the database when the app exits.
func (t *TodoService) ServiceShutdown() error {
	if t.db != nil {
		return t.db.Close()
	}
	return nil
}

func (t *TodoService) GetStats() (TodoStats, error) {
	var s TodoStats
	err := t.db.QueryRow(`
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE completed = 1),
			COUNT(*) FILTER (WHERE completed = 0)
		FROM todos`).Scan(&s.Total, &s.Completed, &s.Active)
	return s, err
}

func (t *TodoService) GetAll() ([]Todo, error) {
	rows, err := t.db.Query(`SELECT id, title, completed FROM todos ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, rows.Err()
}

func (t *TodoService) Add(title string) (*Todo, error) {
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	res, err := t.db.Exec(`INSERT INTO todos (title) VALUES (?)`, title)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Todo{ID: int(id), Title: title, Completed: false}, nil
}

func (t *TodoService) Toggle(id int) error {
	res, err := t.db.Exec(`UPDATE todos SET completed = NOT completed WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("todo not found")
	}
	return nil
}

func (t *TodoService) Delete(id int) error {
	res, err := t.db.Exec(`DELETE FROM todos WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("todo not found")
	}
	return nil
}

func (t *TodoService) ClearCompleted() (int, error) {
	res, err := t.db.Exec(`DELETE FROM todos WHERE completed = 1`)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	return int(n), err
}

func (t *TodoService) GetFiltered(filter string) ([]Todo, error) {
	query := `SELECT id, title, completed FROM todos`
	switch filter {
	case "active":
		query += ` WHERE completed = 0`
	case "completed":
		query += ` WHERE completed = 1`
	}
	query += ` ORDER BY id`

	rows, err := t.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, rows.Err()
}

func (t *TodoService) Update(id int, title string) error {
	if title == "" {
		return errors.New("title cannot be empty")
	}
	res, err := t.db.Exec(`UPDATE todos SET title = ? WHERE id = ?`, title, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("todo not found")
	}
	return nil
}
