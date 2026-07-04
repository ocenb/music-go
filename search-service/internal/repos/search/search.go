package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"

	"github.com/ocenb/music-go/search-service/internal/errs"
	"github.com/ocenb/music-go/search-service/internal/storage/elastic"
)

type Repo struct {
	es *elastic.Client
}

func New(es *elastic.Client) *Repo {
	return &Repo{es: es}
}

func (r *Repo) SearchUsers(ctx context.Context, query string) ([]int64, error) {
	return r.search(ctx, elastic.UsersIndexName, query, []string{"username"})
}

func (r *Repo) SearchAlbums(ctx context.Context, query string) ([]int64, error) {
	return r.search(ctx, elastic.AlbumsIndexName, query, []string{"title"})
}

func (r *Repo) SearchTracks(ctx context.Context, query string) ([]int64, error) {
	return r.search(ctx, elastic.TracksIndexName, query, []string{"title"})
}

func (r *Repo) search(ctx context.Context, index, query string, fields []string) ([]int64, error) {
	searchQuery := &search.Request{
		Query: &types.Query{
			MultiMatch: &types.MultiMatchQuery{
				Query:  query,
				Fields: fields,
			},
		},
	}

	res, err := r.es.Typed().Search().
		Index(index).
		Request(searchQuery).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search %s: %w", index, err)
	}

	var ids []int64
	for _, hit := range res.Hits.Hits {
		var doc struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(hit.Source_, &doc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s document: %w", index, err)
		}
		ids = append(ids, doc.ID)
	}

	return ids, nil
}

func (r *Repo) AddUser(ctx context.Context, id int64, username string) error {
	exists, err := r.es.Typed().Exists(elastic.UsersIndexName, strconv.FormatInt(id, 10)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}
	if exists {
		return errs.ErrUserAlreadyExists
	}

	user := struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	}{ID: id, Username: username}

	_, err = r.es.Typed().Index(elastic.UsersIndexName).
		Id(strconv.FormatInt(id, 10)).
		Document(user).
		Refresh(refresh.Waitfor).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to add user: %w", err)
	}

	return nil
}

func (r *Repo) AddAlbum(ctx context.Context, id int64, title string) error {
	exists, err := r.es.Typed().Exists(elastic.AlbumsIndexName, strconv.FormatInt(id, 10)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if album exists: %w", err)
	}
	if exists {
		return errs.ErrAlbumAlreadyExists
	}

	album := struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}{ID: id, Title: title}

	_, err = r.es.Typed().Index(elastic.AlbumsIndexName).
		Id(strconv.FormatInt(id, 10)).
		Document(album).
		Refresh(refresh.Waitfor).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to add album: %w", err)
	}

	return nil
}

func (r *Repo) AddTrack(ctx context.Context, id int64, title string) error {
	exists, err := r.es.Typed().Exists(elastic.TracksIndexName, strconv.FormatInt(id, 10)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if track exists: %w", err)
	}
	if exists {
		return errs.ErrTrackAlreadyExists
	}

	track := struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}{ID: id, Title: title}

	_, err = r.es.Typed().Index(elastic.TracksIndexName).
		Id(strconv.FormatInt(id, 10)).
		Document(track).
		Refresh(refresh.Waitfor).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to add track: %w", err)
	}

	return nil
}

func (r *Repo) UpdateUser(ctx context.Context, id int64, username string) error {
	exists, err := r.es.Typed().Exists(elastic.UsersIndexName, strconv.FormatInt(id, 10)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}
	if !exists {
		return r.AddUser(ctx, id, username)
	}

	user := struct {
		Username string `json:"username"`
	}{Username: username}

	_, err = r.es.Typed().Update(elastic.UsersIndexName, strconv.FormatInt(id, 10)).
		Doc(user).
		Refresh(refresh.Waitfor).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *Repo) UpdateAlbum(ctx context.Context, id int64, title string) error {
	exists, err := r.es.Typed().Exists(elastic.AlbumsIndexName, strconv.FormatInt(id, 10)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if album exists: %w", err)
	}
	if !exists {
		return r.AddAlbum(ctx, id, title)
	}

	album := struct {
		Title string `json:"title"`
	}{Title: title}

	_, err = r.es.Typed().Update(elastic.AlbumsIndexName, strconv.FormatInt(id, 10)).
		Doc(album).
		Refresh(refresh.Waitfor).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to update album: %w", err)
	}

	return nil
}

func (r *Repo) UpdateTrack(ctx context.Context, id int64, title string) error {
	exists, err := r.es.Typed().Exists(elastic.TracksIndexName, strconv.FormatInt(id, 10)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if track exists: %w", err)
	}
	if !exists {
		return r.AddTrack(ctx, id, title)
	}

	track := struct {
		Title string `json:"title"`
	}{Title: title}

	_, err = r.es.Typed().Update(elastic.TracksIndexName, strconv.FormatInt(id, 10)).
		Doc(track).
		Refresh(refresh.Waitfor).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to update track: %w", err)
	}

	return nil
}

func (r *Repo) DeleteUser(ctx context.Context, id int64) error {
	return r.delete(ctx, elastic.UsersIndexName, id, errs.ErrUserNotFound)
}

func (r *Repo) DeleteAlbum(ctx context.Context, id int64) error {
	return r.delete(ctx, elastic.AlbumsIndexName, id, errs.ErrAlbumNotFound)
}

func (r *Repo) DeleteTrack(ctx context.Context, id int64) error {
	return r.delete(ctx, elastic.TracksIndexName, id, errs.ErrTrackNotFound)
}

func (r *Repo) delete(ctx context.Context, index string, id int64, notFound *errs.Error) error {
	exists, err := r.es.Typed().Exists(index, strconv.FormatInt(id, 10)).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if document exists in %s: %w", index, err)
	}
	if !exists {
		return notFound
	}

	_, err = r.es.Typed().Delete(index, strconv.FormatInt(id, 10)).
		Refresh(refresh.Waitfor).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete document from %s: %w", index, err)
	}

	return nil
}
