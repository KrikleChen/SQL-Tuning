package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/KrikleChen/SQL-Tuning/internal/output"
)

func main() {
	color := flag.String("color", string(output.ColorAuto), "color output: auto, always, never")
	flag.Parse()

	mode := output.ResolveColorMode(output.ColorMode(*color), output.Environment{
		NoColor:     os.Getenv("NO_COLOR"),
		Term:        os.Getenv("TERM"),
		StderrIsTTY: isTerminal(os.Stderr),
	})
	fmt.Fprintln(os.Stderr, output.FormatStatus(output.StatusInfo, "sqlopt framework initialized", mode))
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
