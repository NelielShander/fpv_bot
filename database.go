package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
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
	getLstWeekReportsQuery = `
		SELECT username, original_text, message_date
		FROM messages
		WHERE message_date >= date_trunc('week', CURRENT_DATE) - INTERVAL '7 days'
		  AND message_date <  date_trunc('week', CURRENT_DATE)
		ORDER BY message_date;
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
	db.pool.Close()
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

func GetLastWeekReports(ctx context.Context, db *Database) ([]Report, error) {
	var reports []Report

	rows, err := db.pool.Query(ctx, getLstWeekReportsQuery)
	if err != nil {
		return reports, err
	}
	defer rows.Close()

	for rows.Next() {
		var report Report

		if err := rows.Scan(
			&report.Username,
			&report.MessageText,
			&report.MessageDate,
		); err != nil {
			return nil, err
		}

		reports = append(reports, report)
	}

	return reports, err
}
