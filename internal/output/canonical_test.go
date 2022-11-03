package output

import "testing"

func Test_canonicalFormat(t *testing.T) {
	type args struct {
		format string
		a      []any
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "broad test",
			args: args{
				format: "test format %d %q %v..?!..?\n\n\n\n",
				a:      []any{25, "foo", 1.245},
			},
			want: "Test format 25 \"foo\" 1.245?\n",
		},
		{
			name: "narrow test",
			args: args{
				format: "1. test format %d %q %v",
				a:      []any{25, "foo", 1.245},
			},
			want: "1. test format 25 \"foo\" 1.245.\n",
		},
		{
			name: "nothing but newlines",
			args: args{
				format: "\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
			},
			want: "\n",
		},
		{
			name: "nothing but terminal punctuation",
			args: args{
				format: "!!?.!?.",
			},
			want: ".\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canonicalFormat(tt.args.format, tt.args.a...); got != tt.want {
				t.Errorf("canonicalFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_renderCanonical(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "broad test",
			args: args{
				s: "test format 25 \"foo\" 1.245..?!..?\n\n\n\n",
			},
			want: "Test format 25 \"foo\" 1.245?\n",
		},
		{
			name: "narrow test",
			args: args{
				s: "1. test format 25 \"foo\" 1.245",
			},
			want: "1. test format 25 \"foo\" 1.245.\n",
		},
		{
			name: "nothing but newlines",
			args: args{
				s: "\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
			},
			want: "\n",
		},
		{
			name: "nothing but terminal punctuation",
			args: args{
				s: "!!?.!?.",
			},
			want: ".\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renderCanonical(tt.args.s); got != tt.want {
				t.Errorf("renderCanonical() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fixTerminalPunctuation(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "broad test",
			args: args{
				s: "test format 25 \"foo\" 1.245..?!..?",
			},
			want: "test format 25 \"foo\" 1.245?",
		},
		{
			name: "narrow test",
			args: args{
				s: "1. test format 25 \"foo\" 1.245",
			},
			want: "1. test format 25 \"foo\" 1.245.",
		},
		{
			name: "nothing",
			args: args{
				s: "",
			},
			want: "",
		},
		{
			name: "nothing but terminal punctuation",
			args: args{
				s: "!!?.!?.",
			},
			want: ".",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixTerminalPunctuation(tt.args.s); got != tt.want {
				t.Errorf("fixTerminalPunctuation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stripTrailingNewlines(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "broad test",
			args: args{
				s: "test format 25 \"foo\" 1.245..?!..?\n\n\n\n",
			},
			want: "test format 25 \"foo\" 1.245..?!..?",
		},
		{
			name: "narrow test",
			args: args{
				s: "1. test format 25 \"foo\" 1.245",
			},
			want: "1. test format 25 \"foo\" 1.245",
		},
		{
			name: "nothing but newlines",
			args: args{
				s: "\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
			},
			want: "",
		},
		{
			name: "nothing",
			args: args{
				s: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripTrailingNewlines(tt.args.s); got != tt.want {
				t.Errorf("stripTrailingNewlines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_capitalize(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "broad test",
			args: args{
				s: "test format 25 \"foo\" 1.245..?!..?\n\n\n\n",
			},
			want: "Test format 25 \"foo\" 1.245..?!..?\n\n\n\n",
		},
		{
			name: "narrow test",
			args: args{
				s: "1. test format 25 \"foo\" 1.245",
			},
			want: "1. test format 25 \"foo\" 1.245",
		},
		{
			name: "nothing but newlines",
			args: args{
				s: "\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
			},
			want: "\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
		},
		{
			name: "nothing",
			args: args{
				s: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := capitalize(tt.args.s); got != tt.want {
				t.Errorf("capitalize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isSentenceTerminatingPunctuation(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args
		want bool
	}{
		{name: "period", args: args{s: "."}, want: true},
		{name: "question mark", args: args{s: "?"}, want: true},
		{name: "exclamation point", args: args{s: "!"}, want: true},
		{name: "comma", args: args{s: ","}, want: false},
		{name: "semicolon", args: args{s: ";"}, want: false},
		{name: "colon", args: args{s: ":"}, want: false},
		{name: "en dash", args: args{s: "\u2013"}, want: false},
		{name: "em dash", args: args{s: "—"}, want: false},
		{name: "hyphen", args: args{s: "-"}, want: false},
		{name: "left parenthesis", args: args{s: "("}, want: false},
		{name: "right parenthesis", args: args{s: ")"}, want: false},
		{name: "left bracket", args: args{s: "["}, want: false},
		{name: "right bracket", args: args{s: "]"}, want: false},
		{name: "left brace", args: args{s: "{"}, want: false},
		{name: "right brace", args: args{s: "}"}, want: false},
		{name: "apostrophe", args: args{s: "'"}, want: false},
		{name: "plain quotation mark", args: args{s: "\""}, want: false},
		{name: "ellipsis", args: args{s: "…"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSentenceTerminatingPunctuation(tt.args.s); got != tt.want {
				t.Errorf("isSentenceTerminatingPunctuation() = %v, want %v", got, tt.want)
			}
		})
	}
}
