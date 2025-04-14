package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/ocenb/music-go/content-service/internal/clients/cloudinaryclient"
	"github.com/ocenb/music-go/content-service/internal/clients/notificationclient"
	"github.com/ocenb/music-go/content-service/internal/clients/searchclient"
	"github.com/ocenb/music-go/content-service/internal/clients/userclient"
	"github.com/ocenb/music-go/content-service/internal/config"
	"github.com/ocenb/music-go/content-service/internal/modules/all"
	"github.com/ocenb/music-go/content-service/internal/modules/file"
	"github.com/ocenb/music-go/content-service/internal/modules/history"
	"github.com/ocenb/music-go/content-service/internal/modules/playlist"
	"github.com/ocenb/music-go/content-service/internal/modules/playlist/playlisttracks"
	"github.com/ocenb/music-go/content-service/internal/modules/track"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type App struct {
	server *http.Server
	log    *slog.Logger
}

func New(postgres *sql.DB, cfg *config.Config, log *slog.Logger, cloudinary cloudinaryclient.CloudinaryClientInterface,
	searchServiceClient *searchclient.SearchServiceClient, userServiceClient *userclient.UserServiceClient,
	notificationClient notificationclient.NotificationClientInterface,
) *App {
	fileService := file.NewFileService(
		cloudinary,
		log,
		cfg,
	)
	trackRepo := track.NewTrackRepo(postgres, log)
	trackService := track.NewTrackService(log, trackRepo, fileService, searchServiceClient, notificationClient)
	trackHandler := track.NewTrackHandler(trackService)
	playlistRepo := playlist.NewPlaylistRepo(postgres, log)
	playlistService := playlist.NewPlaylistService(log, playlistRepo, fileService)
	playlistHandler := playlist.NewPlaylistHandler(playlistService)
	playlistTracksRepo := playlisttracks.NewPlaylistTracksRepo(postgres, log)
	playlistTracksService := playlisttracks.NewPlaylistTracksService(log, playlistTracksRepo, playlistRepo, trackRepo)
	playlistTracksHandler := playlisttracks.NewHandlers(playlistTracksService)
	historyRepo := history.NewHistoryRepo(postgres, log)
	historyService := history.NewHistoryService(log, historyRepo, trackService)
	historyHandler := history.NewHistoryHandler(historyService)
	allRepo := all.NewAllRepo(postgres, log)
	allService := all.NewAllService(log, allRepo, fileService)
	allHandler := all.NewAllHandler(allService)

	if cfg.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(loggerMiddleware(log))

	api := router.Group("/api")
	apiWithoutAuth := router.Group("/api")
	api.Use(authMiddleware(userServiceClient))

	trackHandler.RegisterHandlers(api)
	playlistHandler.RegisterHandlers(api)
	playlistTracksHandler.RegisterHandlers(api)
	historyHandler.RegisterHandlers(api)
	allHandler.RegisterHandlers(apiWithoutAuth)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	return &App{
		server: server,
		log:    log,
	}
}

func (a *App) Run() {
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		a.log.Error("Error starting HTTP server", "error", err)
		panic(err)
	}
}

func (a *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("Error stopping HTTP server", "error", err)
	}
}

func loggerMiddleware(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		log.Info("HTTP request",
			"status", status,
			"method", method,
			"path", path,
			"ip", clientIP,
			"latency", latency,
		)
	}
}

func authMiddleware(userServiceClient *userclient.UserServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := c.GetHeader("Authorization")
		if accessToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		outMD := metadata.New(map[string]string{
			"authorization": accessToken,
		})
		outCtx := metadata.NewOutgoingContext(c.Request.Context(), outMD)

		res, err := userServiceClient.Client.CheckAuth(outCtx, &emptypb.Empty{})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set("user", res.User)

		c.Next()
	}
}
