ALTER TABLE transaction ADD COLUMN deleted_at timestamp;

---- create above / drop below ----

ALTER TABLE transaction DROP COLUMN deleted_at;
