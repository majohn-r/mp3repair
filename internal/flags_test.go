package internal

import "testing"

func TestDecorateBoolFlagUsage(t *testing.T) {
	fnName := "DecorateBoolFlagUsage()"
	type args struct {
		usage        string
		defaultValue bool
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "non-default value",
			args: args{
				usage:        "this is my boolean flag",
				defaultValue: true,
			},
			want: "this is my boolean flag",
		},
		{
			name: "non-default value",
			args: args{
				usage:        "this is my boolean flag",
				defaultValue: false,
			},
			want: "this is my boolean flag (default false)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecorateBoolFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestDecorateIntFlagUsage(t *testing.T) {
	fnName := "DecorateIntFlagUsage()"
	type args struct {
		usage        string
		defaultValue int
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "non-default value",
			args: args{
				usage:        "this is my int flag",
				defaultValue: 1,
			},
			want: "this is my int flag",
		},
		{
			name: "non-default value",
			args: args{
				usage:        "this is my int flag",
				defaultValue: 0,
			},
			want: "this is my int flag (default 0)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecorateIntFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestDecorateStringFlagUsage(t *testing.T) {
	fnName := "DecorateStringFlagUsage()"
	type args struct {
		usage        string
		defaultValue string
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "non-default value",
			args: args{
				usage:        "this is my string flag",
				defaultValue: "foo",
			},
			want: "this is my string flag",
		},
		{
			name: "non-default value",
			args: args{
				usage:        "this is my string flag",
				defaultValue: "",
			},
			want: "this is my string flag (default \"\")",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecorateStringFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}
