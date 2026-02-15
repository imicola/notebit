package database

import (
	"time"

	"gorm.io/gorm"
)

// File represents a markdown file's metadata in the database
type File struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// User-facing fields
	Path         string `gorm:"uniqueIndex;not null" json:"path"`  // Relative path from basePath
	Title        string `gorm:"index" json:"title"`                // Extracted from filename or first # heading
	ContentHash  string `gorm:"index;size:64" json:"content_hash"` // SHA-256 for change detection
	LastModified int64  `json:"last_modified"`                     // Unix timestamp
	FileSize     int64  `json:"file_size"`                         // Bytes

	// Relationships
	Chunks []Chunk `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"chunks,omitempty"`
	Tags   []Tag   `gorm:"many2many:file_tags;" json:"tags,omitempty"`
}

// TableName specifies the table name for File
func (File) TableName() string {
	return "files"
}

// Chunk represents a text segment from a file for vectorization
type Chunk struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Content fields
	FileID  uint   `gorm:"not null;index" json:"file_id"`
	File    *File  `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	Content string `gorm:"type:text" json:"content"` // Text content
	Heading string `json:"heading"`                  // Associated heading (if any)

	// Vector fields
	Embedding          []float32  `gorm:"type:json;serializer:json" json:"embedding"` // Legacy JSON storage (fallback)
	EmbeddingBlob      []byte     `gorm:"type:blob" json:"-"`                         // Binary storage for vec_chunks migration
	EmbeddingModel     string     `gorm:"size:64" json:"embedding_model"`             // Model name/version
	EmbeddingCreatedAt *time.Time `json:"embedding_created_at"`                       // NULL until embedded
	VecIndexed         bool       `gorm:"index;default:false" json:"vec_indexed"`     // Whether embedding is written to vec_chunks
	EmbeddingDim       int        `gorm:"-" json:"embedding_dim,omitempty"`           // Computed field for UI
}

// TableName specifies the table name for Chunk
func (Chunk) TableName() string {
	return "chunks"
}

// GetEmbedding returns the embedding vector from either blob or JSON field
// Returns nil if no valid embedding is found
func (c *Chunk) GetEmbedding() []float32 {
	// Prefer EmbeddingBlob (optimized binary format)
	if len(c.EmbeddingBlob) > 0 {
		vec := bytesToFloats(c.EmbeddingBlob)
		if len(vec) > 0 {
			return vec
		}
		// Blob is malformed or empty, fall through to JSON
	}
	// Fallback to legacy JSON field
	if len(c.Embedding) > 0 {
		return c.Embedding
	}
	return nil
}

// Tag represents a tag that can be associated with files
type Tag struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name  string `gorm:"uniqueIndex;not null;size:100" json:"name"`
	Color string `gorm:"size:20" json:"color"` // Hex color for UI
}

// TableName specifies the table name for Tag
func (Tag) TableName() string {
	return "tags"
}

// FileTag represents the many-to-many relationship between Files and Tags
type FileTag struct {
	FileID uint `gorm:"primaryKey"`
	TagID  uint `gorm:"primaryKey"`
	File   File `gorm:"constraint:OnDelete:CASCADE"`
	Tag    Tag  `gorm:"constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for FileTag
func (FileTag) TableName() string {
	return "file_tags"
}
