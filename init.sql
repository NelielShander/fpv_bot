ALTER DATABASE fpv_lists SET timezone TO 'Europe/Moscow';

CREATE TABLE IF NOT EXISTS public.messages
(
    id                  BIGSERIAL PRIMARY KEY,
    telegram_message_id BIGINT    NOT NULL,
    username            TEXT,
    original_text       TEXT      NOT NULL,
    message_date        TIMESTAMP NOT NULL,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);