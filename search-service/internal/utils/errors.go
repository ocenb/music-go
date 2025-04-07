package utils

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	InternalError = func(err error, msg string) error {
		return status.Errorf(codes.Internal, "%s: %v", msg, err)
	}

	AlreadyExistsError = func(err error) error {
		return status.Errorf(codes.AlreadyExists, "%s", err.Error())
	}

	UnauthenticatedError = func(msg string) error {
		return status.Errorf(codes.Unauthenticated, "%s", msg)
	}

	NotFoundError = func(err error) error {
		return status.Errorf(codes.NotFound, "%s", err.Error())
	}
)
