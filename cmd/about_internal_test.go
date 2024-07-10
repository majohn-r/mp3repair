/*
 * Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
 */

package cmd

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"testing"
)

func Test_interpretStyle(t *testing.T) {
	tests := map[string]struct {
		flag cmdtoolkit.CommandFlag[string]
		want cmdtoolkit.FlowerBoxStyle
	}{
		"lc ascii": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "ascii"},
			want: cmdtoolkit.ASCIIFlowerBox,
		},
		"uc ascii": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "ASCII"},
			want: cmdtoolkit.ASCIIFlowerBox,
		},
		"lc rounded": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "rounded"},
			want: cmdtoolkit.CurvedFlowerBox,
		},
		"uc rounded": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "ROUNDED"},
			want: cmdtoolkit.CurvedFlowerBox,
		},
		"lc light": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "light"},
			want: cmdtoolkit.LightLinedFlowerBox,
		},
		"uc light": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "LIGHT"},
			want: cmdtoolkit.LightLinedFlowerBox,
		},
		"lc heavy": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "heavy"},
			want: cmdtoolkit.HeavyLinedFlowerBox,
		},
		"uc heavy": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "HEAVY"},
			want: cmdtoolkit.HeavyLinedFlowerBox,
		},
		"lc double": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "double"},
			want: cmdtoolkit.DoubleLinedFlowerBox,
		},
		"uc double": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "DOUBLE"},
			want: cmdtoolkit.DoubleLinedFlowerBox,
		},
		"empty": {
			flag: cmdtoolkit.CommandFlag[string]{Value: ""},
			want: cmdtoolkit.CurvedFlowerBox,
		},
		"garbage": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "abc"},
			want: cmdtoolkit.CurvedFlowerBox,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := interpretStyle(tt.flag); got != tt.want {
				t.Errorf("interpretStyle() = %v, want %v", got, tt.want)
			}
		})
	}
}
