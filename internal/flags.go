package internal

import "fmt"

// DecorateBoolFlagUsage appends a default value to the provided usage if the
// default value is false. This is a work-around for the flag package's
// defaultUsage function, which displays each flag's usage, along with its
// default value - but it only includes the default value only if the default
// value is not the zero value for the flag's type.
func DecorateBoolFlagUsage(usage string, defaultValue bool) string {
	if defaultValue {
		return usage
	}
	return fmt.Sprintf("%s (default false)", usage)
}

// DecorateIntFlagUsage appends a default value to the provided usage if the
// default value is 0. This is a work-around for the flag package's defaultUsage
// function, which displays each flag's usage, along with its default value -
// but it only includes the default value only if the default value is not the
// zero value for the flag's type.
func DecorateIntFlagUsage(usage string, defaultValue int) string {
	if defaultValue != 0 {
		return usage
	}
	return fmt.Sprintf("%s (default 0)", usage)
}

// DecorateIntFlagUsage appends a default value to the provided usage if the
// default value is the empty string. This is a work-around for the flag
// package's defaultUsage function, which displays each flag's usage, along with
// its default value - but it only includes the default value only if the
// default value is not the zero value for the flag's type.
func DecorateStringFlagUsage(usage string, defaultValue string) string {
	if len(defaultValue) != 0 {
		return usage
	}
	return fmt.Sprintf("%s (default \"\")", usage)
}
