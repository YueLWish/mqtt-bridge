package errors

import (
	stdErrors "errors"
	"fmt"
	"github.com/go-errors/errors"
)

func Is(e error, original error) bool {
	return errors.Is(e, original)
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// WithStack 新增 堆栈信息
func WithStack(err error) error {
	return errors.Wrap(err, 1)
}

// Wrap 附加消息 并 新增 堆栈信息
func Wrap(err error, message string) error {
	return errors.WrapPrefix(err, message, 1)
}

// Wrapf 附加消息 并 新增 堆栈信息
func Wrapf(err error, format string, args ...interface{}) error {
	return errors.WrapPrefix(err, fmt.Sprintf(format, args...), 1)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func New(text string) error {
	return stdErrors.New(text)
}

func Errorf(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

func Join(errs ...error) error {
	return errors.Join(errs...)
}

func WithMessage(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

func WithMessagef(err error, format string, a ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, a...), err)
}

// ErrorStack 输出堆栈信息
func ErrorStack(err error) string {
	var tErr *errors.Error
	if errors.As(err, &tErr) {
		return tErr.ErrorStack()
	} else {
		return fmt.Sprintf("%+v", err)
	}
}
