package output

import "fmt"

type ColorMode string

const (
	ColorAuto   ColorMode = "auto"
	ColorAlways ColorMode = "always"
	ColorNever  ColorMode = "never"
)

type Environment struct {
	NoColor     string
	Term        string
	StderrIsTTY bool
}

type Status string

const (
	StatusOK       Status = "ok"
	StatusInfo     Status = "info"
	StatusWarn     Status = "warn"
	StatusError    Status = "error"
	StatusFallback Status = "fallback"
	StatusHistory  Status = "history"
)

func FormatStatus(status Status, message string, color ColorMode) string {
	label, ansi := statusStyle(status)
	text := fmt.Sprintf("%s %s", label, message)
	if color == ColorNever {
		return text
	}
	return ansi + text + "\x1b[0m"
}

func ResolveColorMode(requested ColorMode, env Environment) ColorMode {
	switch requested {
	case ColorAlways:
		return ColorAlways
	case ColorNever:
		return ColorNever
	}
	if env.NoColor != "" || env.Term == "dumb" || !env.StderrIsTTY {
		return ColorNever
	}
	return ColorAlways
}

func statusStyle(status Status) (string, string) {
	switch status {
	case StatusOK:
		return "[OK]", "\x1b[92m"
	case StatusWarn:
		return "[WARN]", "\x1b[93m"
	case StatusError:
		return "[ERROR]", "\x1b[91m"
	case StatusFallback:
		return "[FALLBACK]", "\x1b[93m"
	case StatusHistory:
		return "[HISTORY]", "\x1b[96m"
	default:
		return "[INFO]", "\x1b[96m"
	}
}
