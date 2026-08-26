-- =============================================================
-- Migration 000022: Revert dummy seed data
-- =============================================================

DELETE FROM exam_answers WHERE attempt_id IN (
    '11111111-ffff-aaaa-bbbb-cccccccc0001', '11111111-ffff-aaaa-bbbb-cccccccc0002',
    '11111111-ffff-aaaa-bbbb-cccccccc0003', '11111111-ffff-aaaa-bbbb-cccccccc0004',
    '11111111-ffff-aaaa-bbbb-cccccccc0005', '11111111-ffff-aaaa-bbbb-cccccccc0006',
    '11111111-ffff-aaaa-bbbb-cccccccc0007', '11111111-ffff-aaaa-bbbb-cccccccc0008',
    '11111111-ffff-aaaa-bbbb-cccccccc0009', '11111111-ffff-aaaa-bbbb-cccccccc0010',
    '11111111-ffff-aaaa-bbbb-cccccccc0011', '11111111-ffff-aaaa-bbbb-cccccccc0012'
);

DELETE FROM exam_attempts WHERE id IN (
    '11111111-ffff-aaaa-bbbb-cccccccc0001', '11111111-ffff-aaaa-bbbb-cccccccc0002',
    '11111111-ffff-aaaa-bbbb-cccccccc0003', '11111111-ffff-aaaa-bbbb-cccccccc0004',
    '11111111-ffff-aaaa-bbbb-cccccccc0005', '11111111-ffff-aaaa-bbbb-cccccccc0006',
    '11111111-ffff-aaaa-bbbb-cccccccc0007', '11111111-ffff-aaaa-bbbb-cccccccc0008',
    '11111111-ffff-aaaa-bbbb-cccccccc0009', '11111111-ffff-aaaa-bbbb-cccccccc0010',
    '11111111-ffff-aaaa-bbbb-cccccccc0011', '11111111-ffff-aaaa-bbbb-cccccccc0012'
);

DELETE FROM exam_participants WHERE exam_id IN (
    '11111111-cccc-dddd-eeee-ffffffff0001',
    '11111111-cccc-dddd-eeee-ffffffff0002',
    '11111111-cccc-dddd-eeee-ffffffff0003'
);

DELETE FROM exam_schedules WHERE exam_id IN (
    '11111111-cccc-dddd-eeee-ffffffff0001',
    '11111111-cccc-dddd-eeee-ffffffff0002',
    '11111111-cccc-dddd-eeee-ffffffff0003'
);

DELETE FROM exam_section_questions WHERE section_id IN (
    '11111111-dddd-eeee-ffff-aaaaaaaa0001',
    '11111111-dddd-eeee-ffff-aaaaaaaa0002',
    '11111111-dddd-eeee-ffff-aaaaaaaa0003',
    '11111111-dddd-eeee-ffff-aaaaaaaa0004'
);

DELETE FROM exam_sections WHERE id IN (
    '11111111-dddd-eeee-ffff-aaaaaaaa0001',
    '11111111-dddd-eeee-ffff-aaaaaaaa0002',
    '11111111-dddd-eeee-ffff-aaaaaaaa0003',
    '11111111-dddd-eeee-ffff-aaaaaaaa0004'
);

DELETE FROM exams WHERE id IN (
    '11111111-cccc-dddd-eeee-ffffffff0001',
    '11111111-cccc-dddd-eeee-ffffffff0002',
    '11111111-cccc-dddd-eeee-ffffffff0003'
);

DELETE FROM question_options WHERE question_id IN (
    '11111111-bbbb-cccc-dddd-eeeeeeee0001', '11111111-bbbb-cccc-dddd-eeeeeeee0002',
    '11111111-bbbb-cccc-dddd-eeeeeeee0003', '11111111-bbbb-cccc-dddd-eeeeeeee0004',
    '11111111-bbbb-cccc-dddd-eeeeeeee0005', '11111111-bbbb-cccc-dddd-eeeeeeee0006',
    '11111111-bbbb-cccc-dddd-eeeeeeee0007', '11111111-bbbb-cccc-dddd-eeeeeeee0008',
    '11111111-bbbb-cccc-dddd-eeeeeeee0009', '11111111-bbbb-cccc-dddd-eeeeeeee000a'
);

DELETE FROM questions WHERE id IN (
    '11111111-bbbb-cccc-dddd-eeeeeeee0001', '11111111-bbbb-cccc-dddd-eeeeeeee0002',
    '11111111-bbbb-cccc-dddd-eeeeeeee0003', '11111111-bbbb-cccc-dddd-eeeeeeee0004',
    '11111111-bbbb-cccc-dddd-eeeeeeee0005', '11111111-bbbb-cccc-dddd-eeeeeeee0006',
    '11111111-bbbb-cccc-dddd-eeeeeeee0007', '11111111-bbbb-cccc-dddd-eeeeeeee0008',
    '11111111-bbbb-cccc-dddd-eeeeeeee0009', '11111111-bbbb-cccc-dddd-eeeeeeee000a'
);

DELETE FROM question_banks WHERE id IN (
    '11111111-aaaa-bbbb-cccc-dddddddd0001', '11111111-aaaa-bbbb-cccc-dddddddd0002',
    '11111111-aaaa-bbbb-cccc-dddddddd0003', '11111111-aaaa-bbbb-cccc-dddddddd0004'
);

DELETE FROM students WHERE user_id IN (
    'f1111111-1111-1111-1111-111111111111', 'f2222222-2222-2222-2222-222222222222',
    'f3333333-3333-3333-3333-333333333333', 'f4444444-4444-4444-4444-444444444444',
    'f5555555-5555-5555-5555-555555555555', 'f6666666-6666-6666-6666-666666666666',
    'f7777777-7777-7777-7777-777777777777', 'f8888888-8888-8888-8888-888888888888',
    'f9999999-9999-9999-9999-999999999999', 'faaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
);

DELETE FROM teachers WHERE user_id IN (
    'c1111111-1111-1111-1111-111111111111', 'c2222222-2222-2222-2222-222222222222',
    'c3333333-3333-3333-3333-333333333333', 'c4444444-4444-4444-4444-444444444444'
);

DELETE FROM user_roles WHERE user_id IN (
    'c1111111-1111-1111-1111-111111111111', 'c2222222-2222-2222-2222-222222222222',
    'c3333333-3333-3333-3333-333333333333', 'c4444444-4444-4444-4444-444444444444',
    'f1111111-1111-1111-1111-111111111111', 'f2222222-2222-2222-2222-222222222222',
    'f3333333-3333-3333-3333-333333333333', 'f4444444-4444-4444-4444-444444444444',
    'f5555555-5555-5555-5555-555555555555', 'f6666666-6666-6666-6666-666666666666',
    'f7777777-7777-7777-7777-777777777777', 'f8888888-8888-8888-8888-888888888888',
    'f9999999-9999-9999-9999-999999999999', 'faaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
);

DELETE FROM users WHERE id IN (
    'c1111111-1111-1111-1111-111111111111', 'c2222222-2222-2222-2222-222222222222',
    'c3333333-3333-3333-3333-333333333333', 'c4444444-4444-4444-4444-444444444444',
    'f1111111-1111-1111-1111-111111111111', 'f2222222-2222-2222-2222-222222222222',
    'f3333333-3333-3333-3333-333333333333', 'f4444444-4444-4444-4444-444444444444',
    'f5555555-5555-5555-5555-555555555555', 'f6666666-6666-6666-6666-666666666666',
    'f7777777-7777-7777-7777-777777777777', 'f8888888-8888-8888-8888-888888888888',
    'f9999999-9999-9999-9999-999999999999', 'faaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
);

DELETE FROM classes WHERE id IN (
    'e1111111-1111-1111-1111-111111111111', 'e2222222-2222-2222-2222-222222222222',
    'e3333333-3333-3333-3333-333333333333', 'e4444444-4444-4444-4444-444444444444'
);

DELETE FROM academic_years WHERE id IN (
    'b1111111-1111-1111-1111-111111111111',
    'b2222222-2222-2222-2222-222222222222'
);

DELETE FROM subjects WHERE id IN (
    'a1111111-1111-1111-1111-111111111111', 'a2222222-2222-2222-2222-222222222222',
    'a3333333-3333-3333-3333-333333333333', 'a4444444-4444-4444-4444-444444444444',
    'a5555555-5555-5555-5555-555555555555', 'a6666666-6666-6666-6666-666666666666'
);
