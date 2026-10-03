-- Отчет с 14 позавчера по 14 вчера
SELECT *
FROM fpv_lists.public.messages
WHERE message_date >= (
    (CURRENT_TIMESTAMP AT TIME ZONE 'Europe/Moscow')::date
        - INTERVAL '2 days'
        + INTERVAL '14 hours'
    )
  AND message_date < (
    (CURRENT_TIMESTAMP AT TIME ZONE 'Europe/Moscow')::date
        - INTERVAL '1 day'
        + INTERVAL '14 hours'
    )
ORDER BY message_date;