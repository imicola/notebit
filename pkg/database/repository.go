package database

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"

	"notebit/pkg/logger"

	"gorm.io/gorm"
)

// Repository provides data access methods
type Repository struct {
	db           *gorm.DB
	vectorEngine VectorSearchEngine
	revision     atomic.Uint64
}

// NewRepository creates a new repository
func (m *Manager) Repository() *Repository {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.repo == nil {
		m.repo = &Repository{db: m.db}
		m.repo.vectorEngine = NewBruteForceVectorEngine()
	} else if m.repo.vectorEngine == nil {
		// Ensure vectorEngine is always initialized
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
	if result.Error == nil {
		r.revision.Add(1)
	}
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

func (r *Repository) ListFilesWithChunks() ([]File, error) {
	var files []File
	err := r.db.Preload("Chunks").Find(&files).Error
	return files, err
}

// DeleteFile removes a file from the index (cascade deletes chunks)
func (r *Repository) DeleteFile(path string) error {
	var chunkIDs []uint
	_ = r.db.Model(&Chunk{}).
		Joins("JOIN files ON files.id = chunks.file_id").
		Where("files.path = ?", path).
		Pluck("chunks.id", &chunkIDs).Error

	if len(chunkIDs) > 0 {
		_ = r.db.Exec("DELETE FROM vec_chunks WHERE chunk_id IN ?", chunkIDs).Error
	}

	err := r.db.Where("path = ?", path).Delete(&File{}).Error
	if err == nil {
		r.revision.Add(1)
	}
	return err
}

// RenameFile updates a file's path in the index
func (r *Repository) RenameFile(oldPath, newPath string) error {
	err := r.db.Model(&File{}).Where("path = ?", oldPath).Update("path", newPath).Error
	if err == nil {
		r.revision.Add(1)
	}
	return err
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
	var oldIDs []uint
	_ = r.db.Model(&Chunk{}).Where("file_id = ?", fileID).Pluck("id", &oldIDs).Error
	if len(oldIDs) > 0 {
		_ = r.db.Exec("DELETE FROM vec_chunks WHERE chunk_id IN ?", oldIDs).Error
	}

	err := r.db.Where("file_id = ?", fileID).Delete(&Chunk{}).Error
	if err == nil {
		r.revision.Add(1)
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

	var existingChunkIDs []uint
	if err := tx.Model(&Chunk{}).Where("file_id = ?", file.ID).Pluck("id", &existingChunkIDs).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(existingChunkIDs) > 0 {
		if err := tx.Exec("DELETE FROM vec_chunks WHERE chunk_id IN ?", existingChunkIDs).Error; err != nil {
			logger.Warn("failed to delete old vec_chunks rows: %v", err)
		}
	}

	if err := tx.Where("file_id = ?", file.ID).Delete(&Chunk{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	var vecTableExists bool
	_ = tx.Raw("SELECT COUNT(*) > 0 FROM sqlite_master WHERE type='table' AND name='vec_chunks'").Scan(&vecTableExists).Error

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
			chunk.VecIndexed = false
		}

		if err := tx.Create(&chunk).Error; err != nil {
			tx.Rollback()
			return err
		}

		if vecTableExists && len(chunkInput.Embedding) > 0 {
			if err := insertVecChunk(tx, chunk.ID, chunkInput.Embedding); err != nil {
				// Log with higher visibility - production systems should monitor this
				logger.Warn("[VECTOR_INDEX] Failed to insert vec chunk %d, vector search acceleration disabled for this chunk: %v", chunk.ID, err)
				// TODO: Add metric counter for vec_insert_failures
			} else if err := tx.Model(&Chunk{}).Where("id = ?", chunk.ID).Update("vec_indexed", true).Error; err != nil {
				logger.Warn("[VECTOR_INDEX] Failed to mark vec_indexed for chunk %d: %v", chunk.ID, err)
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}
	r.revision.Add(1)
	return nil
}

func (r *Repository) GetRevision() uint64 {
	return r.revision.Load()
}
