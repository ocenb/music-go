package track_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ocenb/music-go/content-service/internal/errs"
	"github.com/ocenb/music-go/content-service/internal/services/track"
	"github.com/ocenb/music-go/content-service/internal/storage/transactor"
)

type TrackServiceSuite struct {
	suite.Suite
	mockRepo               *track.MockRepo
	mockFileService        *track.MockFileService
	mockSearchClient       *track.MockSearchClient
	mockNotificationClient *track.MockNotificationClient
	tm                     transactor.Runner
	service                *track.Service
}

func (s *TrackServiceSuite) SetupTest() {
	s.mockRepo = track.NewMockRepo(s.T())
	s.mockFileService = track.NewMockFileService(s.T())
	s.mockSearchClient = track.NewMockSearchClient(s.T())
	s.mockNotificationClient = track.NewMockNotificationClient(s.T())
	s.tm = transactor.NewMock()
	s.service = track.New(s.mockRepo, s.mockFileService, s.mockSearchClient, s.mockNotificationClient, s.tm)
}

func TestTrackServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(TrackServiceSuite))
}

func (s *TrackServiceSuite) TestGetOneByID_NotFound() {
	ctx := context.Background()

	s.mockRepo.On("GetByID", ctx, int64(999), int64(1)).Return(nil, errs.ErrTrackNotFound)

	_, err := s.service.GetOneByID(ctx, 1, 999)
	s.Require().Error(err)
	s.ErrorIs(err, errs.ErrTrackNotFound)
}

func (s *TrackServiceSuite) TestUpload_DuplicateTitle() {
	ctx := context.Background()

	s.mockRepo.On("CheckTitle", ctx, int64(1), "duplicate").Return(true, nil)

	_, err := s.service.Upload(ctx, 1, "user", "a@b.c", "duplicate", "id", nil, nil)
	s.Require().Error(err)
	s.ErrorIs(err, errs.ErrTitleExists)
}
