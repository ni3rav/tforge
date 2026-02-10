package cli

import (
	"fmt"
	"io"
)

const (
	prefixInfo  = "[tforge]"
	prefixWarn  = "[tforge:warn]"
	prefixError = "[tforge:error]"
)

func Info(out io.Writer, format string, args ...any) {
	fmt.Fprintf(out, "%s %s\n", prefixInfo, fmt.Sprintf(format, args...))
}

func Warn(out io.Writer, format string, args ...any) {
	fmt.Fprintf(out, "%s %s\n", prefixWarn, fmt.Sprintf(format, args...))
}

func Error(out io.Writer, format string, args ...any) {
	fmt.Fprintf(out, "%s %s\n", prefixError, fmt.Sprintf(format, args...))
}
