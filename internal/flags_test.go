package internal

import "testing"

func TestDecorateBoolFlagUsage(t *testing.T) {
	const fnName = "DecorateBoolFlagUsage()"
	type args struct {
		usage        string
		defaultValue bool
	}
	tests := map[string]struct {
		args
		want string
	}{
		"true default value":  {args: args{usage: "this is my boolean flag", defaultValue: true}, want: "this is my boolean flag"},
		"false default value": {args: args{usage: "this is my boolean flag", defaultValue: false}, want: "this is my boolean flag (default false)"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := DecorateBoolFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestDecorateIntFlagUsage(t *testing.T) {
	const fnName = "DecorateIntFlagUsage()"
	type args struct {
		usage        string
		defaultValue int
	}
	tests := map[string]struct {
		args
		want string
	}{
		"non-zero value": {args: args{usage: "this is my int flag", defaultValue: 1}, want: "this is my int flag"},
		"zero value":     {args: args{usage: "this is my int flag", defaultValue: 0}, want: "this is my int flag (default 0)"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := DecorateIntFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestDecorateStringFlagUsage(t *testing.T) {
	const fnName = "DecorateStringFlagUsage()"
	type args struct {
		usage        string
		defaultValue string
	}
	tests := map[string]struct {
		args
		want string
	}{
		"non-empty value": {args: args{usage: "this is my string flag", defaultValue: "foo"}, want: "this is my string flag"},
		"empty value":     {args: args{usage: "this is my string flag", defaultValue: ""}, want: "this is my string flag (default \"\")"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := DecorateStringFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}
