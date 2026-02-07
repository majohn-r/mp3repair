/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/

package main

import (
	"mp3repair/cmd"
	"reflect"
	"testing"
)

func Test_main(t *testing.T) {
	originalExecutor := executor
	defer func() {
		executor = originalExecutor
	}()
	executed := false
	executor = func() { executed = true }
	tests := map[string]struct {
		wantExecuted bool
	}{
		"good": {wantExecuted: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			executed = false
			main()
			if executed != tt.wantExecuted {
				t.Errorf("main() got %t want %t", executed, tt.wantExecuted)
			}
		})
	}
}

func Test_isSelfPromoted(t *testing.T) {
	tests := map[string]struct {
		args  []string
		want  []string
		want1 bool
	}{
		"null args": {
			args:  nil,
			want:  nil,
			want1: false,
		},
		"empty args": {
			args:  []string{},
			want:  []string{},
			want1: false,
		},
		"few args": {
			args:  []string{"./mp3repair.exe"},
			want:  []string{"./mp3repair.exe"},
			want1: false,
		},
		"multiple args, no self promotion": {
			args:  []string{"./mp3repair.exe", "scan", "-f"},
			want:  []string{"./mp3repair.exe", "scan", "-f"},
			want1: false,
		},
		"minimal args with self-promotion": {
			args:  []string{"./mp3repair.exe", selfPromotionMarker},
			want:  []string{"./mp3repair.exe"},
			want1: true,
		},
		"more args with self-promotion": {
			args:  []string{"./mp3repair.exe", selfPromotionMarker, "scan", "-f"},
			want:  []string{"./mp3repair.exe", "scan", "-f"},
			want1: true,
		},
		"misplaced self-promotion": {
			args:  []string{"./mp3repair.exe", "scan", selfPromotionMarker, "-f"},
			want:  []string{"./mp3repair.exe", "scan", selfPromotionMarker, "-f"},
			want1: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, got1 := isSelfPromoted(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("isSelfPromoted() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("isSelfPromoted() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_injectSelfPromotion(t *testing.T) {
	tests := map[string]struct {
		args []string
		want []string
	}{
		"null":  {args: nil, want: nil},
		"empty": {args: []string{}, want: []string{}},
		"single arg": {
			args: []string{"./mp3repair.exe"},
			want: []string{"./mp3repair.exe", selfPromotionMarker},
		},
		"multiple args": {
			args: []string{"./mp3repair.exe", "scan", "-f"},
			want: []string{"./mp3repair.exe", selfPromotionMarker, "scan", "-f"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := injectSelfPromotion(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("injectSelfPromotion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_configureExit(t *testing.T) {
	originalScanf := scanf
	originalExit := cmd.Exit
	defer func() {
		scanf = originalScanf
		cmd.Exit = originalExit
	}()
	var scanfInvoked bool
	scanf = func(string, ...any) (int, error) {
		scanfInvoked = true
		return 0, nil
	}
	var exitInvoked bool
	exitFunction := func(int) {
		exitInvoked = true
	}
	tests := map[string]struct {
		selfPromoted bool
	}{
		"promoted":     {selfPromoted: true},
		"not promoted": {selfPromoted: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			scanfInvoked = false
			exitInvoked = false
			cmd.Exit = exitFunction
			configureExit(tt.selfPromoted)
			cmd.Exit(0)
			if !exitInvoked {
				t.Errorf("configureExit() exit not invoked")
			}
			if tt.selfPromoted {
				if !scanfInvoked {
					t.Errorf("configureExit() scanf not invoked")
				}
			} else {
				if scanfInvoked {
					t.Errorf("configureExit() scanf erroneously invoked")
				}
			}
		})
	}
}
