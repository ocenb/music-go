package elastic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/ocenb/music-go/search-service/internal/config"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrAlbumAlreadyExists = errors.New("album already exists")
	ErrTrackAlreadyExists = errors.New("track already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrAlbumNotFound      = errors.New("album not found")
	ErrTrackNotFound      = errors.New("track not found")
)

const (
	UsersIndexName  = "users"
	AlbumsIndexName = "albums"
	TracksIndexName = "tracks"
)

type ElasticClient struct {
	elastic *elasticsearch.TypedClient
}

func New(cfg *config.Config, log *slog.Logger) (*ElasticClient, error) {
	es, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{fmt.Sprintf("http://%s:%s", cfg.ElasticHost, cfg.ElasticPort)},
		Username:  cfg.ElasticUser,
		Password:  cfg.ElasticPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = es.Info().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}

	client := &ElasticClient{
		elastic: es,
	}

	if err := client.createIndices(ctx); err != nil {
		return nil, fmt.Errorf("failed to create indices: %w", err)
	}

	return client, nil
}

func (c *ElasticClient) createIndices(ctx context.Context) error {
	exists, err := c.elastic.Indices.Exists(UsersIndexName).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if users index exists: %w", err)
	}

	if !exists {
		usersMapping := types.TypeMapping{
			Properties: map[string]types.Property{
				"id":       types.NewKeywordProperty(),
				"username": types.NewTextProperty(),
			},
		}

		_, err = c.elastic.Indices.Create(UsersIndexName).
			Request(&create.Request{
				Mappings: &usersMapping,
			}).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to create users index: %w", err)
		}
	}

	exists, err = c.elastic.Indices.Exists(AlbumsIndexName).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if albums index exists: %w", err)
	}

	if !exists {
		albumsMapping := types.TypeMapping{
			Properties: map[string]types.Property{
				"id":    types.NewKeywordProperty(),
				"title": types.NewTextProperty(),
			},
		}

		_, err = c.elastic.Indices.Create(AlbumsIndexName).
			Request(&create.Request{
				Mappings: &albumsMapping,
			}).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to create albums index: %w", err)
		}
	}

	exists, err = c.elastic.Indices.Exists(TracksIndexName).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if tracks index exists: %w", err)
	}

	if !exists {
		tracksMapping := types.TypeMapping{
			Properties: map[string]types.Property{
				"id":    types.NewKeywordProperty(),
				"title": types.NewTextProperty(),
			},
		}

		_, err = c.elastic.Indices.Create(TracksIndexName).
			Request(&create.Request{
				Mappings: &tracksMapping,
			}).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to create tracks index: %w", err)
		}
	}

	return nil
}

func (c *ElasticClient) SearchUsers(ctx context.Context, query string) ([]int64, error) {
	searchQuery := &search.Request{
		Query: &types.Query{
			MultiMatch: &types.MultiMatchQuery{
				Query:  query,
				Fields: []string{"username"},
			},
		},
	}

	res, err := c.elastic.Search().
		Index(UsersIndexName).
		Request(searchQuery).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	var ids []int64
	for _, hit := range res.Hits.Hits {
		var user struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(hit.Source_, &user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user: %w", err)
		}
		ids = append(ids, user.ID)
	}

	return ids, nil
}

func (c *ElasticClient) SearchAlbums(ctx context.Context, query string) ([]int64, error) {
	searchQuery := &search.Request{
		Query: &types.Query{
			MultiMatch: &types.MultiMatchQuery{
				Query:  query,
				Fields: []string{"title"},
			},
		},
	}

	res, err := c.elastic.Search().
		Index(AlbumsIndexName).
		Request(searchQuery).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search albums: %w", err)
	}

	var ids []int64
	for _, hit := range res.Hits.Hits {
		var album struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(hit.Source_, &album); err != nil {
			return nil, fmt.Errorf("failed to unmarshal album: %w", err)
		}
		ids = append(ids, album.ID)
	}

	return ids, nil
}

func (c *ElasticClient) SearchTracks(ctx context.Context, query string) ([]int64, error) {
	searchQuery := &search.Request{
		Query: &types.Query{
			MultiMatch: &types.MultiMatchQuery{
				Query:  query,
				Fields: []string{"title"},
			},
		},
	}

	res, err := c.elastic.Search().
		Index(TracksIndexName).
		Request(searchQuery).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search tracks: %w", err)
	}

	var ids []int64
	for _, hit := range res.Hits.Hits {
		var track struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(hit.Source_, &track); err != nil {
			return nil, fmt.Errorf("failed to unmarshal track: %w", err)
		}
		ids = append(ids, track.ID)
	}

	return ids, nil
}

func (c *ElasticClient) AddUser(ctx context.Context, id int64, username string) error {
	exists, err := c.elastic.Exists(UsersIndexName, fmt.Sprintf("%d", id)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}
	if exists {
		return ErrUserAlreadyExists
	}

	user := struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	}{
		ID:       id,
		Username: username,
	}

	_, err = c.elastic.Index(UsersIndexName).
		Id(fmt.Sprintf("%d", id)).
		Document(user).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to add user: %w", err)
	}

	return nil
}

func (c *ElasticClient) AddAlbum(ctx context.Context, id int64, title string) error {
	exists, err := c.elastic.Exists(AlbumsIndexName, fmt.Sprintf("%d", id)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if album exists: %w", err)
	}
	if exists {
		return ErrAlbumAlreadyExists
	}

	album := struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}{
		ID:    id,
		Title: title,
	}

	_, err = c.elastic.Index(AlbumsIndexName).
		Id(fmt.Sprintf("%d", id)).
		Document(album).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to add album: %w", err)
	}

	return nil
}

func (c *ElasticClient) AddTrack(ctx context.Context, id int64, title string) error {
	exists, err := c.elastic.Exists(TracksIndexName, fmt.Sprintf("%d", id)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if track exists: %w", err)
	}
	if exists {
		return ErrTrackAlreadyExists
	}

	track := struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}{
		ID:    id,
		Title: title,
	}

	_, err = c.elastic.Index(TracksIndexName).
		Id(fmt.Sprintf("%d", id)).
		Document(track).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to add track: %w", err)
	}

	return nil
}

func (c *ElasticClient) UpdateUser(ctx context.Context, id int64, username string) error {
	exists, err := c.elastic.Exists(UsersIndexName, fmt.Sprintf("%d", id)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}
	if !exists {
		return c.AddUser(ctx, id, username)
	}

	user := struct {
		Username string `json:"username"`
	}{
		Username: username,
	}

	_, err = c.elastic.Update(UsersIndexName, fmt.Sprintf("%d", id)).
		Doc(user).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (c *ElasticClient) UpdateAlbum(ctx context.Context, id int64, title string) error {
	exists, err := c.elastic.Exists(AlbumsIndexName, fmt.Sprintf("%d", id)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if album exists: %w", err)
	}
	if !exists {
		return c.AddAlbum(ctx, id, title)
	}

	album := struct {
		Title string `json:"title"`
	}{
		Title: title,
	}

	_, err = c.elastic.Update(AlbumsIndexName, fmt.Sprintf("%d", id)).
		Doc(album).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to update album: %w", err)
	}

	return nil
}

func (c *ElasticClient) UpdateTrack(ctx context.Context, id int64, title string) error {
	exists, err := c.elastic.Exists(TracksIndexName, fmt.Sprintf("%d", id)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if track exists: %w", err)
	}
	if !exists {
		return c.AddTrack(ctx, id, title)
	}

	track := struct {
		Title string `json:"title"`
	}{
		Title: title,
	}

	_, err = c.elastic.Update(TracksIndexName, fmt.Sprintf("%d", id)).
		Doc(track).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to update track: %w", err)
	}

	return nil
}

func (c *ElasticClient) DeleteUser(ctx context.Context, id int64) error {
	exists, err := c.elastic.Exists(UsersIndexName, fmt.Sprintf("%d", id)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}
	if !exists {
		return ErrUserNotFound
	}

	_, err = c.elastic.Delete(UsersIndexName, fmt.Sprintf("%d", id)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (c *ElasticClient) DeleteAlbum(ctx context.Context, id int64) error {
	exists, err := c.elastic.Exists(AlbumsIndexName, fmt.Sprintf("%d", id)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if album exists: %w", err)
	}
	if !exists {
		return ErrAlbumNotFound
	}

	_, err = c.elastic.Delete(AlbumsIndexName, fmt.Sprintf("%d", id)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete album: %w", err)
	}

	return nil
}

func (c *ElasticClient) DeleteTrack(ctx context.Context, id int64) error {
	exists, err := c.elastic.Exists(TracksIndexName, fmt.Sprintf("%d", id)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if track exists: %w", err)
	}
	if !exists {
		return ErrTrackNotFound
	}

	_, err = c.elastic.Delete(TracksIndexName, fmt.Sprintf("%d", id)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete track: %w", err)
	}

	return nil
}
