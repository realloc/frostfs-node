package internal

import (
	"strings"
)

// EnvPrefix is a prefix of ENV variables related
// to storage node configuration.
const EnvPrefix = "FROSTFS"

// EnvSeparator is a section separator in ENV variables.
const EnvSeparator = "_"

// Env returns ENV variable key for a particular config parameter.
func Env(path ...string) string {
	return strings.ToUpper(
		strings.Join(
			append([]string{EnvPrefix}, path...),
			EnvSeparator,
		),
	)
}
