package commands

import "testing"

func TestDeclareDefault(t *testing.T) {
	savedDefaultCommand := defaultCommand
	defer func() {
		defaultCommand = savedDefaultCommand
	}()
	type args struct {
		s string
	}
	tests := map[string]struct {
		args
	}{"basic": {args: args{s: defaultCommand + "_foo"}}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			DeclareDefault(tt.args.s)
			if got := defaultCommand; got != tt.args.s {
				t.Errorf("DeclareDefault got %q want %q", got, tt.args.s)
			}
		})
	}
}

func TestIsDefault(t *testing.T) {
	savedDefaultCommand := defaultCommand
	defer func() {
		defaultCommand = savedDefaultCommand
	}()
	type args struct {
		commandName string
	}
	tests := map[string]struct {
		defaultCommand string
		args
		want bool
	}{
		"miss": {defaultCommand: "foo", args: args{commandName: "bar"}, want: false},
		"hit":  {defaultCommand: "bar", args: args{commandName: "bar"}, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defaultCommand = tt.defaultCommand
			if got := IsDefault(tt.args.commandName); got != tt.want {
				t.Errorf("IsDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
