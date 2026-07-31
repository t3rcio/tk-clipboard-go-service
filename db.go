package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type HistoryItem struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type DB struct {
	conn *sql.DB
}

func initDB() (*DB, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".config", "clipsync")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dir, "history.db")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := conn.Exec(query); err != nil {
		return nil, err
	}

	return &DB{conn: conn}, nil
}

func (d *DB) Add(content string) error {
	if content == "" {
		return nil
	}	
	query := `
	INSERT INTO history (content, created_at) VALUES (?, CURRENT_TIMESTAMP)
	ON CONFLICT(content) DO UPDATE SET created_at = CURRENT_TIMESTAMP;`
	_, err := d.conn.Exec(query, content)
	return err
}

func (d *DB) List(limit int) ([]HistoryItem, error) {
	query := `SELECT id, content, created_at FROM history ORDER BY created_at DESC LIMIT ?`
	rows, err := d.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []HistoryItem
	for rows.Next() {
		var item HistoryItem
		if err := rows.Scan(&item.ID, &item.Content, &item.CreatedAt); err == nil {
			items = append(items, item)
		}
	}
	return items, nil
}

func (d *DB) Clear() error {
	_, err := d.conn.Exec("DELETE FROM history")
	return err
}

func (d *DB) Close() {
	if d.conn != nil {
		d.conn.Close()
	}
}