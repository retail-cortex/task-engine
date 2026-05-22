package model

import (
	"time"
)

// SOP represents a Standard Operating Procedure document record.
type SOP struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title        string    `gorm:"type:varchar(255);not null"`
	CanonicalURL string    `gorm:"column:canonical_url;type:varchar(1024)"`
	Metadata     JSONB     `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
}

// SOPProcess holds records of standard RAG indexing pipeline executions.
type SOPProcess struct {
	ID               string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SOPID            string    `gorm:"column:sop_id;type:uuid;not null;index"`
	ChunkingStrategy string    `gorm:"type:varchar(100);not null"`
	EmbeddingModel   string    `gorm:"type:varchar(100);not null"`
	Status           string    `gorm:"type:varchar(50);not null;default:'PENDING'"`
	IsActive         bool      `gorm:"not null;default:false"`
	CreatedAt        time.Time `gorm:"not null;default:now()"`
}

// SOPChunk contains raw text chunks and associated high-dimensional embeddings for vector similarity search.
type SOPChunk struct {
	ID           string        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SOPID        string        `gorm:"column:sop_id;type:uuid;not null;index"`
	SOPProcessID string        `gorm:"column:sop_process_id;type:uuid;not null;uniqueIndex:idx_sop_chunk,priority:1"`
	ChunkIndex   int           `gorm:"type:int;not null;uniqueIndex:idx_sop_chunk,priority:2"`
	Content      string        `gorm:"type:text;not null"`
	Embedding    Float32Vector `gorm:"type:vector(768);not null"`
}
