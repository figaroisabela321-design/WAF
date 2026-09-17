package httpx

import "errors"

// Common application error codes.
const (
	CodeOK              = 0
	CodeBadRequest      = 40000
	CodeUnauthorized    = 40100
	CodeForbidden       = 40300
	CodeNotFound        = 40400
	CodeConflict        = 40900
	CodeInternal        = 50000
	CodeValidation      = 40001
	CodePayloadTooLarge = 41300
)

// AppError is a typed application error with HTTP status and business code.
type AppError struct {
	HTTPStatus int
	Code       int
	Message    string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func BadRequest(msg string) *AppError {
	return &AppError{HTTPStatus: 400, Code: CodeBadRequest, Message: msg}
}

func Validation(msg string) *AppError {
	return &AppError{HTTPStatus: 400, Code: CodeValidation, Message: msg}
}

func Unauthorized(msg string) *AppError {
	return &AppError{HTTPStatus: 401, Code: CodeUnauthorized, Message: msg}
}

func Forbidden(msg string) *AppError {
	return &AppError{HTTPStatus: 403, Code: CodeForbidden, Message: msg}
}

func NotFound(msg string) *AppError {
	return &AppError{HTTPStatus: 404, Code: CodeNotFound, Message: msg}
}

func Conflict(msg string) *AppError {
	return &AppError{HTTPStatus: 409, Code: CodeConflict, Message: msg}
}

func Internal(msg string, err error) *AppError {
	return &AppError{HTTPStatus: 500, Code: CodeInternal, Message: msg, Err: err}
}

func PayloadTooLarge(msg string) *AppError {
	return &AppError{HTTPStatus: 413, Code: CodePayloadTooLarge, Message: msg}
}

// AsAppError extracts *AppError if present.
func AsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
