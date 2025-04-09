package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ocenb/music-go/content-service/internal/modules/playlist/playlisttracks"
	"github.com/ocenb/music-go/content-service/tests/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPlaylistTracks(t *testing.T) {
	s := suite.New(t)

	token := suite.FakeAuthToken()

	resp, err := s.ContentClient.GetPlaylistTracks(token, suite.TestPlaylistID, 0)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := suite.ReadResponseBody(resp)
	require.NoError(t, err)

	var tracks []*playlisttracks.TrackInPlaylistModel
	err = json.Unmarshal(body, &tracks)
	require.NoError(t, err)

	assert.NotEmpty(t, tracks)
	assert.Equal(t, suite.TestTrackID1, tracks[0].TrackID)
	assert.Equal(t, suite.TestPlaylistID, tracks[0].PlaylistID)
}

func TestAddTrackToPlaylist(t *testing.T) {
	s := suite.New(t)

	token := suite.FakeAuthToken()
	position := suite.FakeTrackPosition()

	resp, err := s.ContentClient.AddTrackToPlaylist(token, suite.TestPlaylistID, suite.TestTrackID2, position)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	respGet, err := s.ContentClient.GetPlaylistTracks(token, suite.TestPlaylistID, 0)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respGet.StatusCode)

	body, err := suite.ReadResponseBody(respGet)
	require.NoError(t, err)

	var tracks []*playlisttracks.TrackInPlaylistModel
	err = json.Unmarshal(body, &tracks)
	require.NoError(t, err)

	var found bool
	for _, track := range tracks {
		if track.TrackID == suite.TestTrackID2 {
			found = true
			assert.Equal(t, position, track.Position)
			break
		}
	}
	assert.True(t, found, "Добавленный трек не найден в плейлисте")

	respDuplicate, err := s.ContentClient.AddTrackToPlaylist(token, suite.TestPlaylistID, suite.TestTrackID2, position)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, respDuplicate.StatusCode)
}

func TestUpdateTrackPosition(t *testing.T) {
	s := suite.New(t)

	token := suite.FakeAuthToken()
	newPosition := suite.FakeTrackPosition()

	resp, err := s.ContentClient.UpdateTrackPosition(token, suite.TestPlaylistID, suite.TestTrackID1, newPosition)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	respGet, err := s.ContentClient.GetPlaylistTracks(token, suite.TestPlaylistID, 0)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respGet.StatusCode)

	body, err := suite.ReadResponseBody(respGet)
	require.NoError(t, err)

	var tracks []*playlisttracks.TrackInPlaylistModel
	err = json.Unmarshal(body, &tracks)
	require.NoError(t, err)

	for _, track := range tracks {
		if track.TrackID == suite.TestTrackID1 {
			assert.Equal(t, newPosition, track.Position)
			break
		}
	}
}

func TestRemoveTrackFromPlaylist(t *testing.T) {
	s := suite.New(t)

	token := suite.FakeAuthToken()

	position := suite.FakeTrackPosition()
	resp, err := s.ContentClient.AddTrackToPlaylist(token, suite.TestPlaylistID, suite.TestTrackID3, position)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	respDelete, err := s.ContentClient.RemoveTrackFromPlaylist(token, suite.TestPlaylistID, suite.TestTrackID3)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, respDelete.StatusCode)

	respGet, err := s.ContentClient.GetPlaylistTracks(token, suite.TestPlaylistID, 0)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respGet.StatusCode)

	body, err := suite.ReadResponseBody(respGet)
	require.NoError(t, err)

	var tracks []*playlisttracks.TrackInPlaylistModel
	err = json.Unmarshal(body, &tracks)
	require.NoError(t, err)

	var found bool
	for _, track := range tracks {
		if track.TrackID == suite.TestTrackID3 {
			found = true
			break
		}
	}
	assert.False(t, found, "Удаленный трек все еще в плейлисте")
}
