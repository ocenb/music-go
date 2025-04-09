package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/ocenb/music-go/content-service/internal/clients/cloudinaryclient"
	"github.com/ocenb/music-go/content-service/internal/clients/notificationclient"
	"github.com/ocenb/music-go/content-service/internal/clients/searchclient"
	"github.com/ocenb/music-go/content-service/internal/clients/userclient"
	"github.com/ocenb/music-go/content-service/internal/config"
	"github.com/ocenb/music-go/content-service/internal/modules/file"
	"github.com/ocenb/music-go/content-service/internal/modules/history"
	"github.com/ocenb/music-go/content-service/internal/modules/playlist"
	"github.com/ocenb/music-go/content-service/internal/modules/playlist/playlisttracks"
	"github.com/ocenb/music-go/content-service/internal/modules/track"
	"github.com/ocenb/music-go/content-service/internal/utils"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type App struct {
	server *http.Server
	log    *slog.Logger
}

func New(postgres *sql.DB, cfg *config.Config, log *slog.Logger) *App {
	cloudinary, err := cloudinaryclient.NewCloudinaryClient(
		cfg.CLOUDINARY_CLOUD_NAME,
		cfg.CLOUDINARY_API_KEY,
		cfg.CLOUDINARY_API_SECRET,
	)
	if err != nil {
		log.Error("Failed to create cloudinary service", utils.ErrLog(err))
		os.Exit(1)
	}

	searchServiceClient, err := searchclient.New(cfg)
	if err != nil {
		log.Error("Failed to connect to search service", utils.ErrLog(err))
		os.Exit(1)
	}
	defer func() {
		log.Info("Closing search service connection")
		err := searchServiceClient.Conn.Close()
		if err != nil {
			log.Error("Failed to close search service connection", utils.ErrLog(err))
		}
	}()

	userServiceClient, err := userclient.New(cfg)
	if err != nil {
		log.Error("Failed to connect to user service", utils.ErrLog(err))
		os.Exit(1)
	}
	defer func() {
		log.Info("Closing user service connection")
		err := userServiceClient.Conn.Close()
		if err != nil {
			log.Error("Failed to close user service connection", utils.ErrLog(err))
		}
	}()

	notificationClient, err := notificationclient.NewNotificationClient(cfg.KafkaBrokers)
	if err != nil {
		log.Error("Failed to create notification client", utils.ErrLog(err))
		os.Exit(1)
	}
	defer func() {
		log.Info("Closing notification client")
		err := notificationClient.Close()
		if err != nil {
			log.Error("Failed to close notification client", utils.ErrLog(err))
		}
	}()

	fileService := file.NewFileService(
		cloudinary,
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
	historyService := history.NewHistoryService(log, historyRepo)
	historyHandler := history.NewHistoryHandler(historyService)

	if cfg.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(loggerMiddleware(log))

	api := router.Group("/api")
	api.Use(authMiddleware(userServiceClient))

	trackHandler.RegisterHandlers(api)
	playlistHandler.RegisterHandlers(api)
	playlistTracksHandler.RegisterHandlers(api)
	historyHandler.RegisterHandlers(api)

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
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		outMD := metadata.New(map[string]string{
			"authorization": accessToken,
		})
		outCtx := metadata.NewOutgoingContext(c.Request.Context(), outMD)

		res, err := userServiceClient.Client.CheckAuth(outCtx, &emptypb.Empty{})
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("userID", res.User.Id)
		c.Set("username", res.User.Username)
		c.Set("email", res.User.Email)

		c.Next()
	}
}
