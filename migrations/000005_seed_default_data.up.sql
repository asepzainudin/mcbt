-- =============================================================
-- Migration 000005: Seed default data
-- =============================================================

INSERT INTO roles (name, description) VALUES
    ('admin',   'Full system administrator'),
    ('teacher', 'Teacher / instructor'),
    ('student', 'Student');

INSERT INTO permissions (name, description) VALUES
    ('users.manage',       'Manage users, roles and permissions'),
    ('master-data.manage', 'Manage academic years, classes and subjects'),
    ('questions.manage',   'Manage question banks and questions'),
    ('exams.take',         'Take exams'),
    ('exams.grade',        'Grade and review exams');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN (
    'users.manage', 'master-data.manage', 'questions.manage',
    'exams.take', 'exams.grade'
)
WHERE r.name = 'admin';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN ('questions.manage', 'exams.grade')
WHERE r.name = 'teacher';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name = 'exams.take'
WHERE r.name = 'student';

INSERT INTO users (role_id, name, email, password_hash)
SELECT r.id,
       'System Administrator',
       'admin@mcbt.local',
       '$2a$10$1ChYi2zzCO.tsbQrqpY5Ku4N0.dT6LeZ8e9wz/GzFFjB4JiaB3fOm'
FROM roles r
WHERE r.name = 'admin';

INSERT INTO academic_years (name, start_date, end_date, is_active)
VALUES ('2025/2026', DATE '2025-07-01', DATE '2026-06-30', TRUE);
