package files

import (
	"encoding/json"
	"time"
)

// FileNode represents a file or directory in the file tree
type FileNode struct {
	Name         string      `json:"name"`
	Path         string      `json:"path"`
	IsDir        bool        `json:"isDir"`
	ModifiedTime JSONTime    `json:"modifiedTime"`
	Size         int64       `json:"size"`
	Children     []*FileNode `json:"children,omitempty"`
}

// JSONTime wraps time.Time for custom JSON marshaling
type JSONTime struct {
	time.Time
}

// MarshalJSON implements json.Marshaler for JSONTime
func (t JSONTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time.Format(time.RFC3339))
}

// UnmarshalJSON implements json.Unmarshaler for JSONTime
func (t *JSONTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// NoteContent represents the content of a markdown file
type NoteContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// FileSystemError represents file system related errors
type FileSystemError struct {
	Op   string
	Path string
	Err  error
}

func (e *FileSystemError) Error() string {
	if e.Path != "" {
		return e.Op + " " + e.Path + ": " + e.Err.Error()
	}
	return e.Op + ": " + e.Err.Error()
}

func (e *FileSystemError) Unwrap() error {
	return e.Err
}
