DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns 
    WHERE table_name = 'account' AND column_name = 'owner_id' AND column_default IS NOT NULL
  ) THEN
    ALTER TABLE account ALTER COLUMN owner_id DROP DEFAULT;
  END IF;
END $$;

ALTER TABLE account DROP CONSTRAINT IF EXISTS account_owner_fkey;
ALTER TABLE account ADD CONSTRAINT account_owner_fkey 
  FOREIGN KEY (owner_id) REFERENCES "user"(id) ON DELETE CASCADE;

---- create above / drop below ----

ALTER TABLE account DROP CONSTRAINT IF EXISTS account_owner_fkey;
ALTER TABLE account ALTER COLUMN owner_id SET DEFAULT 0;
ALTER TABLE account ADD CONSTRAINT account_owner_fkey 
  FOREIGN KEY (owner_id) REFERENCES "user"(id) ON DELETE SET NULL;