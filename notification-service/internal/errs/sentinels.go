package errs

var (
	ErrInvalidNotification = InvalidArgument("invalid notification message")
	ErrSendFailed          = Internal(nil, "failed to send email notification after retries")
)
