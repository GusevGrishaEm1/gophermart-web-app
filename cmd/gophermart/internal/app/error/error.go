package error

type CustomError struct {
	Err        error
	HttpStatus int
}

func (err *CustomError) Error() string {
	return err.Err.Error()
}

func NewError(err error, status int) *CustomError {
	return &CustomError{
		Err:        err,
		HttpStatus: status,
	}
}

func NewErrorWithoutStatus(err error) *CustomError {
	return &CustomError{
		Err: err,
	}
}
