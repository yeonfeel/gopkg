package er

import (
	"fmt"
	"runtime"
)

type errors struct {
	message  string
	err      error
	location string
}

// Error returns a error text
func (e *errors) Error() (text string) {
	if e.message != "" {
		text += fmt.Sprintf("[ERROR] %s", e.message)
	}

	if e.location != "" {
		if text != "" {
			text += "\n"
		}
		text += fmt.Sprintf("[LINE ] %s", e.location)
	}

	if e.err != nil {
		if text != "" {
			text += "\n"
		}
		text += fmt.Sprintf("[CAUSE] %s", e.err)
	}

	return text
}

// Error returns a new error
func Error(err error, a ...interface{}) error {
	return newError(err, 2, a...)
}

// ErrorS returns a new error with formatted text
func ErrorS(format string, a ...interface{}) error {
	return newError(nil, 2, fmt.Sprintf(format, a...))
}

// ErrorF returns a new error with the cause error
func ErrorF(err error, format string, a ...interface{}) error {
	return newError(err, 2, fmt.Sprintf(format, a...))
}

func newError(err error, skip int, a ...interface{}) error {
	var location string
	if _, fname, line, ok := runtime.Caller(skip); ok {
		location = fmt.Sprintf("%s:%d", fname, line)
	}

	return &errors{
		message:  fmt.Sprint(a...),
		err:      err,
		location: location,
	}
}

// Recover returns a recovered error
func Recover(err error) error {
	if r := recover(); r != nil {
		if s, ok := r.(string); ok {
			err = fmt.Errorf(s)
		} else if e, ok := r.(error); ok {
			err = e
		} else {
			err = fmt.Errorf("%#v", r)
		}
	}

	return err
}
