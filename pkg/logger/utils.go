package logger

import (
	"bytes"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// getGID returns the current goroutine ID.
// Note: This is a hacky way to get the GID and may be slow.
// Use with caution in high-performance loops.
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

// getCallerInfo returns the file name, function name, and line number of the caller.
// skip is the number of stack frames to skip.
func getCallerInfo(skip int) (string, string, int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown", "unknown", 0
	}

	// Get function name
	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = fn.Name()
		// Strip package path to keep it cleaner, optional
		if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
			funcName = funcName[lastSlash+1:]
		}
	}

	// Get file name (basename)
	fileName := filepath.Base(file)

	return fileName, funcName, line
}
