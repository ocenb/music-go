package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ocenb/music-go/content-service/internal/config"
	"github.com/ocenb/music-go/content-service/internal/handlers"
	"github.com/ocenb/music-go/content-service/internal/middlewares"
)

type Server struct {
	httpServer *http.Server
	log        *slog.Logger
}

func New(
	trackSvc handlers.TrackService,
	playlistSvc handlers.PlaylistService,
	playlistTracksSvc handlers.PlaylistTracksService,
	historySvc handlers.HistoryService,
	searchSvc handlers.SearchService,
	allSvc handlers.AllService,
	userClient middlewares.UserClient,
	internalSecret string,
	cfg *config.Config,
	log *slog.Logger,
) *Server {
	if cfg.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middlewares.Logging(log))

	engine.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	contentHandler := handlers.NewContentHandler(
		trackSvc,
		playlistSvc,
		playlistTracksSvc,
		historySvc,
		searchSvc,
		allSvc,
		userClient,
		internalSecret,
	)
	contentHandler.RegisterRoutes(engine)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:      engine,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	return &Server{
		httpServer: httpServer,
		log:        log,
	}
}

func (s *Server) Start() error {
	s.log.Info("HTTP server started", slog.String("addr", s.httpServer.Addr))

	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("stopping HTTP server")
	return s.httpServer.Shutdown(ctx)
}
