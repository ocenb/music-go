package all

import (
	"context"
	"fmt"

	"github.com/ocenb/music-go/content-service/internal/storage/transactor"
)

type Repo struct {
	tm transactor.Manager
}

func New(tm transactor.Manager) *Repo {
	return &Repo{tm: tm}
}

func (r *Repo) DeleteAll(ctx context.Context, userID int64) ([]string, []string, []string, error) {
	var trackAudios, trackImages, playlistImages []string

	err := r.tm.Run(ctx, func(ctx context.Context) error {
		q := r.tm.GetQueryEngine(ctx)

		rows, err := q.Query(ctx, "SELECT audio FROM tracks WHERE user_id = $1", userID)
		if err != nil {
			return fmt.Errorf("failed to get track audios: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var audio string
			if scanErr := rows.Scan(&audio); scanErr != nil {
				return fmt.Errorf("failed to scan track audio: %w", scanErr)
			}
			trackAudios = append(trackAudios, audio)
		}
		if iterErr := rows.Err(); iterErr != nil {
			return fmt.Errorf("failed to iterate track audios: %w", iterErr)
		}

		rows, err = q.Query(ctx, "SELECT image FROM tracks WHERE user_id = $1 AND image != 'default'", userID)
		if err != nil {
			return fmt.Errorf("failed to get track images: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var image string
			if scanErr := rows.Scan(&image); scanErr != nil {
				return fmt.Errorf("failed to scan track image: %w", scanErr)
			}
			trackImages = append(trackImages, image)
		}
		if iterErr := rows.Err(); iterErr != nil {
			return fmt.Errorf("failed to iterate track images: %w", iterErr)
		}

		rows, err = q.Query(ctx, "SELECT image FROM playlists WHERE user_id = $1", userID)
		if err != nil {
			return fmt.Errorf("failed to get playlist images: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var image string
			if scanErr := rows.Scan(&image); scanErr != nil {
				return fmt.Errorf("failed to scan playlist image: %w", scanErr)
			}
			playlistImages = append(playlistImages, image)
		}
		if iterErr := rows.Err(); iterErr != nil {
			return fmt.Errorf("failed to iterate playlist images: %w", iterErr)
		}

		if _, err := q.Exec(ctx, "DELETE FROM listening_history WHERE user_id = $1", userID); err != nil {
			return fmt.Errorf("failed to delete listening history: %w", err)
		}

		if _, err := q.Exec(ctx, "DELETE FROM user_liked_tracks WHERE user_id = $1", userID); err != nil {
			return fmt.Errorf("failed to delete user liked tracks: %w", err)
		}

		if _, err := q.Exec(ctx, "DELETE FROM user_saved_playlists WHERE user_id = $1", userID); err != nil {
			return fmt.Errorf("failed to delete user saved playlists: %w", err)
		}

		if _, err := q.Exec(ctx, "DELETE FROM playlists WHERE user_id = $1", userID); err != nil {
			return fmt.Errorf("failed to delete playlists: %w", err)
		}

		if _, err := q.Exec(ctx, "DELETE FROM tracks WHERE user_id = $1", userID); err != nil {
			return fmt.Errorf("failed to delete tracks: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return trackAudios, trackImages, playlistImages, nil
}
