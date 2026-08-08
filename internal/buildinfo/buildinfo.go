package buildinfo

import (
	"runtime/debug"
	"strings"
)

var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

type Details struct {
	Version string
	Commit  string
	Date    string
	Dirty   bool
}

func Current() Details {
	details := Details{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return details
	}
	if details.Version == "dev" &&
		info.Main.Version != "" &&
		info.Main.Version != "(devel)" {
		details.Version = info.Main.Version
	}
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			if details.Commit == "" {
				details.Commit = setting.Value
			}
		case "vcs.time":
			if details.Date == "" {
				details.Date = setting.Value
			}
		case "vcs.modified":
			details.Dirty = setting.Value == "true"
		}
	}
	return details
}

func (d Details) String() string {
	version := strings.TrimSpace(d.Version)
	if version == "" {
		version = "dev"
	}
	result := "maisternia " + version
	if d.Dirty {
		result += " (dirty)"
	}
	return result
}
