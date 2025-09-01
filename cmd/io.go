/*
 * Copyright © 2025 Marc Johnson (marc.johnson27591@gmail.com)
 */

package cmd

import (
	"fmt"
	"math"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

const (
	ioOpenFileLimit     = "maxOpenFiles"
	ioOpenFileLimitFlag = "--" + ioOpenFileLimit
	ioUsage             = "[" + ioOpenFileLimitFlag + " count]"
	ioOpenFileMinimum   = 1
	ioOpenFileDefault   = 1000
	ioOpenFileMaximum   = math.MaxInt16
)

var (
	ioFileLimitBounds = cmdtoolkit.NewIntBounds(ioOpenFileMinimum, ioOpenFileDefault, ioOpenFileMaximum)
	ioFlags           = &cmdtoolkit.FlagSet{
		Name: "io",
		Details: map[string]*cmdtoolkit.FlagDetails{
			ioOpenFileLimit: {
				Usage: fmt.Sprintf(
					"the maximum number of files that can be read simultaneously (at least %d, at most %d, default %d)",
					ioOpenFileMinimum, ioOpenFileMaximum, ioOpenFileDefault),
				ExpectedType: cmdtoolkit.IntType,
				DefaultValue: ioFileLimitBounds},
		},
	}
)

type ioSettings struct {
	openFileLimit int
}

func evaluateIOFlags(o output.Bus, producer cmdtoolkit.FlagProducer) (*ioSettings, bool) {
	values, eSlice := cmdtoolkit.ReadFlags(producer, ioFlags)
	if cmdtoolkit.ProcessFlagErrors(o, eSlice) {
		return processIOFlags(o, values)
	}
	return &ioSettings{}, false
}

func processIOFlags(o output.Bus, values map[string]*cmdtoolkit.CommandFlag[any]) (*ioSettings, bool) {
	value := &ioSettings{}
	rawValue, flagErr := cmdtoolkit.GetInt(o, values, ioOpenFileLimit)
	if flagErr != nil {
		return value, false
	}
	value.openFileLimit = constrainBoundedValue(o, ioOpenFileLimitFlag, rawValue.Value, ioFileLimitBounds)
	return value, true
}

func init() {
	cmdtoolkit.AddDefaults(ioFlags)
}
