package watcher

import (
	"os"
)

// getFileStatRaw gets file info using os.Stat
func getFileStatRaw(path string) (os.FileInfo, error) {
	return os.Stat(path)
}
