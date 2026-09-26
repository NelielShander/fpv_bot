package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const insertMessageQuery = `
	INSERT INTO messages (
		telegram_message_id,
		username,
		original_text,
		message_date
	)
	VALUES ($1, $2, $3, $4)
`

type MessageData struct {
	TelegramMessageID int
	Username          string
	OriginalText      string
	MessageDate       time.Time
}

type Database struct {
	pool *pgxpool.Pool
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
