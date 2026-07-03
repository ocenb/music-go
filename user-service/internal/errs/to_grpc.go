package errs

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToGRPC(err error) error {
	if err == nil {
		return nil
	}

	domainErr, ok := As(err)
	if !ok {
		return status.Error(codes.Internal, "internal error")
	}

	msg := domainErr.Message()
	if msg == "" {
		msg = domainErr.Error()
	}

	switch domainErr.Kind() {
	case KindInternal:
		return status.Error(codes.Internal, msg)
	case KindNotFound:
		return status.Error(codes.NotFound, msg)
	case KindInvalidArgument:
		return status.Error(codes.InvalidArgument, msg)
	case KindAlreadyExists:
		return status.Error(codes.AlreadyExists, msg)
	case KindUnauthenticated:
		return status.Error(codes.Unauthenticated, msg)
	case KindPermissionDenied:
		return status.Error(codes.PermissionDenied, msg)
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
