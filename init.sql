CREATE TABLE IF NOT EXISTS public.records
(
    id         BIGSERIAL PRIMARY KEY,
    type       TEXT,
    number     TEXT,
    date       DATE,
    fpvName    TEXT NOT NULL,
    Count      INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.messages
(
    id                  BIGSERIAL PRIMARY KEY,
    telegram_message_id BIGINT    NOT NULL,
    username            TEXT,
    original_text       TEXT      NOT NULL,
    message_date        TIMESTAMP NOT NULL,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS messages_telegram_message_id_uidx
    ON messages (telegram_message_id);