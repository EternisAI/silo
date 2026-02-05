package logger

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

type Logger struct {
	verbose bool
	silent  bool
}

func New(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

func NewSilent() *Logger {
	return &Logger{silent: true}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	if l.silent {
		return
	}
	fmt.Fprintf(os.Stdout, msg+"\n", args...)
}

func (l *Logger) Success(msg string, args ...interface{}) {
	if l.silent {
		return
	}
	green := color.New(color.FgGreen).SprintfFunc()
	fmt.Fprint(os.Stdout, green("✓ "+msg+"\n", args...))
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	if l.silent {
		return
	}
	yellow := color.New(color.FgYellow).SprintfFunc()
	fmt.Fprint(os.Stderr, yellow("⚠ "+msg+"\n", args...))
}

func (l *Logger) Error(msg string, args ...interface{}) {
	if l.silent {
		return
	}
	red := color.New(color.FgRed).SprintfFunc()
	fmt.Fprint(os.Stderr, red("✗ "+msg+"\n", args...))
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.silent {
		return
	}
	if l.verbose {
		cyan := color.New(color.FgCyan).SprintfFunc()
		fmt.Fprint(os.Stdout, cyan("[DEBUG] "+msg+"\n", args...))
	}
}
