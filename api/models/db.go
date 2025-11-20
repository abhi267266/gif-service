package models

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

type Video struct {
	ID           string    `db:"id" json:"id"`
	Filename     string    `db:"filename" json:"filename"`
	OriginalName string    `db:"original_name" json:"original_name"`
	Size         int64     `db:"size" json:"size"`
	Bucket       string    `db:"bucket" json:"bucket"`
	URL          string    `db:"url" json:"url"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to database")
	return createTables()
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS videos (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		filename TEXT NOT NULL,
		original_name TEXT NOT NULL,
		size BIGINT NOT NULL,
		bucket TEXT NOT NULL,
		url TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	log.Println("Tables created or already exist")
	return nil
}

func CreateVideo(video *Video) error {
	query := `
	INSERT INTO videos (filename, original_name, size, bucket, url)
	VALUES (:filename, :original_name, :size, :bucket, :url)
	RETURNING id, created_at`

	rows, err := DB.NamedQuery(query, video)
	if err != nil {
		return fmt.Errorf("failed to insert video: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&video.ID, &video.CreatedAt); err != nil {
			return fmt.Errorf("failed to scan video id: %w", err)
		}
	}

	return nil
}
