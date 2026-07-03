-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_followers_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE users SET followers_count = followers_count + 1 WHERE id = NEW.user_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE users SET followers_count = followers_count - 1 WHERE id = OLD.user_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER followers_count_trigger
AFTER INSERT OR DELETE ON user_followers
FOR EACH ROW
EXECUTE FUNCTION update_followers_count();

-- +goose Down
DROP TRIGGER IF EXISTS followers_count_trigger ON user_followers;
DROP FUNCTION IF EXISTS update_followers_count();
