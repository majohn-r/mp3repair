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
		flag CommandFlag[string]
		want cmdtoolkit.FlowerBoxStyle
	}{
		"lc ascii": {
			flag: CommandFlag[string]{Value: "ascii"},
			want: cmdtoolkit.ASCIIFlowerBox,
		},
		"uc ascii": {
			flag: CommandFlag[string]{Value: "ASCII"},
			want: cmdtoolkit.ASCIIFlowerBox,
		},
		"lc rounded": {
			flag: CommandFlag[string]{Value: "rounded"},
			want: cmdtoolkit.CurvedFlowerBox,
		},
		"uc rounded": {
			flag: CommandFlag[string]{Value: "ROUNDED"},
			want: cmdtoolkit.CurvedFlowerBox,
		},
		"lc light": {
			flag: CommandFlag[string]{Value: "light"},
			want: cmdtoolkit.LightLinedFlowerBox,
		},
		"uc light": {
			flag: CommandFlag[string]{Value: "LIGHT"},
			want: cmdtoolkit.LightLinedFlowerBox,
		},
		"lc heavy": {
			flag: CommandFlag[string]{Value: "heavy"},
			want: cmdtoolkit.HeavyLinedFlowerBox,
		},
		"uc heavy": {
			flag: CommandFlag[string]{Value: "HEAVY"},
			want: cmdtoolkit.HeavyLinedFlowerBox,
		},
		"lc double": {
			flag: CommandFlag[string]{Value: "double"},
			want: cmdtoolkit.DoubleLinedFlowerBox,
		},
		"uc double": {
			flag: CommandFlag[string]{Value: "DOUBLE"},
			want: cmdtoolkit.DoubleLinedFlowerBox,
		},
		"empty": {
			flag: CommandFlag[string]{Value: ""},
			want: cmdtoolkit.CurvedFlowerBox,
		},
		"garbage": {
			flag: CommandFlag[string]{Value: "abc"},
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
