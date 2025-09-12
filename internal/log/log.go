package log

import (
	"log"
	"os"
)

var verbose = false

var (
	debugLogger = log.New(os.Stderr, "", log.Lmsgprefix)
	infoLogger  = log.New(os.Stderr, "", log.Lmsgprefix)
	warnLogger  = log.New(os.Stderr, "WARNING: ", log.Lmsgprefix)
)

func SetVerbose(b bool) {
	verbose = b
}

func Debug(s string, args ...any) {
	if verbose {
		debugLogger.Printf(s, args...)
	}
}

func Printf(format string, args ...any) {
	infoLogger.Printf(format, args...)
}

func Fatal(format string, args ...any) {
	infoLogger.Printf(format, args...)
	os.Exit(1)
}
