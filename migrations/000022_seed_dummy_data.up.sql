-- =============================================================
-- Migration 000022: Seed comprehensive dummy data
-- =============================================================
-- Guru/siswa default password: "McBT@1234"  → hash $2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG
-- Admin (dari seed 000005): "Admin@123"

-- ─── Subjects ────────────────────────────────────────────────
INSERT INTO subjects (id, code, name, description) VALUES
    ('a1111111-1111-1111-1111-111111111111', 'MTK',  'Matematika',       'Matematika wajib untuk semua jenjang'),
    ('a2222222-2222-2222-2222-222222222222', 'BIN',  'Bahasa Indonesia', 'Bahasa Indonesia dan sastra'),
    ('a3333333-3333-3333-3333-333333333333', 'BIG',  'Bahasa Inggris',   'English language and literature'),
    ('a4444444-4444-4444-4444-444444444444', 'FIS',  'Fisika',           'Fisika untuk jenjang SMA'),
    ('a5555555-5555-5555-5555-555555555555', 'KIM',  'Kimia',            'Kimia untuk jenjang SMA'),
    ('a6666666-6666-6666-6666-666666666666', 'BIO',  'Biologi',          'Biologi untuk jenjang SMA');

-- ─── Academic Years ──────────────────────────────────────────
INSERT INTO academic_years (id, year, semester, start_date, end_date, is_active) VALUES
    ('b1111111-1111-1111-1111-111111111111', '2025/2026', 'SEMESTER', '2026-01-05', '2026-06-30', FALSE),
    ('b2222222-2222-2222-2222-222222222222', '2024/2025', 'ODD',      '2024-07-01', '2025-06-30', FALSE)
ON CONFLICT (year, semester) DO NOTHING;

-- ─── Teacher Users ───────────────────────────────────────────
WITH teacher_users AS (
    INSERT INTO users (id, username, name, email, password_hash, is_active) VALUES
        ('c1111111-1111-1111-1111-111111111111', 'guru_budi',  'Budi Santoso',  'budi@mcbt.local',  '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE),
        ('c2222222-2222-2222-2222-222222222222', 'guru_sari',  'Sari Dewi',     'sari@mcbt.local',  '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE),
        ('c3333333-3333-3333-3333-333333333333', 'guru_ahmad', 'Ahmad Hidayat', 'ahmad@mcbt.local', '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE),
        ('c4444444-4444-4444-4444-444444444444', 'guru_maya',  'Maya Putri',    'maya@mcbt.local',  '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE)
    RETURNING id
)
INSERT INTO user_roles (user_id, role_id)
SELECT tu.id, r.id
FROM teacher_users tu, roles r
WHERE r.name = 'teacher';

-- Teacher profiles
INSERT INTO teachers (id, user_id, nip, phone, address) VALUES
    ('d1111111-1111-1111-1111-111111111111', 'c1111111-1111-1111-1111-111111111111', '198501012010011001', '081234560001', 'Jl. Merdeka No. 1, Bandung'),
    ('d2222222-2222-2222-2222-222222222222', 'c2222222-2222-2222-2222-222222222222', '198703152010012002', '081234560002', 'Jl. Sudirman No. 25, Bandung'),
    ('d3333333-3333-3333-3333-333333333333', 'c3333333-3333-3333-3333-333333333333', '199005202012011003', '081234560003', 'Jl. Asia Afrika No. 10, Bandung'),
    ('d4444444-4444-4444-4444-444444444444', 'c4444444-4444-4444-4444-444444444444', '199208102015012004', '081234560004', 'Jl. Gatot Subroto No. 7, Bandung');

-- ─── Classes ─────────────────────────────────────────────────
WITH active_year AS (
    SELECT id AS academic_year_id FROM academic_years WHERE year = '2025/2026' AND semester = 'ODD' LIMIT 1
)
INSERT INTO classes (id, academic_year_id, homeroom_teacher_id, name, grade_level) VALUES
    ('e1111111-1111-1111-1111-111111111111', (SELECT academic_year_id FROM active_year), 'd1111111-1111-1111-1111-111111111111', 'X-A', 10),
    ('e2222222-2222-2222-2222-222222222222', (SELECT academic_year_id FROM active_year), 'd2222222-2222-2222-2222-222222222222', 'X-B', 10),
    ('e3333333-3333-3333-3333-333333333333', (SELECT academic_year_id FROM active_year), 'd3333333-3333-3333-3333-333333333333', 'XI-A', 11),
    ('e4444444-4444-4444-4444-444444444444', (SELECT academic_year_id FROM active_year), 'd4444444-4444-4444-4444-444444444444', 'XI-B', 11);

-- ─── Student Users ───────────────────────────────────────────
WITH student_users AS (
    INSERT INTO users (id, username, name, email, password_hash, is_active) VALUES
        ('f1111111-1111-1111-1111-111111111111', 'siswa_01', 'Andi Pratama',   'andi@mcbt.local',   '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE),
        ('f2222222-2222-2222-2222-222222222222', 'siswa_02', 'Bunga Citra',    'bunga@mcbt.local',  '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE),
        ('f3333333-3333-3333-3333-333333333333', 'siswa_03', 'Citra Anjani',   'citra@mcbt.local',  '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE),
        ('f4444444-4444-4444-4444-444444444444', 'siswa_04', 'Dimas Ardianto', 'dimas@mcbt.local',  '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE),
        ('f5555555-5555-5555-5555-555555555555', 'siswa_05', 'Eka Wulandari',  'eka@mcbt.local',    '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE),
        ('f6666666-6666-6666-6666-666666666666', 'siswa_06', 'Fajar Nugroho',  'fajar@mcbt.local',  '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE),
        ('f7777777-7777-7777-7777-777777777777', 'siswa_07', 'Gita Permata',   'gita@mcbt.local',   '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE),
        ('f8888888-8888-8888-8888-888888888888', 'siswa_08', 'Hendra Wijaya',  'hendra@mcbt.local', '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE),
        ('f9999999-9999-9999-9999-999999999999', 'siswa_09', 'Indah Sari',     'indah@mcbt.local',  '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE),
        ('faaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'siswa_10', 'Joko Susilo',    'joko@mcbt.local',   '$2a$10$iUTbGnLf6ntuym.4miWcsO4VK3o57lZMiqxQEVn5BPez47Gm9GaAG', TRUE)
    RETURNING id
)
INSERT INTO user_roles (user_id, role_id)
SELECT su.id, r.id
FROM student_users su, roles r
WHERE r.name = 'student';

-- Student profiles
INSERT INTO students (id, user_id, class_id, nis, phone, address) VALUES
    ('11111111-2222-3333-4444-555511111111', 'f1111111-1111-1111-1111-111111111111', 'e1111111-1111-1111-1111-111111111111', '20250001', '085610000001', 'Jl. Dago No. 1, Bandung'),
    ('11111111-2222-3333-4444-555522222222', 'f2222222-2222-2222-2222-222222222222', 'e1111111-1111-1111-1111-111111111111', '20250002', '085610000002', 'Jl. Buah Batu No. 3, Bandung'),
    ('11111111-2222-3333-4444-555533333333', 'f3333333-3333-3333-3333-333333333333', 'e1111111-1111-1111-1111-111111111111', '20250003', '085610000003', 'Jl. Setiabudhi No. 5, Bandung'),
    ('11111111-2222-3333-4444-555544444444', 'f4444444-4444-4444-4444-444444444444', 'e2222222-2222-2222-2222-222222222222', '20250004', '085610000004', 'Jl. Cendana No. 7, Bandung'),
    ('11111111-2222-3333-4444-555555555555', 'f5555555-5555-5555-5555-555555555555', 'e2222222-2222-2222-2222-222222222222', '20250005', '085610000005', 'Jl. Leumbeulang No. 9, Bandung'),
    ('11111111-2222-3333-4444-555566666666', 'f6666666-6666-6666-6666-666666666666', 'e2222222-2222-2222-2222-222222222222', '20250006', '085610000006', 'Jl. Karangarum No. 11, Bandung'),
    ('11111111-2222-3333-4444-555577777777', 'f7777777-7777-7777-7777-777777777777', 'e3333333-3333-3333-3333-333333333333', '20240007', '085610000007', 'Jl. Ciumbuleuit No. 13, Bandung'),
    ('11111111-2222-3333-4444-555588888888', 'f8888888-8888-8888-8888-888888888888', 'e3333333-3333-3333-3333-333333333333', '20240008', '085610000008', 'Jl. Sersan Bajuri No. 15, Bandung'),
    ('11111111-2222-3333-4444-555599999999', 'f9999999-9999-9999-9999-999999999999', 'e4444444-4444-4444-4444-444444444444', '20240009', '085610000009', 'Jl. Cigadung No. 17, Bandung'),
    ('11111111-2222-3333-4444-5555aaaaaaab', 'faaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'e4444444-4444-4444-4444-444444444444', '20240010', '085610000010', 'Jl. Ir. H. Djuanda No. 19, Bandung');

-- ─── Question Banks ──────────────────────────────────────────
INSERT INTO question_banks (id, subject_id, academic_year_id, created_by, title, code, description, status) VALUES
    ('11111111-aaaa-bbbb-cccc-dddddddd0001', 'a1111111-1111-1111-1111-111111111111', NULL, 'c1111111-1111-1111-1111-111111111111', 'Bank Soal Matematika Kelas X',  'QB-MTK-X',  'Soal matematika untuk kelas X',  'published'),
    ('11111111-aaaa-bbbb-cccc-dddddddd0002', 'a2222222-2222-2222-2222-222222222222', NULL, 'c2222222-2222-2222-2222-222222222222', 'Bank Soal B. Indonesia Kelas X', 'QB-BIN-X',  'Soal Bahasa Indonesia untuk kelas X', 'published'),
    ('11111111-aaaa-bbbb-cccc-dddddddd0003', 'a3333333-3333-3333-3333-333333333333', NULL, 'c3333333-3333-3333-3333-333333333333', 'Bank Soal B. Inggris Kelas X',   'QB-BIG-X',  'Soal Bahasa Inggris untuk kelas X', 'draft'),
    ('11111111-aaaa-bbbb-cccc-dddddddd0004', 'a1111111-1111-1111-1111-111111111111', NULL, 'c1111111-1111-1111-1111-111111111111', 'Bank Soal Matematika Kelas XI', 'QB-MTK-XI', 'Soal matematika untuk kelas XI', 'published');

-- ─── Questions: Matematika Kelas X ───────────────────────────
INSERT INTO questions (id, question_bank_id, question_type, content, score_weight, explanation, answer_keys) VALUES
    ('11111111-bbbb-cccc-dddd-eeeeeeee0001', '11111111-aaaa-bbbb-cccc-dddddddd0001', 'multiple_choice', 'Hasil dari 3² + 4² adalah...', 2.0, '3² = 9, 4² = 16, maka 9 + 16 = 25', 'B'),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0002', '11111111-aaaa-bbbb-cccc-dddddddd0001', 'multiple_choice', 'Jika x + 5 = 12, maka x = ...', 2.0, 'x = 12 - 5 = 7', 'C'),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0003', '11111111-aaaa-bbbb-cccc-dddddddd0001', 'multiple_choice', 'Luas lingkaran dengan jari-jari 7 cm adalah ... (π = 22/7)', 3.0, 'L = π × r² = 22/7 × 7² = 154 cm²', 'D'),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0004', '11111111-aaaa-bbbb-cccc-dddddddd0001', 'multiple_choice', 'Nilai dari √144 + √81 adalah ...', 2.0, '√144 = 12, √81 = 9, maka 12 + 9 = 21', 'A'),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0005', '11111111-aaaa-bbbb-cccc-dddddddd0001', 'true_false', 'Bilangan 17 termasuk bilangan prima.', 1.0, '17 hanya habis dibagi 1 dan 17', 'true');

-- Question options: Matematika
INSERT INTO question_options (question_id, label, content, is_correct, position) VALUES
    ('11111111-bbbb-cccc-dddd-eeeeeeee0001', 'A', '7',   FALSE, 0),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0001', 'B', '25',  TRUE,  1),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0001', 'C', '49',  FALSE, 2),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0001', 'D', '12',  FALSE, 3),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0002', 'A', '5',   FALSE, 0),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0002', 'B', '6',   FALSE, 1),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0002', 'C', '7',   TRUE,  2),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0002', 'D', '8',   FALSE, 3),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0003', 'A', '44 cm²',  FALSE, 0),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0003', 'B', '88 cm²',  FALSE, 1),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0003', 'C', '132 cm²', FALSE, 2),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0003', 'D', '154 cm²', TRUE,  3),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0003', 'E', '308 cm²', FALSE, 4),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0004', 'A', '21',  TRUE,  0),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0004', 'B', '22',  FALSE, 1),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0004', 'C', '23',  FALSE, 2),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0004', 'D', '24',  FALSE, 3),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0005', 'A', 'True',  TRUE,  0),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0005', 'B', 'False', FALSE, 1);

-- ─── Questions: Bahasa Indonesia Kelas X ─────────────────────
INSERT INTO questions (id, question_bank_id, question_type, content, score_weight, explanation, answer_keys) VALUES
    ('11111111-bbbb-cccc-dddd-eeeeeeee0006', '11111111-aaaa-bbbb-cccc-dddddddd0002', 'multiple_choice', 'Sinonim dari kata "gembira" adalah ...', 2.0, 'Gembira bermakna senang atau bahagia', 'B'),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0007', '11111111-aaaa-bbbb-cccc-dddddddd0002', 'multiple_choice', 'Antonim dari kata "sederhana" adalah ...', 2.0, 'Sederhana berarti simpel, antonimnya adalah rumit', 'C'),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0008', '11111111-aaaa-bbbb-cccc-dddddddd0002', 'essay', 'Jelaskan perbedaan antara teks eksposisi dan teks argumentasi!', 5.0, 'Teks eksposisi bertujuan memberikan informasi, teks argumentasi bertujuan membujuk', NULL);

INSERT INTO question_options (question_id, label, content, is_correct, position) VALUES
    ('11111111-bbbb-cccc-dddd-eeeeeeee0006', 'A', 'Sedih',   FALSE, 0),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0006', 'B', 'Senang',  TRUE,  1),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0006', 'C', 'Marah',   FALSE, 2),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0006', 'D', 'Takut',   FALSE, 3),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0007', 'A', 'Mudah',     FALSE, 0),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0007', 'B', 'Sederhana', FALSE, 1),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0007', 'C', 'Rumit',     TRUE,  2),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0007', 'D', 'Gampang',   FALSE, 3);

-- ─── Questions: Bahasa Inggris ───────────────────────────────
INSERT INTO questions (id, question_bank_id, question_type, content, score_weight, explanation, answer_keys) VALUES
    ('11111111-bbbb-cccc-dddd-eeeeeeee0009', '11111111-aaaa-bbbb-cccc-dddddddd0003', 'multiple_choice', 'What is the past tense of "go"?', 2.0, 'The past tense of "go" is "went"', 'C'),
    ('11111111-bbbb-cccc-dddd-eeeeeeee000a', '11111111-aaaa-bbbb-cccc-dddddddd0003', 'multiple_choice', '"She ___ a teacher." Choose the correct verb.', 2.0, '"She is a teacher" uses the verb "is"', 'B');

INSERT INTO question_options (question_id, label, content, is_correct, position) VALUES
    ('11111111-bbbb-cccc-dddd-eeeeeeee0009', 'A', 'Goed',  FALSE, 0),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0009', 'B', 'Gone',  FALSE, 1),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0009', 'C', 'Went',  TRUE,  2),
    ('11111111-bbbb-cccc-dddd-eeeeeeee0009', 'D', 'Going', FALSE, 3),
    ('11111111-bbbb-cccc-dddd-eeeeeeee000a', 'A', 'am',   FALSE, 0),
    ('11111111-bbbb-cccc-dddd-eeeeeeee000a', 'B', 'is',   TRUE,  1),
    ('11111111-bbbb-cccc-dddd-eeeeeeee000a', 'C', 'are',  FALSE, 2),
    ('11111111-bbbb-cccc-dddd-eeeeeeee000a', 'D', 'be',   FALSE, 3);

-- ─── Exams ───────────────────────────────────────────────────
WITH active_year AS (
    SELECT id AS academic_year_id FROM academic_years WHERE year = '2025/2026' AND semester = 'ODD' LIMIT 1
)
INSERT INTO exams (id, title, description, subject_id, academic_year_id, question_bank_id, created_by, status, duration_minutes, max_attempts, passing_grade, randomize_questions, randomize_options, allow_backtrack, auto_submit, show_result_immediately, negative_marking, negative_value, token_enabled, exam_token, results_published, allow_discussion) VALUES
    ('11111111-cccc-dddd-eeee-ffffffff0001', 'UTS Matematika Kelas X',       'Ujian Tengah Semester Matematika',        'a1111111-1111-1111-1111-111111111111', (SELECT academic_year_id FROM active_year), '11111111-aaaa-bbbb-cccc-dddddddd0001', 'c1111111-1111-1111-1111-111111111111', 'published', 90,  2, 70.00, TRUE,  TRUE,  TRUE, TRUE, FALSE, FALSE, 0.00, FALSE, NULL, FALSE, FALSE),
    ('11111111-cccc-dddd-eeee-ffffffff0002', 'UTS Bahasa Indonesia Kelas X', 'Ujian Tengah Semester Bahasa Indonesia', 'a2222222-2222-2222-2222-222222222222', (SELECT academic_year_id FROM active_year), '11111111-aaaa-bbbb-cccc-dddddddd0002', 'c2222222-2222-2222-2222-222222222222', 'published', 60,  1, 75.00, FALSE, FALSE, TRUE, TRUE, TRUE,  FALSE, 0.00, TRUE,  'UTS2026', FALSE, FALSE),
    ('11111111-cccc-dddd-eeee-ffffffff0003', 'Quiz Matematika Kelas XI',     'Quiz bab Aljabar',                       'a1111111-1111-1111-1111-111111111111', (SELECT academic_year_id FROM active_year), '11111111-aaaa-bbbb-cccc-dddddddd0004', 'c1111111-1111-1111-1111-111111111111', 'draft',    30,  1, 60.00, FALSE, FALSE, TRUE, TRUE, FALSE, FALSE, 0.00, FALSE, NULL, FALSE, FALSE);

-- ─── Exam Sections ───────────────────────────────────────────
INSERT INTO exam_sections (id, exam_id, name, sequence) VALUES
    ('11111111-dddd-eeee-ffff-aaaaaaaa0001', '11111111-cccc-dddd-eeee-ffffffff0001', 'Pilihan Ganda', 1),
    ('11111111-dddd-eeee-ffff-aaaaaaaa0002', '11111111-cccc-dddd-eeee-ffffffff0001', 'Benar / Salah', 2),
    ('11111111-dddd-eeee-ffff-aaaaaaaa0003', '11111111-cccc-dddd-eeee-ffffffff0002', 'Pilihan Ganda', 1),
    ('11111111-dddd-eeee-ffff-aaaaaaaa0004', '11111111-cccc-dddd-eeee-ffffffff0002', 'Esai',          2);

-- ─── Section ↔ Questions ─────────────────────────────────────
INSERT INTO exam_section_questions (section_id, question_id) VALUES
    ('11111111-dddd-eeee-ffff-aaaaaaaa0001', '11111111-bbbb-cccc-dddd-eeeeeeee0001'),
    ('11111111-dddd-eeee-ffff-aaaaaaaa0001', '11111111-bbbb-cccc-dddd-eeeeeeee0002'),
    ('11111111-dddd-eeee-ffff-aaaaaaaa0001', '11111111-bbbb-cccc-dddd-eeeeeeee0003'),
    ('11111111-dddd-eeee-ffff-aaaaaaaa0001', '11111111-bbbb-cccc-dddd-eeeeeeee0004'),
    ('11111111-dddd-eeee-ffff-aaaaaaaa0002', '11111111-bbbb-cccc-dddd-eeeeeeee0005'),
    ('11111111-dddd-eeee-ffff-aaaaaaaa0003', '11111111-bbbb-cccc-dddd-eeeeeeee0006'),
    ('11111111-dddd-eeee-ffff-aaaaaaaa0003', '11111111-bbbb-cccc-dddd-eeeeeeee0007'),
    ('11111111-dddd-eeee-ffff-aaaaaaaa0004', '11111111-bbbb-cccc-dddd-eeeeeeee0008');

-- ─── Exam Schedules ──────────────────────────────────────────
INSERT INTO exam_schedules (id, exam_id, start_time, end_time, token) VALUES
    ('11111111-eeee-ffff-aaaa-bbbbbbbbb001', '11111111-cccc-dddd-eeee-ffffffff0001', '2026-03-10 08:00:00+07', '2026-03-10 09:30:00+07', 'MTK2026'),
    ('11111111-eeee-ffff-aaaa-bbbbbbbbb002', '11111111-cccc-dddd-eeee-ffffffff0002', '2026-03-11 10:00:00+07', '2026-03-11 11:00:00+07', 'BIN2026');

-- ─── Exam Participants ───────────────────────────────────────
INSERT INTO exam_participants (exam_id, student_id, assigned_via) VALUES
    ('11111111-cccc-dddd-eeee-ffffffff0001', '11111111-2222-3333-4444-555511111111', 'class'),
    ('11111111-cccc-dddd-eeee-ffffffff0001', '11111111-2222-3333-4444-555522222222', 'class'),
    ('11111111-cccc-dddd-eeee-ffffffff0001', '11111111-2222-3333-4444-555533333333', 'class'),
    ('11111111-cccc-dddd-eeee-ffffffff0001', '11111111-2222-3333-4444-555544444444', 'class'),
    ('11111111-cccc-dddd-eeee-ffffffff0001', '11111111-2222-3333-4444-555555555555', 'class'),
    ('11111111-cccc-dddd-eeee-ffffffff0001', '11111111-2222-3333-4444-555566666666', 'class'),
    ('11111111-cccc-dddd-eeee-ffffffff0002', '11111111-2222-3333-4444-555511111111', 'class'),
    ('11111111-cccc-dddd-eeee-ffffffff0002', '11111111-2222-3333-4444-555522222222', 'class'),
    ('11111111-cccc-dddd-eeee-ffffffff0002', '11111111-2222-3333-4444-555533333333', 'class'),
    ('11111111-cccc-dddd-eeee-ffffffff0002', '11111111-2222-3333-4444-555544444444', 'class'),
    ('11111111-cccc-dddd-eeee-ffffffff0002', '11111111-2222-3333-4444-555555555555', 'class'),
    ('11111111-cccc-dddd-eeee-ffffffff0002', '11111111-2222-3333-4444-555566666666', 'class');

-- ─── Exam Attempts ───────────────────────────────────────────
-- UTS Matematika: 6 siswa (X-A + X-B), beragam skor
INSERT INTO exam_attempts (id, exam_id, student_id, attempt_no, status, started_at, expires_at, submitted_at, score) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0001', '11111111-cccc-dddd-eeee-ffffffff0001', '11111111-2222-3333-4444-555511111111', 1, 'submitted',   '2026-03-10 08:05:00+07', '2026-03-10 09:35:00+07', '2026-03-10 08:45:00+07', 85.00),
    ('11111111-ffff-aaaa-bbbb-cccccccc0002', '11111111-cccc-dddd-eeee-ffffffff0001', '11111111-2222-3333-4444-555522222222', 1, 'submitted',   '2026-03-10 08:10:00+07', '2026-03-10 09:40:00+07', '2026-03-10 09:20:00+07', 72.50),
    ('11111111-ffff-aaaa-bbbb-cccccccc0003', '11111111-cccc-dddd-eeee-ffffffff0001', '11111111-2222-3333-4444-555533333333', 1, 'submitted',   '2026-03-10 08:15:00+07', '2026-03-10 09:45:00+07', '2026-03-10 09:10:00+07', 58.00),
    ('11111111-ffff-aaaa-bbbb-cccccccc0005', '11111111-cccc-dddd-eeee-ffffffff0001', '11111111-2222-3333-4444-555544444444', 1, 'submitted',   '2026-03-10 08:02:00+07', '2026-03-10 09:32:00+07', '2026-03-10 08:55:00+07', 90.00),
    ('11111111-ffff-aaaa-bbbb-cccccccc0006', '11111111-cccc-dddd-eeee-ffffffff0001', '11111111-2222-3333-4444-555555555555', 1, 'submitted',   '2026-03-10 08:08:00+07', '2026-03-10 09:38:00+07', '2026-03-10 09:35:00+07', 65.00),
    ('11111111-ffff-aaaa-bbbb-cccccccc0007', '11111111-cccc-dddd-eeee-ffffffff0001', '11111111-2222-3333-4444-555566666666', 1, 'submitted',   '2026-03-10 08:12:00+07', '2026-03-10 09:42:00+07', '2026-03-10 09:05:00+07', 78.00);

-- UTS Bahasa Indonesia: 6 siswa (X-A + X-B), beragam skor
INSERT INTO exam_attempts (id, exam_id, student_id, attempt_no, status, started_at, expires_at, submitted_at, score) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0004', '11111111-cccc-dddd-eeee-ffffffff0002', '11111111-2222-3333-4444-555511111111', 1, 'submitted',   '2026-03-11 10:02:00+07', '2026-03-11 11:02:00+07', '2026-03-11 10:40:00+07', 80.00),
    ('11111111-ffff-aaaa-bbbb-cccccccc0008', '11111111-cccc-dddd-eeee-ffffffff0002', '11111111-2222-3333-4444-555522222222', 1, 'submitted',   '2026-03-11 10:05:00+07', '2026-03-11 11:05:00+07', '2026-03-11 10:55:00+07', 88.00),
    ('11111111-ffff-aaaa-bbbb-cccccccc0009', '11111111-cccc-dddd-eeee-ffffffff0002', '11111111-2222-3333-4444-555533333333', 1, 'submitted',   '2026-03-11 10:03:00+07', '2026-03-11 11:03:00+07', '2026-03-11 10:30:00+07', 70.00),
    ('11111111-ffff-aaaa-bbbb-cccccccc0010', '11111111-cccc-dddd-eeee-ffffffff0002', '11111111-2222-3333-4444-555544444444', 1, 'submitted',   '2026-03-11 10:01:00+07', '2026-03-11 11:01:00+07', '2026-03-11 10:48:00+07', 62.00),
    ('11111111-ffff-aaaa-bbbb-cccccccc0011', '11111111-cccc-dddd-eeee-ffffffff0002', '11111111-2222-3333-4444-555555555555', 1, 'submitted',   '2026-03-11 10:04:00+07', '2026-03-11 11:04:00+07', '2026-03-11 10:25:00+07', 50.00),
    ('11111111-ffff-aaaa-bbbb-cccccccc0012', '11111111-cccc-dddd-eeee-ffffffff0002', '11111111-2222-3333-4444-555566666666', 1, 'submitted',   '2026-03-11 10:06:00+07', '2026-03-11 11:06:00+07', '2026-03-11 10:50:00+07', 75.00);

-- ─── Exam Answers ────────────────────────────────────────────
-- Attempt 0001 (Andi - MTK): answered all 5 → score 85
INSERT INTO exam_answers (attempt_id, question_id, answer_value, is_flagged, answered_at, score, is_correct) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0001', '11111111-bbbb-cccc-dddd-eeeeeeee0001', 'B', FALSE, '2026-03-10 08:10:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0001', '11111111-bbbb-cccc-dddd-eeeeeeee0002', 'C', FALSE, '2026-03-10 08:15:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0001', '11111111-bbbb-cccc-dddd-eeeeeeee0003', 'D', TRUE,  '2026-03-10 08:25:00+07', 3.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0001', '11111111-bbbb-cccc-dddd-eeeeeeee0004', 'A', FALSE, '2026-03-10 08:35:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0001', '11111111-bbbb-cccc-dddd-eeeeeeee0005', 'A', FALSE, '2026-03-10 08:40:00+07', 1.00, TRUE);

-- Attempt 0002 (Bunga - MTK): answered 4/5, 2 wrong → score 72.5
INSERT INTO exam_answers (attempt_id, question_id, answer_value, is_flagged, answered_at, score, is_correct) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0002', '11111111-bbbb-cccc-dddd-eeeeeeee0001', 'B', FALSE, '2026-03-10 08:20:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0002', '11111111-bbbb-cccc-dddd-eeeeeeee0002', 'A', FALSE, '2026-03-10 08:30:00+07', 0.00, FALSE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0002', '11111111-bbbb-cccc-dddd-eeeeeeee0003', 'D', FALSE, '2026-03-10 08:45:00+07', 3.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0002', '11111111-bbbb-cccc-dddd-eeeeeeee0004', 'B', FALSE, '2026-03-10 09:00:00+07', 0.00, FALSE);

-- Attempt 0003 (Citra - MTK): answered 4/5, 3 wrong → score 58
INSERT INTO exam_answers (attempt_id, question_id, answer_value, is_flagged, answered_at, score, is_correct) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0003', '11111111-bbbb-cccc-dddd-eeeeeeee0001', 'C', FALSE, '2026-03-10 08:22:00+07', 0.00, FALSE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0003', '11111111-bbbb-cccc-dddd-eeeeeeee0002', 'C', FALSE, '2026-03-10 08:35:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0003', '11111111-bbbb-cccc-dddd-eeeeeeee0004', 'C', FALSE, '2026-03-10 08:50:00+07', 0.00, FALSE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0003', '11111111-bbbb-cccc-dddd-eeeeeeee0005', 'B', FALSE, '2026-03-10 09:05:00+07', 0.00, FALSE);

-- Attempt 0005 (Dimas - MTK): all correct → score 90
INSERT INTO exam_answers (attempt_id, question_id, answer_value, is_flagged, answered_at, score, is_correct) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0005', '11111111-bbbb-cccc-dddd-eeeeeeee0001', 'B', FALSE, '2026-03-10 08:08:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0005', '11111111-bbbb-cccc-dddd-eeeeeeee0002', 'C', FALSE, '2026-03-10 08:18:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0005', '11111111-bbbb-cccc-dddd-eeeeeeee0003', 'D', FALSE, '2026-03-10 08:28:00+07', 3.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0005', '11111111-bbbb-cccc-dddd-eeeeeeee0004', 'A', FALSE, '2026-03-10 08:38:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0005', '11111111-bbbb-cccc-dddd-eeeeeeee0005', 'A', FALSE, '2026-03-10 08:48:00+07', 1.00, TRUE);

-- Attempt 0006 (Eka - MTK): 3/5 correct → score 65
INSERT INTO exam_answers (attempt_id, question_id, answer_value, is_flagged, answered_at, score, is_correct) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0006', '11111111-bbbb-cccc-dddd-eeeeeeee0001', 'A', FALSE, '2026-03-10 08:15:00+07', 0.00, FALSE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0006', '11111111-bbbb-cccc-dddd-eeeeeeee0002', 'C', FALSE, '2026-03-10 08:25:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0006', '11111111-bbbb-cccc-dddd-eeeeeeee0003', 'C', FALSE, '2026-03-10 08:40:00+07', 0.00, FALSE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0006', '11111111-bbbb-cccc-dddd-eeeeeeee0004', 'A', FALSE, '2026-03-10 09:00:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0006', '11111111-bbbb-cccc-dddd-eeeeeeee0005', 'A', FALSE, '2026-03-10 09:20:00+07', 1.00, TRUE);

-- Attempt 0007 (Fajar - MTK): 4/5 correct → score 78
INSERT INTO exam_answers (attempt_id, question_id, answer_value, is_flagged, answered_at, score, is_correct) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0007', '11111111-bbbb-cccc-dddd-eeeeeeee0001', 'B', FALSE, '2026-03-10 08:18:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0007', '11111111-bbbb-cccc-dddd-eeeeeeee0002', 'D', FALSE, '2026-03-10 08:28:00+07', 0.00, FALSE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0007', '11111111-bbbb-cccc-dddd-eeeeeeee0003', 'D', FALSE, '2026-03-10 08:42:00+07', 3.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0007', '11111111-bbbb-cccc-dddd-eeeeeeee0004', 'A', FALSE, '2026-03-10 08:52:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0007', '11111111-bbbb-cccc-dddd-eeeeeeee0005', 'A', FALSE, '2026-03-10 09:00:00+07', 1.00, TRUE);

-- Attempt 0004 (Andi - BIN): answered 2 MC + 1 essay → score 80
INSERT INTO exam_answers (attempt_id, question_id, answer_value, is_flagged, answered_at, score, is_correct) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0004', '11111111-bbbb-cccc-dddd-eeeeeeee0006', 'B', FALSE, '2026-03-11 10:10:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0004', '11111111-bbbb-cccc-dddd-eeeeeeee0007', 'C', FALSE, '2026-03-11 10:20:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0004', '11111111-bbbb-cccc-dddd-eeeeeeee0008', 'Teks eksposisi memberikan informasi...', FALSE, '2026-03-11 10:35:00+07', NULL, NULL);

-- Attempt 0008 (Bunga - BIN): all MC correct → score 88
INSERT INTO exam_answers (attempt_id, question_id, answer_value, is_flagged, answered_at, score, is_correct) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0008', '11111111-bbbb-cccc-dddd-eeeeeeee0006', 'B', FALSE, '2026-03-11 10:12:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0008', '11111111-bbbb-cccc-dddd-eeeeeeee0007', 'C', FALSE, '2026-03-11 10:22:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0008', '11111111-bbbb-cccc-dddd-eeeeeeee0008', 'Teks eksposisi bertujuan informasi, argumentasi membujuk', FALSE, '2026-03-11 10:45:00+07', NULL, NULL);

-- Attempt 0009 (Citra - BIN): 1 MC wrong + essay → score 70
INSERT INTO exam_answers (attempt_id, question_id, answer_value, is_flagged, answered_at, score, is_correct) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0009', '11111111-bbbb-cccc-dddd-eeeeeeee0006', 'A', FALSE, '2026-03-11 10:08:00+07', 0.00, FALSE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0009', '11111111-bbbb-cccc-dddd-eeeeeeee0007', 'C', FALSE, '2026-03-11 10:18:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0009', '11111111-bbbb-cccc-dddd-eeeeeeee0008', 'Eksposisi informatif, argumentasi persuasif', FALSE, '2026-03-11 10:28:00+07', NULL, NULL);

-- Attempt 0010 (Dimas - BIN): 2 MC correct → score 62
INSERT INTO exam_answers (attempt_id, question_id, answer_value, is_flagged, answered_at, score, is_correct) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0010', '11111111-bbbb-cccc-dddd-eeeeeeee0006', 'B', FALSE, '2026-03-11 10:06:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0010', '11111111-bbbb-cccc-dddd-eeeeeeee0007', 'B', FALSE, '2026-03-11 10:16:00+07', 0.00, FALSE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0010', '11111111-bbbb-cccc-dddd-eeeeeeee0008', 'Eksposisi dan argumentasi berbeda tujuan', FALSE, '2026-03-11 10:40:00+07', NULL, NULL);

-- Attempt 0011 (Eka - BIN): 1 MC correct → score 50
INSERT INTO exam_answers (attempt_id, question_id, answer_value, is_flagged, answered_at, score, is_correct) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0011', '11111111-bbbb-cccc-dddd-eeeeeeee0006', 'C', FALSE, '2026-03-11 10:10:00+07', 0.00, FALSE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0011', '11111111-bbbb-cccc-dddd-eeeeeeee0007', 'C', FALSE, '2026-03-11 10:15:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0011', '11111111-bbbb-cccc-dddd-eeeeeeee0008', 'Eksposisi menjelaskan, argumentasi membujuk', FALSE, '2026-03-11 10:20:00+07', NULL, NULL);

-- Attempt 0012 (Fajar - BIN): 2 MC correct → score 75
INSERT INTO exam_answers (attempt_id, question_id, answer_value, is_flagged, answered_at, score, is_correct) VALUES
    ('11111111-ffff-aaaa-bbbb-cccccccc0012', '11111111-bbbb-cccc-dddd-eeeeeeee0006', 'B', FALSE, '2026-03-11 10:10:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0012', '11111111-bbbb-cccc-dddd-eeeeeeee0007', 'C', FALSE, '2026-03-11 10:25:00+07', 2.00, TRUE),
    ('11111111-ffff-aaaa-bbbb-cccccccc0012', '11111111-bbbb-cccc-dddd-eeeeeeee0008', 'Eksposisi untuk informasi, argumentasi untuk ajakan', FALSE, '2026-03-11 10:42:00+07', NULL, NULL);
