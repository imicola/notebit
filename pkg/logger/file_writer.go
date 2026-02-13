package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type FileWriter struct {
	config      Config
	file        *os.File
	currentSize int64
	mu          sync.Mutex
	openDate    string // Date string of when the current file was opened
}

func NewFileWriter(config Config) (*FileWriter, error) {
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return nil, err
	}

	fw := &FileWriter{
		config: config,
	}

	if err := fw.openFile(); err != nil {
		return nil, err
	}

	return fw, nil
}

func (fw *FileWriter) openFile() error {
	filePath := filepath.Join(fw.config.LogDir, fw.config.FileName)

	// Check if file exists to get current size
	info, err := os.Stat(filePath)
	if err == nil {
		fw.currentSize = info.Size()
		// If existing file is too big, rotate immediately
		if fw.currentSize >= fw.config.MaxFileSize && fw.config.MaxFileSize > 0 {
			if err := fw.rotate(); err != nil {
				return err
			}
			// Rotate creates a new file, so we need to stat again or reset size
			fw.currentSize = 0
		}
	} else {
		fw.currentSize = 0
	}

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	fw.file = f
	fw.openDate = time.Now().Format("2006-01-02")
	return nil
}

func (fw *FileWriter) Write(data []byte) (int, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Check for rotation needs
	if fw.shouldRotate() {
		if err := fw.rotate(); err != nil {
			return 0, err // Or log to stderr
		}
	}

	n, err := fw.file.Write(data)
	if err == nil {
		fw.currentSize += int64(n)
	}
	return n, err
}

func (fw *FileWriter) shouldRotate() bool {
	// Check date
	currentDate := time.Now().Format("2006-01-02")
	if currentDate != fw.openDate {
		return true
	}

	// Check size
	if fw.config.MaxFileSize > 0 && fw.currentSize >= fw.config.MaxFileSize {
		return true
	}

	return false
}

func (fw *FileWriter) rotate() error {
	if fw.file != nil {
		fw.file.Close()
	}

	// Rename current file
	// Format: filename.YYYY-MM-DD-HH-MM-SS.bak
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	oldPath := filepath.Join(fw.config.LogDir, fw.config.FileName)
	newPath := filepath.Join(fw.config.LogDir, fmt.Sprintf("%s.%s.log", fw.config.FileName, timestamp))

	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	// Clean up old files
	go fw.cleanUp() // Async cleanup

	return fw.openFile()
}

func (fw *FileWriter) cleanUp() {
	if fw.config.MaxBackups <= 0 {
		return
	}

	files, err := os.ReadDir(fw.config.LogDir)
	if err != nil {
		return
	}

	var logFiles []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		// Filter files that look like rotated logs
		if strings.HasPrefix(f.Name(), fw.config.FileName+".") && strings.HasSuffix(f.Name(), ".log") {
			logFiles = append(logFiles, filepath.Join(fw.config.LogDir, f.Name()))
		}
	}

	// Sort by name (which includes timestamp, so effectively by date)
	// Or even better, sort by modification time if we wanted, but name is reliable with our format
	sort.Strings(logFiles)

	// Delete oldest if we have too many
	if len(logFiles) > fw.config.MaxBackups {
		filesToDelete := logFiles[:len(logFiles)-fw.config.MaxBackups]
		for _, f := range filesToDelete {
			os.Remove(f)
		}
	}
}

func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if fw.file != nil {
		return fw.file.Close()
	}
	return nil
}

var _ io.Writer = (*FileWriter)(nil)
