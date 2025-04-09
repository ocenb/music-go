package suite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func (c *ContentClient) GetPlaylistTracks(token string, playlistID int64, take int) (*http.Response, error) {
	endpoint := fmt.Sprintf("%s/api/v1/playlists/%d/tracks", c.BaseURL, playlistID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	if take > 0 {
		q := req.URL.Query()
		q.Add("take", strconv.Itoa(take))
		req.URL.RawQuery = q.Encode()
	}

	return c.HTTPClient.Do(req)
}

func (c *ContentClient) AddTrackToPlaylist(token string, playlistID, trackID int64, position int) (*http.Response, error) {
	endpoint := fmt.Sprintf("%s/api/v1/playlists/%d/tracks/%d", c.BaseURL, playlistID, trackID)

	requestBody := map[string]int{"position": position}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return c.HTTPClient.Do(req)
}

func (c *ContentClient) UpdateTrackPosition(token string, playlistID, trackID int64, position int) (*http.Response, error) {
	endpoint := fmt.Sprintf("%s/api/v1/playlists/%d/tracks/%d", c.BaseURL, playlistID, trackID)

	requestBody := map[string]int{"position": position}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return c.HTTPClient.Do(req)
}

func (c *ContentClient) RemoveTrackFromPlaylist(token string, playlistID, trackID int64) (*http.Response, error) {
	endpoint := fmt.Sprintf("%s/api/v1/playlists/%d/tracks/%d", c.BaseURL, playlistID, trackID)

	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	return c.HTTPClient.Do(req)
}

func ReadResponseBody(resp *http.Response) ([]byte, error) {
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Println("Error closing response body", err)
		}
	}()
	return io.ReadAll(resp.Body)
}
