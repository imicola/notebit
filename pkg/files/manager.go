package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Manager handles file system operations for notes
type Manager struct {
	basePath string
	mu       sync.RWMutex
}

// NewManager creates a new file system manager
func NewManager() *Manager {
	return &Manager{}
}

// SetBasePath sets the base directory for notes
func (m *Manager) SetBasePath(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if path exists and is a directory
	info, err := os.Stat(path)
	if err != nil {
		return &FileSystemError{Op: "stat", Path: path, Err: err}
	}

	if !info.IsDir() {
		return &FileSystemError{
			Op:   "validate",
			Path: path,
			Err:  fmt.Errorf("not a directory"),
		}
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return &FileSystemError{Op: "absolute", Path: path, Err: err}
	}

	m.basePath = absPath
	return nil
}

// GetBasePath returns the current base path
func (m *Manager) GetBasePath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.basePath
}

// ListFiles returns the file tree structure
func (m *Manager) ListFiles() (*FileNode, error) {
	m.mu.RLock()
	basePath := m.basePath
	m.mu.RUnlock()

	if basePath == "" {
		return nil, &FileSystemError{
			Op:  "list",
			Err: fmt.Errorf("no base path set"),
		}
	}

	return m.buildTree(basePath, "")
}

// buildTree recursively builds the file tree
func (m *Manager) buildTree(rootPath, relativePath string) (*FileNode, error) {
	fullPath := filepath.Join(rootPath, relativePath)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, &FileSystemError{Op: "stat", Path: fullPath, Err: err}
	}

	node := &FileNode{
		Name:         info.Name(),
		Path:         filepath.ToSlash(relativePath),
		IsDir:        info.IsDir(),
		ModifiedTime: JSONTime{info.ModTime()},
		Size:         info.Size(),
	}

	if !info.IsDir() {
		return node, nil
	}

	// Read directory contents
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return node, &FileSystemError{Op: "readdir", Path: fullPath, Err: err}
	}

	// Filter markdown files and directories
	var children []*FileNode
	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files/directories
		if strings.HasPrefix(name, ".") {
			continue
		}

		// Only include directories and markdown files
		if entry.IsDir() {
			childPath := filepath.Join(relativePath, name)
			child, err := m.buildTree(rootPath, childPath)
			if err != nil {
				continue // Skip problematic entries
			}
			children = append(children, child)
		} else if strings.HasSuffix(strings.ToLower(name), ".md") {
			childPath := filepath.Join(relativePath, name)
			child, err := m.buildTree(rootPath, childPath)
			if err != nil {
				continue
			}
			children = append(children, child)
		}
	}

	// Sort: directories first, then files, both alphabetically
	sort.Slice(children, func(i, j int) bool {
		if children[i].IsDir != children[j].IsDir {
			return children[i].IsDir
		}
		return children[i].Name < children[j].Name
	})

	node.Children = children
	return node, nil
}

// ReadFile reads the content of a markdown file
func (m *Manager) ReadFile(relativePath string) (*NoteContent, error) {
	m.mu.RLock()
	basePath := m.basePath
	m.mu.RUnlock()

	if basePath == "" {
		return nil, &FileSystemError{
			Op:  "read",
			Err: fmt.Errorf("no base path set"),
		}
	}

	fullPath := filepath.Join(basePath, relativePath)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, &FileSystemError{Op: "read", Path: fullPath, Err: err}
	}

	return &NoteContent{
		Path:    filepath.ToSlash(relativePath),
		Content: string(content),
	}, nil
}

// SaveFile saves content to a markdown file
func (m *Manager) SaveFile(relativePath, content string) error {
	m.mu.RLock()
	basePath := m.basePath
	m.mu.RUnlock()

	if basePath == "" {
		return &FileSystemError{
			Op:  "save",
			Err: fmt.Errorf("no base path set"),
		}
	}

	fullPath := filepath.Join(basePath, relativePath)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &FileSystemError{Op: "mkdir", Path: dir, Err: err}
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return &FileSystemError{Op: "write", Path: fullPath, Err: err}
	}

	return nil
}

// CreateFile creates a new markdown file
func (m *Manager) CreateFile(relativePath, content string) error {
	m.mu.RLock()
	basePath := m.basePath
	m.mu.RUnlock()

	if basePath == "" {
		return &FileSystemError{
			Op:  "create",
			Err: fmt.Errorf("no base path set"),
		}
	}

	fullPath := filepath.Join(basePath, relativePath)

	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		return &FileSystemError{
			Op:   "create",
			Path: fullPath,
			Err:  fmt.Errorf("file already exists"),
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &FileSystemError{Op: "mkdir", Path: dir, Err: err}
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return &FileSystemError{Op: "write", Path: fullPath, Err: err}
	}

	return nil
}

// DeleteFile deletes a markdown file or directory
func (m *Manager) DeleteFile(relativePath string) error {
	m.mu.RLock()
	basePath := m.basePath
	m.mu.RUnlock()

	if basePath == "" {
		return &FileSystemError{
			Op:  "delete",
			Err: fmt.Errorf("no base path set"),
		}
	}

	fullPath := filepath.Join(basePath, relativePath)

	if err := os.RemoveAll(fullPath); err != nil {
		return &FileSystemError{Op: "delete", Path: fullPath, Err: err}
	}

	return nil
}

// RenameFile renames a file or directory
func (m *Manager) RenameFile(oldPath, newPath string) error {
	m.mu.RLock()
	basePath := m.basePath
	m.mu.RUnlock()

	if basePath == "" {
		return &FileSystemError{
			Op:  "rename",
			Err: fmt.Errorf("no base path set"),
		}
	}

	oldFullPath := filepath.Join(basePath, oldPath)
	newFullPath := filepath.Join(basePath, newPath)

	// Ensure new directory exists
	newDir := filepath.Dir(newFullPath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return &FileSystemError{Op: "mkdir", Path: newDir, Err: err}
	}

	if err := os.Rename(oldFullPath, newFullPath); err != nil {
		return &FileSystemError{Op: "rename", Path: oldFullPath, Err: err}
	}

	return nil
}

// FileExists checks if a file exists
func (m *Manager) FileExists(relativePath string) bool {
	m.mu.RLock()
	basePath := m.basePath
	m.mu.RUnlock()

	if basePath == "" {
		return false
	}

	fullPath := filepath.Join(basePath, relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}
