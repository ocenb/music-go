-- +goose Up
ALTER TABLE users ADD COLUMN followers_count INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS followers_count;
