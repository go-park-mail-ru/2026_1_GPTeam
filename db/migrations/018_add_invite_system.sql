ALTER TABLE account ADD COLUMN IF NOT EXISTS owner_id int;

ALTER TABLE account_user ADD COLUMN IF NOT EXISTS status text;
ALTER TABLE account_user ADD COLUMN IF NOT EXISTS created_at timestamp;

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
ALTER TABLE account_user ALTER COLUMN status SET NOT NULL;
ALTER TABLE account_user ALTER COLUMN status SET DEFAULT 'pending';
ALTER TABLE account_user ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE account_user ALTER COLUMN created_at SET DEFAULT now();

ALTER TABLE account ADD CONSTRAINT account_owner_fkey FOREIGN KEY (owner_id) REFERENCES "user"(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_account_user_status ON account_user(status);
CREATE INDEX IF NOT EXISTS idx_account_user_created_at ON account_user(created_at);
CREATE INDEX IF NOT EXISTS idx_account_owner_id ON account(owner_id);

---- create above / drop below ----

DROP INDEX IF EXISTS idx_account_owner_id;
DROP INDEX IF EXISTS idx_account_user_created_at;
DROP INDEX IF EXISTS idx_account_user_status;

ALTER TABLE account DROP CONSTRAINT IF EXISTS account_owner_fkey;

ALTER TABLE account ALTER COLUMN owner_id DROP NOT NULL;
ALTER TABLE account_user ALTER COLUMN status DROP NOT NULL;
ALTER TABLE account_user ALTER COLUMN status DROP DEFAULT;
ALTER TABLE account_user ALTER COLUMN created_at DROP NOT NULL;
ALTER TABLE account_user ALTER COLUMN created_at DROP DEFAULT;

ALTER TABLE account_user DROP COLUMN IF EXISTS status;
ALTER TABLE account_user DROP COLUMN IF EXISTS created_at;
ALTER TABLE account DROP COLUMN IF EXISTS owner_id;