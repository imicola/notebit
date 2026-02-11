package main

import (
	"context"
	"notebit/pkg/files"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	fm      *files.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		fm: files.NewManager(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// OpenFolder opens a directory dialog and sets the base path
func (a *App) OpenFolder() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Notes Folder",
	})
	if err != nil {
		return "", err
	}

	if dir == "" {
		return "", nil
	}

	if err := a.fm.SetBasePath(dir); err != nil {
		return "", err
	}

	return dir, nil
}

// ListFiles returns the file tree structure
func (a *App) ListFiles() (*files.FileNode, error) {
	return a.fm.ListFiles()
}

// ReadFile reads the content of a markdown file
func (a *App) ReadFile(path string) (*files.NoteContent, error) {
	return a.fm.ReadFile(path)
}

// SaveFile saves content to a markdown file
func (a *App) SaveFile(path, content string) error {
	return a.fm.SaveFile(path, content)
}

// CreateFile creates a new markdown file
func (a *App) CreateFile(path, content string) error {
	return a.fm.CreateFile(path, content)
}

// DeleteFile deletes a markdown file or directory
func (a *App) DeleteFile(path string) error {
	return a.fm.DeleteFile(path)
}

// RenameFile renames a file or directory
func (a *App) RenameFile(oldPath, newPath string) error {
	return a.fm.RenameFile(oldPath, newPath)
}

// GetBasePath returns the current base path
func (a *App) GetBasePath() string {
	return a.fm.GetBasePath()
}
