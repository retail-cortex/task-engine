package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// JSONB represents a raw JSON byte slice mapped to PostgreSQL JSONB.
type JSONB json.RawMessage

// Scan implements the sql.Scanner interface.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	result := make([]byte, len(bytes))
	copy(result, bytes)
	*j = result
	return nil
}

// Value implements the driver.Valuer interface.
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

// Float32Vector represents pgvector column values as a slice of float32 values.
type Float32Vector []float32

// Scan implements the sql.Scanner interface.
func (v *Float32Vector) Scan(value interface{}) error {
	if value == nil {
		*v = nil
		return nil
	}
	str, ok := value.(string)
	if !ok {
		bytes, ok := value.([]byte)
		if !ok {
			return errors.New("failed to scan vector: incompatible type")
		}
		str = string(bytes)
	}
	str = strings.Trim(str, "[]")
	if str == "" {
		*v = []float32{}
		return nil
	}
	parts := strings.Split(str, ",")
	res := make([]float32, len(parts))
	for i, p := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(p), 32)
		if err != nil {
			return fmt.Errorf("failed to parse vector element: %w", err)
		}
		res[i] = float32(val)
	}
	*v = res
	return nil
}

// Value implements the driver.Valuer interface.
func (v Float32Vector) Value() (driver.Value, error) {
	if len(v) == 0 {
		return nil, nil
	}
	parts := make([]string, len(v))
	for i, val := range v {
		parts[i] = strconv.FormatFloat(float64(val), 'f', -1, 32)
	}
	return "[" + strings.Join(parts, ",") + "]", nil
}
