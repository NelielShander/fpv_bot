package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	insertMessageQuery = `
		INSERT INTO messages (
			telegram_message_id,
			username,
			original_text,
			message_date
		)
		VALUES ($1, $2, $3, $4)
	`

	getLastWeekReportsQuery = `
		SELECT
			username,
			original_text,
			message_date
		FROM messages
		WHERE message_date >= date_trunc('week', CURRENT_DATE) - INTERVAL '7 days'
		  AND message_date < date_trunc('week', CURRENT_DATE)
		ORDER BY message_date
	`
)

type Database struct {
	pool *pgxpool.Pool
}

type Report struct {
	Username    string
	MessageText string
	MessageDate time.Time
}

func NewDatabase(
	ctx context.Context,
	url string,
) (*Database, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &Database{
		pool: pool,
	}, nil
}

func (db *Database) Close() {
	if db != nil && db.pool != nil {
		db.pool.Close()
	}
}

func (db *Database) SaveMessage(
	ctx context.Context,
	data MessageData,
) error {
	_, err := db.pool.Exec(
		ctx,
		insertMessageQuery,
		data.TelegramMessageID,
		data.Username,
		data.OriginalText,
		data.MessageDate,
	)

	return err
}

func GetLastWeekReports(
	ctx context.Context,
	db *Database,
) ([]Report, error) {
	rows, err := db.pool.Query(
		ctx,
		getLastWeekReportsQuery,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reports := make([]Report, 0)

	for rows.Next() {
		var report Report

		err := rows.Scan(
			&report.Username,
			&report.MessageText,
			&report.MessageDate,
		)
		if err != nil {
			return nil, err
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}
