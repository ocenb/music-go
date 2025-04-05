DELETE FROM users WHERE username IN ('admin', 'todelete', 'tochange');
DELETE FROM users WHERE email IN ('admin@example.com', 'todelete@example.com', 'tochange@example.com');