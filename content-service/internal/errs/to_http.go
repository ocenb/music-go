package errs

import "net/http"

func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	domainErr, ok := As(err)
	if !ok {
		return http.StatusInternalServerError
	}

	switch domainErr.Kind() {
	case KindInternal:
		return http.StatusInternalServerError
	case KindNotFound:
		return http.StatusNotFound
	case KindInvalidArgument:
		return http.StatusBadRequest
	case KindAlreadyExists:
		return http.StatusConflict
	case KindUnauthenticated:
		return http.StatusUnauthorized
	case KindPermissionDenied:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func HTTPMessage(err error) string {
	if err == nil {
		return ""
	}

	domainErr, ok := As(err)
	if !ok {
		return "internal error"
	}

	if domainErr.Message() != "" {
		return domainErr.Message()
	}

	return domainErr.Error()
}
