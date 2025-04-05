CREATE TABLE IF NOT EXISTS user_followers (
    user_id INT NOT NULL,
    follower_id INT NOT NULL,
    followed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, follower_id),
    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_follower_id FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_id ON user_followers(user_id);
CREATE INDEX idx_follower_id ON user_followers(follower_id);