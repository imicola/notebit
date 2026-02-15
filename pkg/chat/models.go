package chat

import "time"

type Session struct {
	ID            string `gorm:"primaryKey;size:64" json:"id"`
	Title         string `gorm:"index;size:255" json:"title"`
	Category      string `gorm:"index;size:128" json:"category"`
	Archived      bool   `gorm:"index" json:"archived"`
	Favorite      bool   `gorm:"index" json:"favorite"`
	CreatedAtUnix int64  `gorm:"index" json:"created_at_unix"`
	UpdatedAtUnix int64  `gorm:"index" json:"updated_at_unix"`
	LastMessageAt int64  `gorm:"index" json:"last_message_at"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Session) TableName() string {
	return "chat_sessions"
}

type Message struct {
	ID               string `gorm:"primaryKey;size:64" json:"id"`
	SessionID        string `gorm:"index;size:64;not null" json:"session_id"`
	Role             string `gorm:"index;size:16" json:"role"`
	Content          string `gorm:"type:text" json:"content"`
	Encrypted        bool   `gorm:"index" json:"encrypted"`
	Sources          string `gorm:"type:text" json:"sources"`
	SourcesEncrypted bool   `gorm:"index" json:"sources_encrypted"`
	Status           string `gorm:"index;size:16" json:"status"`
	Timestamp        int64  `gorm:"index" json:"timestamp"`
	TokensUsed       *int   `json:"tokens_used,omitempty"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (Message) TableName() string {
	return "chat_messages"
}

type SessionTag struct {
	SessionID string `gorm:"primaryKey;size:64" json:"session_id"`
	Tag       string `gorm:"primaryKey;size:64;index" json:"tag"`
	CreatedAt time.Time
}

func (SessionTag) TableName() string {
	return "chat_session_tags"
}

type Setting struct {
	Scope     string `gorm:"primaryKey;size:64" json:"scope"`
	Key       string `gorm:"primaryKey;size:64" json:"key"`
	Value     string `gorm:"type:text" json:"value"`
	UpdatedAt time.Time
	CreatedAt time.Time
}

func (Setting) TableName() string {
	return "chat_settings"
}
