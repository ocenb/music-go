package search_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ocenb/music-go/search-service/internal/errs"
	"github.com/ocenb/music-go/search-service/internal/services/search"
)

type mockRepo struct {
	searchUsersFn func(ctx context.Context, query string) ([]int64, error)
	addUserFn     func(ctx context.Context, id int64, username string) error
}

func (m *mockRepo) SearchUsers(ctx context.Context, query string) ([]int64, error) {
	return m.searchUsersFn(ctx, query)
}

func (m *mockRepo) SearchAlbums(context.Context, string) ([]int64, error) { return nil, nil }
func (m *mockRepo) SearchTracks(context.Context, string) ([]int64, error) { return nil, nil }

func (m *mockRepo) AddUser(ctx context.Context, id int64, username string) error {
	if m.addUserFn != nil {
		return m.addUserFn(ctx, id, username)
	}
	return nil
}

func (m *mockRepo) AddAlbum(context.Context, int64, string) error   { return nil }
func (m *mockRepo) AddTrack(context.Context, int64, string) error   { return nil }
func (m *mockRepo) UpdateUser(context.Context, int64, string) error { return nil }
func (m *mockRepo) UpdateAlbum(context.Context, int64, string) error {
	return nil
}
func (m *mockRepo) UpdateTrack(context.Context, int64, string) error { return nil }
func (m *mockRepo) DeleteUser(context.Context, int64) error          { return nil }
func (m *mockRepo) DeleteAlbum(context.Context, int64) error         { return nil }
func (m *mockRepo) DeleteTrack(context.Context, int64) error         { return nil }

func TestService_SearchUsers_Success(t *testing.T) {
	repo := &mockRepo{
		searchUsersFn: func(_ context.Context, query string) ([]int64, error) {
			require.Equal(t, "alice", query)
			return []int64{1, 2}, nil
		},
	}

	svc := search.New(repo)
	ids, err := svc.SearchUsers(context.Background(), "alice")
	require.NoError(t, err)
	require.Equal(t, []int64{1, 2}, ids)
}

func TestService_AddUser_AlreadyExists(t *testing.T) {
	repo := &mockRepo{
		addUserFn: func(context.Context, int64, string) error {
			return errs.ErrUserAlreadyExists
		},
	}
	svc := search.New(repo)

	err := svc.AddUser(context.Background(), 1, "alice")
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrUserAlreadyExists))
}
