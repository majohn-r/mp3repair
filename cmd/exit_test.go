package cmd

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewExitUserError(t *testing.T) {
	tests := map[string]struct {
		cmd  string
		want *ExitError
	}{
		"typical": {
			cmd:  "someCommand",
			want: &ExitError{command: "someCommand", errorCode: userError},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NewExitUserError(tt.cmd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewExitUserError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewExitSystemError(t *testing.T) {
	tests := map[string]struct {
		cmd  string
		want *ExitError
	}{
		"typical": {
			cmd:  "someCommand",
			want: &ExitError{command: "someCommand", errorCode: systemError},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NewExitSystemError(tt.cmd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewExitSystemError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewExitProgrammingError(t *testing.T) {
	tests := map[string]struct {
		cmd  string
		want *ExitError
	}{
		"typical": {
			cmd:  "someCommand",
			want: &ExitError{command: "someCommand", errorCode: programError},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NewExitProgrammingError(tt.cmd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewExitProgrammingError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExitError_Error(t *testing.T) {
	tests := map[string]struct {
		e    *ExitError
		want string
	}{
		"user error": {
			e:    NewExitUserError("command1"),
			want: `command "command1" terminated with an error: user error`,
		},
		"programming error": {
			e:    NewExitProgrammingError("command2"),
			want: `command "command2" terminated with an error: programming error`,
		},
		"system error": {
			e:    NewExitSystemError("command3"),
			want: `command "command3" terminated with an error: system call failed`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("ExitError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExitError_status(t *testing.T) {
	tests := map[string]struct {
		e    *ExitError
		want int
	}{
		"user error":        {e: NewExitUserError("command1"), want: 1},
		"programming error": {e: NewExitProgrammingError("command2"), want: 2},
		"system error":      {e: NewExitSystemError("command3"), want: 3},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.e.Status(); got != tt.want {
				t.Errorf("ExitError.Status() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToErrInterface(t *testing.T) {
	tests := map[string]struct {
		e       *ExitError
		wantErr bool
	}{
		"user error":        {e: NewExitUserError("cmd"), wantErr: true},
		"programming error": {e: NewExitProgrammingError("cmd"), wantErr: true},
		"system error":      {e: NewExitSystemError("cmd"), wantErr: true},
		"no error":          {e: nil, wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotErr := ToErrorInterface(tt.e)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("ToErrorInterface() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
			if gotErr == nil {
				var exitError *ExitError
				if errors.As(gotErr, &exitError) {
					t.Errorf("ToErrorInterface() returned nil that is *ExitError")
				}
			} else if !reflect.DeepEqual(gotErr, tt.e) {
				t.Errorf("ToErrorInterface() got %v want %v", gotErr, tt.e)
			}
		})
	}
}
