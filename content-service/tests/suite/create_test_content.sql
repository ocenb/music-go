INSERT INTO playlists (
    id,
    user_id,
    username,
    title,
    changeable_id,
    image,
    created_at,
    updated_at
)
VALUES 
    (1, 1, 'admin', 'Test Playlist', 'test-playlist', 'test-image.jpg', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO tracks (
    id,
    user_id,
    username,
    title,
    changeable_id,
    duration,
    plays,
    audio,
    image,
    created_at,
    updated_at
)
VALUES 
    (1, 1, 'admin', 'Test Track 1', 'test-track1', 180, 100, 'test-audio1.mp3', 'test-image1.jpg', NOW(), NOW()),
    (2, 1, 'admin', 'Test Track 2', 'test-track2', 240, 150, 'test-audio2.mp3', 'test-image2.jpg', NOW(), NOW()),
    (3, 1, 'admin', 'Test Track 3', 'test-track3', 300, 200, 'test-audio3.mp3', 'test-image3.jpg', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO playlist_tracks (
    playlist_id,
    track_id,
    position,
    added_at
)
VALUES 
    (1, 1, 1, NOW())
ON CONFLICT (playlist_id, track_id) DO NOTHING;
