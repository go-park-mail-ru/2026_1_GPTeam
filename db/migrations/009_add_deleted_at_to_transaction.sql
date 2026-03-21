ALTER TABLE transaction ADD COLUMN IF NOT EXISTS deleted_at timestamp;

---- create above / drop below ----

ALTER TABLE transaction DROP COLUMN IF EXISTS deleted_at;