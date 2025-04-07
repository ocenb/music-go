package tests

import (
	"slices"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/ocenb/music-go/search-service/tests/suite"
	"github.com/ocenb/music-protos/gen/searchservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchUsers(t *testing.T) {
	ctx, s := suite.New(t)

	username := gofakeit.Username()
	updatedUsername := gofakeit.Username()
	userId := gofakeit.Int64()

	addResp, err := s.SearchClient.AddUser(ctx, &searchservice.AddOrUpdateRequest{
		Id:   userId,
		Name: username,
	})
	require.NoError(t, err)
	require.NotNil(t, addResp)
	assert.True(t, addResp.Success)

	time.Sleep(1 * time.Second)

	searchResp, err := s.SearchClient.SearchUsers(ctx, &searchservice.SearchRequest{
		Query: username,
	})
	require.NoError(t, err)
	require.NotNil(t, searchResp)

	found := slices.Contains(searchResp.Ids, userId)
	assert.True(t, found, "User should be found in search results")

	updateResp, err := s.SearchClient.UpdateUser(ctx, &searchservice.AddOrUpdateRequest{
		Id:   userId,
		Name: updatedUsername,
	})
	require.NoError(t, err)
	require.NotNil(t, updateResp)
	assert.True(t, updateResp.Success)

	time.Sleep(1 * time.Second)

	searchUpdatedResp, err := s.SearchClient.SearchUsers(ctx, &searchservice.SearchRequest{
		Query: updatedUsername,
	})
	require.NoError(t, err)
	require.NotNil(t, searchUpdatedResp)

	found = slices.Contains(searchUpdatedResp.Ids, userId)
	assert.True(t, found, "Updated user should be found in search results")

	deleteResp, err := s.SearchClient.DeleteUser(ctx, &searchservice.DeleteRequest{
		Id: userId,
	})
	require.NoError(t, err)
	require.NotNil(t, deleteResp)
	assert.True(t, deleteResp.Success)

	time.Sleep(1 * time.Second)

	searchAfterDeleteResp, err := s.SearchClient.SearchUsers(ctx, &searchservice.SearchRequest{
		Query: updatedUsername,
	})
	require.NoError(t, err)
	require.NotNil(t, searchAfterDeleteResp)

	found = slices.Contains(searchAfterDeleteResp.Ids, userId)
	assert.False(t, found, "User should not be found after deletion")
}

func TestSearchAlbums(t *testing.T) {
	ctx, s := suite.New(t)

	albumTitle := gofakeit.Word()
	updatedAlbumTitle := gofakeit.Word()
	albumId := gofakeit.Int64()

	addResp, err := s.SearchClient.AddAlbum(ctx, &searchservice.AddOrUpdateRequest{
		Id:   albumId,
		Name: albumTitle,
	})
	require.NoError(t, err)
	require.NotNil(t, addResp)
	assert.True(t, addResp.Success)

	time.Sleep(1 * time.Second)

	searchResp, err := s.SearchClient.SearchAlbums(ctx, &searchservice.SearchRequest{
		Query: albumTitle,
	})
	require.NoError(t, err)
	require.NotNil(t, searchResp)

	found := slices.Contains(searchResp.Ids, albumId)
	assert.True(t, found, "Album should be found in search results")

	updateResp, err := s.SearchClient.UpdateAlbum(ctx, &searchservice.AddOrUpdateRequest{
		Id:   albumId,
		Name: updatedAlbumTitle,
	})
	require.NoError(t, err)
	require.NotNil(t, updateResp)
	assert.True(t, updateResp.Success)

	time.Sleep(1 * time.Second)

	searchUpdatedResp, err := s.SearchClient.SearchAlbums(ctx, &searchservice.SearchRequest{
		Query: updatedAlbumTitle,
	})
	require.NoError(t, err)
	require.NotNil(t, searchUpdatedResp)

	found = slices.Contains(searchUpdatedResp.Ids, albumId)
	assert.True(t, found, "Updated album should be found in search results")

	deleteResp, err := s.SearchClient.DeleteAlbum(ctx, &searchservice.DeleteRequest{
		Id: albumId,
	})
	require.NoError(t, err)
	require.NotNil(t, deleteResp)
	assert.True(t, deleteResp.Success)

	time.Sleep(1 * time.Second)

	searchAfterDeleteResp, err := s.SearchClient.SearchAlbums(ctx, &searchservice.SearchRequest{
		Query: updatedAlbumTitle,
	})
	require.NoError(t, err)
	require.NotNil(t, searchAfterDeleteResp)

	found = slices.Contains(searchAfterDeleteResp.Ids, albumId)
	assert.False(t, found, "Album should not be found after deletion")
}

func TestSearchTracks(t *testing.T) {
	ctx, s := suite.New(t)

	trackTitle := gofakeit.Word()
	updatedTrackTitle := gofakeit.Word()
	trackId := gofakeit.Int64()

	addResp, err := s.SearchClient.AddTrack(ctx, &searchservice.AddOrUpdateRequest{
		Id:   trackId,
		Name: trackTitle,
	})
	require.NoError(t, err)
	require.NotNil(t, addResp)
	assert.True(t, addResp.Success)

	time.Sleep(1 * time.Second)

	searchResp, err := s.SearchClient.SearchTracks(ctx, &searchservice.SearchRequest{
		Query: trackTitle,
	})
	require.NoError(t, err)
	require.NotNil(t, searchResp)

	found := slices.Contains(searchResp.Ids, trackId)
	assert.True(t, found, "Track should be found in search results")

	updateResp, err := s.SearchClient.UpdateTrack(ctx, &searchservice.AddOrUpdateRequest{
		Id:   trackId,
		Name: updatedTrackTitle,
	})
	require.NoError(t, err)
	require.NotNil(t, updateResp)
	assert.True(t, updateResp.Success)

	time.Sleep(1 * time.Second)

	searchUpdatedResp, err := s.SearchClient.SearchTracks(ctx, &searchservice.SearchRequest{
		Query: updatedTrackTitle,
	})
	require.NoError(t, err)
	require.NotNil(t, searchUpdatedResp)

	found = slices.Contains(searchUpdatedResp.Ids, trackId)
	assert.True(t, found, "Updated track should be found in search results")

	deleteResp, err := s.SearchClient.DeleteTrack(ctx, &searchservice.DeleteRequest{
		Id: trackId,
	})
	require.NoError(t, err)
	require.NotNil(t, deleteResp)
	assert.True(t, deleteResp.Success)

	time.Sleep(1 * time.Second)

	searchAfterDeleteResp, err := s.SearchClient.SearchTracks(ctx, &searchservice.SearchRequest{
		Query: updatedTrackTitle,
	})
	require.NoError(t, err)
	require.NotNil(t, searchAfterDeleteResp)

	found = slices.Contains(searchAfterDeleteResp.Ids, trackId)
	assert.False(t, found, "Track should not be found after deletion")
}
