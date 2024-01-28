/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/

package main

import (
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
