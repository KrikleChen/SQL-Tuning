package output

import (
	"strings"
	"testing"
)

func TestStatusMessageNeverUsesBlackOrDarkGray(t *testing.T) {
	msg := FormatStatus(StatusInfo, "history hit", ColorAlways)
	if strings.Contains(msg, "\x1b[30m") || strings.Contains(msg, "\x1b[90m") {
		t.Fatalf("status message uses unreadable dark ANSI color: %q", msg)
	}
}

func TestStatusMessageIncludesTextLabel(t *testing.T) {
	msg := FormatStatus(StatusHistory, "reused optimized SQL", ColorNever)
	if !strings.Contains(msg, "[HISTORY]") {
		t.Fatalf("expected history label, got %q", msg)
	}
}

func TestStatusMessagesIncludeReadableLabels(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		label  string
	}{
		{name: "info", status: StatusInfo, label: "[INFO]"},
		{name: "history", status: StatusHistory, label: "[HISTORY]"},
		{name: "error", status: StatusError, label: "[ERROR]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := FormatStatus(tt.status, "status message", ColorNever)
			if !strings.Contains(msg, tt.label) {
				t.Fatalf("expected %s label, got %q", tt.label, msg)
			}
		})
	}
}

func TestColorNeverDisablesANSI(t *testing.T) {
	msg := FormatStatus(StatusError, "failed", ColorNever)
	if strings.Contains(msg, "\x1b[") {
		t.Fatalf("expected no ANSI color when color disabled, got %q", msg)
	}
}

func TestResolveColorAutoDisablesANSIForNoColor(t *testing.T) {
	got := ResolveColorMode(ColorAuto, Environment{
		NoColor:     "1",
		Term:        "xterm-256color",
		StderrIsTTY: true,
	})
	if got != ColorNever {
		t.Fatalf("got %q, want %q", got, ColorNever)
	}
}

func TestResolveColorAutoDisablesANSIForDumbTerminal(t *testing.T) {
	got := ResolveColorMode(ColorAuto, Environment{
		Term:        "dumb",
		StderrIsTTY: true,
	})
	if got != ColorNever {
		t.Fatalf("got %q, want %q", got, ColorNever)
	}
}

func TestResolveColorAutoDisablesANSIForNonTTY(t *testing.T) {
	got := ResolveColorMode(ColorAuto, Environment{
		Term:        "xterm-256color",
		StderrIsTTY: false,
	})
	if got != ColorNever {
		t.Fatalf("got %q, want %q", got, ColorNever)
	}
}

func TestResolveColorAlwaysOverridesEnvironment(t *testing.T) {
	got := ResolveColorMode(ColorAlways, Environment{
		NoColor:     "1",
		Term:        "dumb",
		StderrIsTTY: false,
	})
	if got != ColorAlways {
		t.Fatalf("got %q, want %q", got, ColorAlways)
	}
}
