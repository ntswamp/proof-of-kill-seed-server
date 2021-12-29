package lib_error

import (
	"fmt"
)

type SaleError struct {
	ErrorType int
	FuncName  string
	Message   string
}

func (err *SaleError) Error() string {
	return fmt.Sprintf("%s(%v)", err.Message, err.ErrorType)
}

func NewSaleError(errorType int, message string, args ...interface{}) error {
	if 0 < len(args) {
		message = fmt.Sprintf(message, args...)
	}
	return &SaleError{
		ErrorType: errorType,
		Message:   message,
	}
}
