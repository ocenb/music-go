package tests_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/ocenb/music-protos/gen/searchservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchUsers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	env := setupTestEnv(ctx, t)
	authCtx := authContext(ctx)

	t.Run("lifecycle", func(t *testing.T) {
		username := gofakeit.Username()
		updatedUsername := gofakeit.Username()
		userID := gofakeit.Int64()

		addResp, err := env.Client.AddUser(ctx, &searchservice.AddOrUpdateRequest{
			Id:   userID,
			Name: username,
		})
		require.NoError(t, err)
		require.NotNil(t, addResp)
		assert.True(t, addResp.Success)

		searchResp, err := env.Client.SearchUsers(authCtx, &searchservice.SearchRequest{Query: username})
		require.NoError(t, err)
		require.NotNil(t, searchResp)
		assert.True(t, slices.Contains(searchResp.Ids, userID))

		updateResp, err := env.Client.UpdateUser(ctx, &searchservice.AddOrUpdateRequest{
			Id:   userID,
			Name: updatedUsername,
		})
		require.NoError(t, err)
		require.NotNil(t, updateResp)
		assert.True(t, updateResp.Success)

		searchUpdatedResp, err := env.Client.SearchUsers(authCtx, &searchservice.SearchRequest{Query: updatedUsername})
		require.NoError(t, err)
		require.NotNil(t, searchUpdatedResp)
		assert.True(t, slices.Contains(searchUpdatedResp.Ids, userID))

		deleteResp, err := env.Client.DeleteUser(ctx, &searchservice.DeleteRequest{Id: userID})
		require.NoError(t, err)
		require.NotNil(t, deleteResp)
		assert.True(t, deleteResp.Success)

		searchAfterDeleteResp, err := env.Client.SearchUsers(authCtx, &searchservice.SearchRequest{Query: updatedUsername})
		require.NoError(t, err)
		require.NotNil(t, searchAfterDeleteResp)
		assert.False(t, slices.Contains(searchAfterDeleteResp.Ids, userID))
	})
}

func TestSearchAlbums(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	env := setupTestEnv(ctx, t)
	authCtx := authContext(ctx)

	t.Run("lifecycle", func(t *testing.T) {
		albumTitle := gofakeit.Word()
		updatedAlbumTitle := gofakeit.Word()
		albumID := gofakeit.Int64()

		addResp, err := env.Client.AddAlbum(ctx, &searchservice.AddOrUpdateRequest{
			Id:   albumID,
			Name: albumTitle,
		})
		require.NoError(t, err)
		require.NotNil(t, addResp)
		assert.True(t, addResp.Success)

		searchResp, err := env.Client.SearchAlbums(authCtx, &searchservice.SearchRequest{Query: albumTitle})
		require.NoError(t, err)
		require.NotNil(t, searchResp)
		assert.True(t, slices.Contains(searchResp.Ids, albumID))

		updateResp, err := env.Client.UpdateAlbum(ctx, &searchservice.AddOrUpdateRequest{
			Id:   albumID,
			Name: updatedAlbumTitle,
		})
		require.NoError(t, err)
		require.NotNil(t, updateResp)
		assert.True(t, updateResp.Success)

		searchUpdatedResp, err := env.Client.SearchAlbums(authCtx, &searchservice.SearchRequest{Query: updatedAlbumTitle})
		require.NoError(t, err)
		require.NotNil(t, searchUpdatedResp)
		assert.True(t, slices.Contains(searchUpdatedResp.Ids, albumID))

		deleteResp, err := env.Client.DeleteAlbum(ctx, &searchservice.DeleteRequest{Id: albumID})
		require.NoError(t, err)
		require.NotNil(t, deleteResp)
		assert.True(t, deleteResp.Success)

		searchAfterDeleteResp, err := env.Client.SearchAlbums(authCtx, &searchservice.SearchRequest{Query: updatedAlbumTitle})
		require.NoError(t, err)
		require.NotNil(t, searchAfterDeleteResp)
		assert.False(t, slices.Contains(searchAfterDeleteResp.Ids, albumID))
	})
}

func TestSearchTracks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	env := setupTestEnv(ctx, t)
	authCtx := authContext(ctx)

	t.Run("lifecycle", func(t *testing.T) {
		trackTitle := gofakeit.Word()
		updatedTrackTitle := gofakeit.Word()
		trackID := gofakeit.Int64()

		addResp, err := env.Client.AddTrack(ctx, &searchservice.AddOrUpdateRequest{
			Id:   trackID,
			Name: trackTitle,
		})
		require.NoError(t, err)
		require.NotNil(t, addResp)
		assert.True(t, addResp.Success)

		searchResp, err := env.Client.SearchTracks(authCtx, &searchservice.SearchRequest{Query: trackTitle})
		require.NoError(t, err)
		require.NotNil(t, searchResp)
		assert.True(t, slices.Contains(searchResp.Ids, trackID))

		updateResp, err := env.Client.UpdateTrack(ctx, &searchservice.AddOrUpdateRequest{
			Id:   trackID,
			Name: updatedTrackTitle,
		})
		require.NoError(t, err)
		require.NotNil(t, updateResp)
		assert.True(t, updateResp.Success)

		searchUpdatedResp, err := env.Client.SearchTracks(authCtx, &searchservice.SearchRequest{Query: updatedTrackTitle})
		require.NoError(t, err)
		require.NotNil(t, searchUpdatedResp)
		assert.True(t, slices.Contains(searchUpdatedResp.Ids, trackID))

		deleteResp, err := env.Client.DeleteTrack(ctx, &searchservice.DeleteRequest{Id: trackID})
		require.NoError(t, err)
		require.NotNil(t, deleteResp)
		assert.True(t, deleteResp.Success)

		searchAfterDeleteResp, err := env.Client.SearchTracks(authCtx, &searchservice.SearchRequest{Query: updatedTrackTitle})
		require.NoError(t, err)
		require.NotNil(t, searchAfterDeleteResp)
		assert.False(t, slices.Contains(searchAfterDeleteResp.Ids, trackID))
	})
}

const testTimeout = 5 * time.Minute
