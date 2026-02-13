package database

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// Repository provides data access methods
type Repository struct {
	db                *gorm.DB
	vectorCache       map[uint][]float32
	vectorCacheLoaded bool
	vectorCacheMu     sync.RWMutex
	vectorEngine      VectorSearchEngine
}

// NewRepository creates a new repository
func (m *Manager) Repository() *Repository {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.repo == nil {
		m.repo = &Repository{db: m.db}
		m.repo.vectorEngine = NewBruteForceVectorEngine()
	}
	return m.repo
}

// ============ FILE OPERATIONS ============

// ChunkInput represents input data for creating a chunk
type ChunkInput struct {
	Content        string
	Heading        string
	Embedding      []float32
	EmbeddingModel string
}

// IndexFile indexes a file in the database
func (r *Repository) IndexFile(path, content string, lastModified int64, fileSize int64) error {
	// Calculate content hash
	hash := sha256.Sum256([]byte(content))
	contentHash := hex.EncodeToString(hash[:])

	// Extract title (first # heading or filename)
	title := extractTitle(path, content)

	file := File{
		Path:         path,
		Title:        title,
		ContentHash:  contentHash,
		LastModified: lastModified,
		FileSize:     fileSize,
	}

	// Use FirstOrCreate to handle updates
	result := r.db.Where("path = ?", path).Assign(file).FirstOrCreate(&file)
	return result.Error
}

// GetFileByPath retrieves a file by its path
func (r *Repository) GetFileByPath(path string) (*File, error) {
	var file File
	err := r.db.Where("path = ?", path).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// ListFiles retrieves all indexed files
func (r *Repository) ListFiles() ([]File, error) {
	var files []File
	err := r.db.Find(&files).Error
	return files, err
}

// DeleteFile removes a file from the index (cascade deletes chunks)
func (r *Repository) DeleteFile(path string) error {
	err := r.db.Where("path = ?", path).Delete(&File{}).Error
	if err == nil {
		r.invalidateVectorCache()
	}
	return err
}

// RenameFile updates a file's path in the index
func (r *Repository) RenameFile(oldPath, newPath string) error {
	return r.db.Model(&File{}).Where("path = ?", oldPath).Update("path", newPath).Error
}

// FileNeedsIndexing checks if a file needs to be re-indexed based on content hash
func (r *Repository) FileNeedsIndexing(path string, content string) (bool, error) {
	hash := sha256.Sum256([]byte(content))
	contentHash := hex.EncodeToString(hash[:])

	var existingFile File
	err := r.db.Where("path = ?", path).First(&existingFile).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return true, nil // File not indexed yet
		}
		return false, err
	}

	if existingFile.ContentHash != contentHash {
		return true, nil
	}

	// Content unchanged: still re-index when embeddings are missing or incomplete.
	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" {
		return false, nil
	}

	var totalChunks int64
	if err := r.db.Model(&Chunk{}).Where("file_id = ?", existingFile.ID).Count(&totalChunks).Error; err != nil {
		return false, err
	}
	if totalChunks == 0 {
		return true, nil
	}

	var embeddedChunks int64
	if err := r.db.Model(&Chunk{}).
		Where("file_id = ?", existingFile.ID).
		Where("embedding_blob IS NOT NULL").
		Where("length(embedding_blob) > 0").
		Count(&embeddedChunks).Error; err != nil {
		return false, err
	}

	return embeddedChunks < totalChunks, nil
}

// ============ CHUNK OPERATIONS ============

// CreateChunks creates chunks for a file (for future vectorization)
func (r *Repository) CreateChunks(fileID uint, chunks []Chunk) error {
	// Delete existing chunks for this file
	if err := r.db.Where("file_id = ?", fileID).Delete(&Chunk{}).Error; err != nil {
		return err
	}

	// Create new chunks
	for i := range chunks {
		chunks[i].FileID = fileID
		// chunks[i].ChunkIndex = i
	}

	if len(chunks) > 0 {
		if err := r.db.Create(&chunks).Error; err != nil {
			return err
		}
	}
	r.invalidateVectorCache()
	return nil
}

// GetChunksByFileID retrieves all chunks for a file
func (r *Repository) GetChunksByFileID(fileID uint) ([]Chunk, error) {
	var chunks []Chunk
	err := r.db.Where("file_id = ?", fileID).Order("id ASC").Find(&chunks).Error
	return chunks, err
}

// GetChunkByID retrieves a single chunk by ID
func (r *Repository) GetChunkByID(chunkID uint) (*Chunk, error) {
	var chunk Chunk
	err := r.db.First(&chunk, chunkID).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

// DeleteChunksForFile removes all chunks associated with a file
func (r *Repository) DeleteChunksForFile(fileID uint) error {
	err := r.db.Where("file_id = ?", fileID).Delete(&Chunk{}).Error
	if err == nil {
		r.invalidateVectorCache()
	}
	return err
}

// ============ TAG OPERATIONS ============

// GetOrCreateTag retrieves a tag by name or creates it
func (r *Repository) GetOrCreateTag(name string) (*Tag, error) {
	var tag Tag
	err := r.db.Where("name = ?", name).FirstOrCreate(&tag, Tag{Name: name}).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// ListTags retrieves all tags
func (r *Repository) ListTags() ([]Tag, error) {
	var tags []Tag
	err := r.db.Find(&tags).Error
	return tags, err
}

// AddTagToFile associates a tag with a file
func (r *Repository) AddTagToFile(fileID, tagID uint) error {
	return r.db.Exec("INSERT OR IGNORE INTO file_tags (file_id, tag_id) VALUES (?, ?)", fileID, tagID).Error
}

// RemoveTagFromFile removes a tag association from a file
func (r *Repository) RemoveTagFromFile(fileID, tagID uint) error {
	return r.db.Exec("DELETE FROM file_tags WHERE file_id = ? AND tag_id = ?", fileID, tagID).Error
}

// ============ UTILITY FUNCTIONS ============

// extractTitle extracts the title from content (first # heading) or filename
func extractTitle(path, content string) string {
	// Normalize line endings for cross-platform compatibility
	content = strings.ReplaceAll(content, "\r\n", "\n")

	// Try to find first heading (markdown # heading)
	re := regexp.MustCompile(`^#\s+(.+)$`)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				return strings.TrimSpace(matches[1])
			}
		}
	}

	// Fallback to filename without extension
	filename := filepath.Base(path)
	if idx := strings.LastIndex(filename, "."); idx > 0 {
		filename = filename[:idx]
	}
	return filename
}

// GetFilesByTag retrieves all files associated with a tag
func (r *Repository) GetFilesByTag(tagID uint) ([]File, error) {
	var files []File
	err := r.db.Joins("JOIN file_tags ON files.id = file_tags.file_id").
		Where("file_tags.tag_id = ?", tagID).
		Find(&files).Error
	return files, err
}

// GetTagsForFile retrieves all tags associated with a file
func (r *Repository) GetTagsForFile(fileID uint) ([]Tag, error) {
	var tags []Tag
	err := r.db.Joins("JOIN file_tags ON tags.id = file_tags.tag_id").
		Where("file_tags.file_id = ?", fileID).
		Find(&tags).Error
	return tags, err
}

// SearchFilesByTitle searches files by title (partial match)
func (r *Repository) SearchFilesByTitle(query string) ([]File, error) {
	var files []File
	err := r.db.Where("title LIKE ?", "%"+query+"%").Find(&files).Error
	return files, err
}

// GetStats returns database statistics
func (r *Repository) GetStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	var fileCount, chunkCount, tagCount int64

	if err := r.db.Model(&File{}).Count(&fileCount).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&Chunk{}).Count(&chunkCount).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&Tag{}).Count(&tagCount).Error; err != nil {
		return nil, err
	}

	stats["files"] = fileCount
	stats["chunks"] = chunkCount
	stats["tags"] = tagCount

	return stats, nil
}

// IndexFileWithChunks indexes a file with its chunks including embeddings
func (r *Repository) IndexFileWithChunks(path, content string, lastModified int64, fileSize int64, chunks []ChunkInput) error {
	// Calculate content hash
	hash := sha256.Sum256([]byte(content))
	contentHash := hex.EncodeToString(hash[:])

	// Extract title (first # heading or filename)
	title := extractTitle(path, content)

	// Start transaction
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Ensure transaction is rolled back on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create or update file
	file := File{
		Path:         path,
		Title:        title,
		ContentHash:  contentHash,
		LastModified: lastModified,
		FileSize:     fileSize,
	}

	// FirstOrCreate to handle updates
	if err := tx.Where("path = ?", path).Assign(file).FirstOrCreate(&file).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete existing chunks for this file
	if err := tx.Where("file_id = ?", file.ID).Delete(&Chunk{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create new chunks with embeddings
	now := r.db.NowFunc()
	for _, chunkInput := range chunks {
		chunk := Chunk{
			FileID:         file.ID,
			Content:        chunkInput.Content,
			Heading:        chunkInput.Heading,
			Embedding:      chunkInput.Embedding,
			EmbeddingModel: chunkInput.EmbeddingModel,
		}

		// Only set embedding timestamp if embedding is provided
		if len(chunkInput.Embedding) > 0 {
			chunk.EmbeddingCreatedAt = &now
			chunk.EmbeddingBlob = floatsToBytes(chunkInput.Embedding)
		}

		if err := tx.Create(&chunk).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}
	r.invalidateVectorCache()
	return nil
}
