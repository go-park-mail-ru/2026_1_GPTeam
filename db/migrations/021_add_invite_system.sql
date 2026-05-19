ALTER TABLE account ADD COLUMN IF NOT EXISTS owner_id int;

ALTER TABLE account_user ADD COLUMN IF NOT EXISTS status text;
ALTER TABLE account_user ADD COLUMN IF NOT EXISTS created_at timestamp;
ALTER TABLE account_user ADD COLUMN IF NOT EXISTS deleted_at timestamp;
ALTER TABLE account_user ADD COLUMN IF NOT EXISTS deleted_reason text
    CHECK (deleted_reason IN ('kicked', 'left'));

UPDATE account a
SET owner_id = (
    SELECT au.user_id
    FROM account_user au
    WHERE au.account_id = a.id
    LIMIT 1
)
WHERE a.owner_id IS NULL;

UPDATE account_user
SET status = 'accepted',
    created_at = COALESCE(created_at, now())
WHERE status IS NULL;

ALTER TABLE account ALTER COLUMN owner_id SET NOT NULL;
ALTER TABLE account_user ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE account_user ALTER COLUMN created_at SET DEFAULT now();

ALTER TABLE account DROP CONSTRAINT IF EXISTS account_owner_fkey;
ALTER TABLE account ADD CONSTRAINT account_owner_fkey
    FOREIGN KEY (owner_id) REFERENCES "user"(id) ON DELETE CASCADE;

CREATE TYPE account_user_status AS ENUM ('pending', 'accepted', 'declined');
ALTER TABLE account_user
    ALTER COLUMN status TYPE account_user_status USING status::account_user_status,
    ALTER COLUMN status SET NOT NULL,
    ALTER COLUMN status SET DEFAULT 'pending'::account_user_status;

---- create above / drop below ----

ALTER TABLE account DROP CONSTRAINT IF EXISTS account_owner_fkey;

ALTER TABLE account_user ALTER COLUMN status DROP DEFAULT;
ALTER TABLE account_user ALTER COLUMN status TYPE text USING status::text;
DROP TYPE IF EXISTS account_user_status;

ALTER TABLE account_user DROP COLUMN IF EXISTS deleted_reason;
ALTER TABLE account_user DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE account_user DROP COLUMN IF EXISTS status;
ALTER TABLE account_user DROP COLUMN IF EXISTS created_at;
ALTER TABLE account DROP COLUMN IF EXISTS owner_id;