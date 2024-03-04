package error

type CustomError struct {
	Err        error
	HTTPStatus int
}

func (err *CustomError) Error() string {
	return err.Err.Error()
}

func NewError(err error, status int) *CustomError {
	return &CustomError{
		Err:        err,
		HTTPStatus: status,
	}
}

func NewErrorWithoutStatus(err error) *CustomError {
	return &CustomError{
		Err: err,
	}
}
