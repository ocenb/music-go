package handlers

import (
	"google.golang.org/grpc"

	"github.com/ocenb/music-go/user-service/internal/config"
	"github.com/ocenb/music-go/user-service/internal/services/auth"
	"github.com/ocenb/music-go/user-service/internal/services/user"
	"github.com/ocenb/music-protos/gen/userservice"
)

type UserServer struct {
	userservice.UnimplementedUserServiceServer
	authService auth.AuthServiceInterface
	userService user.UserServiceInterface
	cfg         *config.Config
}

func NewUserServer(gRPCServer *grpc.Server, cfg *config.Config, authService auth.AuthServiceInterface, userService user.UserServiceInterface) {
	userservice.RegisterUserServiceServer(gRPCServer, &UserServer{authService: authService, userService: userService, cfg: cfg})
}
