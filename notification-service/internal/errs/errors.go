package errs

type Kind int

const (
	KindInternal Kind = iota
	KindInvalidArgument
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

func InvalidArgument(msg string) *Error {
	return &Error{kind: KindInvalidArgument, msg: msg}
}
