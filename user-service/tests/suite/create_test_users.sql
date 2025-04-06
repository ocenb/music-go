INSERT INTO users (
    username,
    email,
    password,
    is_verified,
    created_at
)
VALUES 
    ('admin', 'admin@example.com', '$2a$12$3X/t9v3nbSOtPVYW664KzurkKhgyh5wJrDl4hXDdgBnnjGLZM.n2W', TRUE, NOW()),
    ('tochange', 'tochange@example.com', '$2a$12$3X/t9v3nbSOtPVYW664KzurkKhgyh5wJrDl4hXDdgBnnjGLZM.n2W', TRUE, NOW()),
    ('tofollow', 'tofollow@example.com', '$2a$12$3X/t9v3nbSOtPVYW664KzurkKhgyh5wJrDl4hXDdgBnnjGLZM.n2W', TRUE, NOW()),
    ('todelete', 'todelete@example.com', '$2a$12$3X/t9v3nbSOtPVYW664KzurkKhgyh5wJrDl4hXDdgBnnjGLZM.n2W', TRUE, NOW())
ON CONFLICT (username) DO NOTHING;