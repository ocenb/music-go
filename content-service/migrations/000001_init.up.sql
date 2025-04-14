CREATE TABLE IF NOT EXISTS tracks (
    id SERIAL PRIMARY KEY,
    changeable_id TEXT NOT NULL,
    title TEXT NOT NULL,
    duration INT NOT NULL,
    plays INT NOT NULL DEFAULT 0,
    audio TEXT NOT NULL,
    image TEXT NOT NULL DEFAULT 'default',
    user_id INT NOT NULL,
		username TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_user_changeable_id UNIQUE (user_id, changeable_id),
    CONSTRAINT unique_user_title UNIQUE (user_id, title)
);

CREATE INDEX IF NOT EXISTS idx_tracks_user_id ON tracks(user_id);
CREATE INDEX IF NOT EXISTS idx_tracks_username_changeable_id ON tracks(username, changeable_id);

CREATE TABLE IF NOT EXISTS playlists (
    id SERIAL PRIMARY KEY,
    changeable_id TEXT NOT NULL,
    title TEXT NOT NULL,
    image TEXT NOT NULL,
    user_id INT NOT NULL,
    username TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_user_playlist_changeable_id UNIQUE (user_id, changeable_id),
    CONSTRAINT unique_user_playlist_title UNIQUE (user_id, title)
);

CREATE INDEX IF NOT EXISTS idx_playlists_user_id ON playlists(user_id);
CREATE INDEX IF NOT EXISTS idx_playlists_username_changeable_id ON playlists(username, changeable_id);

CREATE TABLE IF NOT EXISTS user_liked_tracks (
    user_id INT NOT NULL,
    track_id INT NOT NULL,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, track_id),
    CONSTRAINT fk_user_liked_tracks_track FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_liked_tracks_user_id ON user_liked_tracks(user_id);
CREATE INDEX IF NOT EXISTS idx_user_liked_tracks_track_id ON user_liked_tracks(track_id);

CREATE TABLE IF NOT EXISTS user_saved_playlists (
    user_id INT NOT NULL,
    playlist_id INT NOT NULL,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, playlist_id),
    CONSTRAINT fk_user_saved_playlists_playlist FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_saved_playlists_user_id ON user_saved_playlists(user_id);
CREATE INDEX IF NOT EXISTS idx_user_saved_playlists_playlist_id ON user_saved_playlists(playlist_id);

CREATE TABLE IF NOT EXISTS playlist_tracks (
    playlist_id INT NOT NULL,
    track_id INT NOT NULL,
    position INT NOT NULL,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (playlist_id, track_id),
    CONSTRAINT fk_playlist_tracks_playlist FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE,
    CONSTRAINT fk_playlist_tracks_track FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS listening_history (
    user_id INT NOT NULL,
    track_id INT NOT NULL,
    played_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, track_id),
    CONSTRAINT fk_listening_history_track FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_tracks_updated_at
BEFORE UPDATE ON tracks
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_playlists_updated_at
BEFORE UPDATE ON playlists
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
