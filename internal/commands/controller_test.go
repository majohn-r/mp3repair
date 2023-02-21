package commands

import "testing"

func TestIsDefault(t *testing.T) {
	type args struct {
		commandName string
	}
	tests := map[string]struct {
		defaultCommand string
		args
		want bool
	}{
		"miss": {args: args{commandName: "bar"}, want: false},
		"hit":  {args: args{commandName: "list"}, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := IsDefault(tt.args.commandName); got != tt.want {
				t.Errorf("IsDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tests := map[string]struct {
	}{"dummy": {}}
	for name := range tests {
		t.Run(name, func(t *testing.T) {
			Load()
		})
	}
}
