package utils

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	InternalError = func(err error, msg string) error {
		return status.Errorf(codes.Internal, "%s: %v", msg, err)
	}

	NotFoundError = func(msg string) error {
		return status.Errorf(codes.NotFound, "%s", msg)
	}

	InvalidArgumentError = func(msg string) error {
		return status.Errorf(codes.InvalidArgument, "%s", msg)
	}

	AlreadyExistsError = func(msg string) error {
		return status.Errorf(codes.AlreadyExists, "%s", msg)
	}

	UnauthenticatedError = func(msg string) error {
		return status.Errorf(codes.Unauthenticated, "%s", msg)
	}

	PermissionDeniedError = func(msg string) error {
		return status.Errorf(codes.PermissionDenied, "%s", msg)
	}
)
