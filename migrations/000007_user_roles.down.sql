ALTER TABLE users ADD COLUMN role_id UUID REFERENCES roles (id);

UPDATE users u
SET role_id = ur.role_id
FROM (
    SELECT DISTINCT ON (user_id) user_id, role_id
    FROM user_roles
    ORDER BY user_id, created_at ASC
) ur
WHERE ur.user_id = u.id;

CREATE INDEX idx_users_role_id ON users (role_id);

DROP TABLE IF EXISTS user_roles;
