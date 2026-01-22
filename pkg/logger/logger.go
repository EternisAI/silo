package logger

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

type Logger struct {
	verbose bool
}

func New(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, msg+"\n", args...)
}

func (l *Logger) Success(msg string, args ...interface{}) {
	green := color.New(color.FgGreen).SprintfFunc()
	fmt.Fprintf(os.Stdout, green("✓ "+msg+"\n", args...))
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	yellow := color.New(color.FgYellow).SprintfFunc()
	fmt.Fprintf(os.Stderr, yellow("⚠ "+msg+"\n", args...))
}

func (l *Logger) Error(msg string, args ...interface{}) {
	red := color.New(color.FgRed).SprintfFunc()
	fmt.Fprintf(os.Stderr, red("✗ "+msg+"\n", args...))
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.verbose {
		cyan := color.New(color.FgCyan).SprintfFunc()
		fmt.Fprintf(os.Stdout, cyan("[DEBUG] "+msg+"\n", args...))
	}
}
