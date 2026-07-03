package errs

import "errors"

type Kind int

const (
	KindInternal Kind = iota
	KindNotFound
	KindInvalidArgument
	KindAlreadyExists
	KindUnauthenticated
	KindPermissionDenied
)

type Error struct {
	kind Kind
	msg  string
	err  error
}

func (e *Error) Error() string {
	if e.msg != "" && e.err != nil {
		return e.msg + ": " + e.err.Error()
	}
	if e.msg != "" {
		return e.msg
	}
	if e.err != nil {
		return e.err.Error()
	}
	return "error"
}

func (e *Error) Unwrap() error { return e.err }

func (e *Error) Kind() Kind { return e.kind }

func (e *Error) Message() string { return e.msg }

func (e *Error) Is(target error) bool {
	te, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.kind == te.kind && e.msg == te.msg
}

func Internal(cause error, msg string) *Error {
	return &Error{kind: KindInternal, msg: msg, err: cause}
}

func NotFound(msg string) *Error {
	return &Error{kind: KindNotFound, msg: msg}
}

func InvalidArgument(msg string) *Error {
	return &Error{kind: KindInvalidArgument, msg: msg}
}

func AlreadyExists(msg string) *Error {
	return &Error{kind: KindAlreadyExists, msg: msg}
}

func Unauthenticated(msg string) *Error {
	return &Error{kind: KindUnauthenticated, msg: msg}
}

func PermissionDenied(msg string) *Error {
	return &Error{kind: KindPermissionDenied, msg: msg}
}

func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
