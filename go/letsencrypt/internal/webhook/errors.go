package webhook

import "fmt"

func newWebhookError(message string, code int, err error) *webhookError {
	return &webhookError{
		message: message,
		code:    code,
		err:     err,
	}
}

type webhookError struct {
	message string
	code    int
	err     error
}

func (e *webhookError) Unwrap() error {
	return e.err
}

func (e *webhookError) Error() string {
	if e.err == nil {
		return e.message
	}

	return fmt.Sprintf("%s: %s", e.message, e.err)
}
