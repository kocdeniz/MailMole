package buildinfo

import "strings"

// These values are set at build time with -ldflags -X.
var (
	Version   = "1.2.0"
	BuildDate = "unknown"
)

func IntroLabel() string {
	v := strings.TrimSpace(Version)
	if v == "" {
		v = "dev"
	}
	b := strings.TrimSpace(BuildDate)
	if b == "" || b == "unknown" {
		return "MailMole v" + v
	}
	return "MailMole v" + v + "  build " + b
}
