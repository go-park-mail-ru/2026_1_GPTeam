ALTER TABLE transaction ADD COLUMN IF NOT EXISTS deleted_at timestamp constraint deleted_at_not_in_past check ( deleted_at is null or deleted_at >= created_at );

---- create above / drop below ----

ALTER TABLE transaction DROP COLUMN IF EXISTS deleted_at;