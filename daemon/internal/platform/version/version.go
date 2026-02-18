package version

import (
	"fmt"
	"strings"
)

var (
	// These variables are overridden via -ldflags at compile time
	Version   = "dev"
	Commit    = "none"
	Date      = "unknown"
	GoVersion = "unknown"
)

type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
}

func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: GoVersion,
	}
}

func (i Info) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "backerd %s ", i.Version)
	fmt.Fprintf(&sb, "(%s) ", i.Commit)
	fmt.Fprintf(&sb, "built %s ", i.Date)
	fmt.Fprintf(&sb, "go=%s", i.GoVersion)
	return sb.String()
}
