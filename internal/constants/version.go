package constants

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var versionFile string

// Version is the application version loaded from the VERSION file.
var Version = strings.TrimSpace(versionFile)
