package main

import (
	"errors"
	"runtime/debug"
)

var version = ""

var errVersionShown = errors.New("version shown")

func currentVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}
