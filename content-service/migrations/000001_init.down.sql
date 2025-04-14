DROP TRIGGER IF EXISTS update_playlists_updated_at ON playlists;
DROP TRIGGER IF EXISTS update_tracks_updated_at ON tracks;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS listening_history;
DROP TABLE IF EXISTS playlist_tracks;
DROP TABLE IF EXISTS user_saved_playlists;
DROP TABLE IF EXISTS user_liked_tracks;
DROP TABLE IF EXISTS playlists;
DROP TABLE IF EXISTS tracks;
