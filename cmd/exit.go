package cmd

import (
	"fmt"
	"reflect"
)

type errorCode int

const (
	unknown      errorCode = iota
	userError              // user did something silly
	programError           // program code error
	systemError            // unexpected errors, like file not found
)

var (
	strStatusMap = map[errorCode]string{
		userError:    "user error",
		programError: "programming error",
		systemError:  "system call failed",
	}
)

type ExitError struct {
	errorCode
	command string
}

func NewExitUserError(cmd string) *ExitError {
	return &ExitError{command: cmd, errorCode: userError}
}

func NewExitProgrammingError(cmd string) *ExitError {
	return &ExitError{command: cmd, errorCode: programError}
}

func NewExitSystemError(cmd string) *ExitError {
	return &ExitError{command: cmd, errorCode: systemError}
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("command %q terminated with an error: %s", e.command, strStatusMap[e.errorCode])
}

func (e *ExitError) Status() int {
	return int(e.errorCode)
}

func ToErrorInterface(e *ExitError) (err error) {
	if reflect.ValueOf(e).IsNil() {
		err = nil
	} else {
		err = e
	}
	return
}
