ALTER TABLE users ADD COLUMN username VARCHAR(50);

UPDATE users
SET username = CASE
    WHEN email = 'admin@mcbt.local' THEN 'admin123'
    ELSE split_part(email, '@', 1) || '_' || left(id::text, 4)
END;

ALTER TABLE users ALTER COLUMN username SET NOT NULL;

CREATE UNIQUE INDEX uq_users_username ON users (username);

ALTER TABLE users ADD COLUMN token_version INTEGER NOT NULL DEFAULT 1;
