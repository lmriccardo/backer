package version

import (
	"fmt"
	"strings"
)

const LATEST_DB_VERSION = 2 // The latest db version
const API_VERSION = 1       // Current version of the API

var (
	// These variables are overridden via -ldflags at compile time
	Version   = "dev"
	Commit    = "none"
	Date      = "unknown"
	GoVersion = "unknown"
)

// @name VersionInfo
type Info struct {
	Version    string `json:"version"`
	Commit     string `json:"commit"`
	Date       string `json:"date"`
	GoVersion  string `json:"go_version"`
	ApiVersion int    `json:"api"`
}

func Get() Info {
	return Info{
		Version:    Version,
		Commit:     Commit,
		Date:       Date,
		GoVersion:  GoVersion,
		ApiVersion: API_VERSION,
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
