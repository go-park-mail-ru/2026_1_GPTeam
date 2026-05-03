-- db/migrations/013_add_deleted_at_to_account.sql

ALTER TABLE account
ADD COLUMN IF NOT EXISTS deleted_at timestamp
CONSTRAINT account_deleted_at_not_in_past
CHECK (deleted_at IS NULL OR deleted_at >= created_at);

---- create above / drop below ----

ALTER TABLE account DROP COLUMN IF EXISTS deleted_at;