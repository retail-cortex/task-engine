// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"time"
)

// SOPMetadata represents structured parameters stored inside the JSONB Metadata column of the SOP entity.
// Bridges standard ARTS and Schema.org metadata targets (caching dates, checksum finger prints, processing runs).
type SOPMetadata struct {
	FileType           string     `json:"file_type"`                     // e.g. "PDF", "HTML", "DOCX", "XLSX"
	SHA256Checksum     string     `json:"sha256_checksum,omitempty"`      // Hex checksum for binary files check
	CacheEffectiveDate *time.Time `json:"cache_effective_date,omitempty"` // Caching date context for HTML/web targets
	ETag               string     `json:"etag,omitempty"`                 // HTTP caching validator
	RawContentLength   int64      `json:"raw_content_length"`             // Byte size context
	ProcessedChunks    int        `json:"processed_chunks"`               // Chunks count mapping
	LastCheckedAt      time.Time  `json:"last_checked_at"`                // Expiry periodically checks sweep date
}

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
	Status           string    `gorm:"type:varchar(50);not null;default:'PENDING';index:idx_sop_processes_status_locked_at,priority:1"`
	IsActive         bool       `gorm:"not null;default:false"`
	LockedAt         *time.Time `gorm:"default:null;index:idx_sop_processes_status_locked_at,priority:2"`
	LockedBy         *string    `gorm:"type:varchar(255);default:null"`
	RetryCount       int        `gorm:"not null;default:0"`
	MaxRetries       int        `gorm:"not null;default:3"`
	LastError        *string    `gorm:"type:text;default:null"`
	CreatedAt        time.Time  `gorm:"not null;default:now()"`
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
