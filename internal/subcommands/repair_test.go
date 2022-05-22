package subcommands

import (
	"flag"
	"mp3/internal"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func Test_newRepairSubCommand(t *testing.T) {
	topDir := "loadTest"
	fnName := "newRepairSubCommand()"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, topDir, err)
	}
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %s: %v", fnName, topDir, err)
	}
	if err := internal.CreateDefaultYamlFile(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
		internal.DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	type args struct {
		v *viper.Viper
	}
	tests := []struct {
		name       string
		args       args
		wantDryRun bool
	}{
		{
			name:       "ordinary defaults",
			args:       args{v: nil},
			wantDryRun: false,
		},
		{
			name:       "overridden defaults",
			args:       args{v: internal.ReadDefaultsYaml("./mp3")},
			wantDryRun: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repair := newRepairSubCommand(tt.args.v, flag.NewFlagSet("ls", flag.ContinueOnError))
			if s := repair.sf.ProcessArgs(os.Stdout, []string{"-topDir", topDir, "-ext", ".mp3"}); s != nil {
				if *repair.dryRun != tt.wantDryRun {
					t.Errorf("%s %s: got dryRun %t want %t", fnName, tt.name, *repair.dryRun, tt.wantDryRun)
				}
			} else {
				t.Errorf("%s %s: error processing arguments", fnName, tt.name)
			}
		})
	}
}
