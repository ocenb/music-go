TRUNCATE TABLE user_followers, tokens, users RESTART IDENTITY CASCADE;

INSERT INTO users (username, email, password, is_verified, created_at)
VALUES
    ('admin', 'admin@example.com', '$2a$12$3X/t9v3nbSOtPVYW664KzurkKhgyh5wJrDl4hXDdgBnnjGLZM.n2W', TRUE, NOW()),
    ('tochange', 'tochange@example.com', '$2a$12$3X/t9v3nbSOtPVYW664KzurkKhgyh5wJrDl4hXDdgBnnjGLZM.n2W', TRUE, NOW()),
    ('changename', 'changename@example.com', '$2a$12$3X/t9v3nbSOtPVYW664KzurkKhgyh5wJrDl4hXDdgBnnjGLZM.n2W', TRUE, NOW()),
    ('tofollow', 'tofollow@example.com', '$2a$12$3X/t9v3nbSOtPVYW664KzurkKhgyh5wJrDl4hXDdgBnnjGLZM.n2W', TRUE, NOW()),
    ('todelete', 'todelete@example.com', '$2a$12$3X/t9v3nbSOtPVYW664KzurkKhgyh5wJrDl4hXDdgBnnjGLZM.n2W', TRUE, NOW());

SELECT setval(pg_get_serial_sequence('users', 'id'), (SELECT COALESCE(MAX(id), 1) FROM users));
