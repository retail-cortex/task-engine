package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONB_ScanAndValue(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		var j JSONB
		err := j.Scan(nil)
		assert.NoError(t, err)
		assert.Nil(t, j)

		val, err := j.Value()
		assert.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("valid bytes", func(t *testing.T) {
		var j JSONB
		input := []byte(`{"key": "value"}`)
		err := j.Scan(input)
		assert.NoError(t, err)
		assert.Equal(t, JSONB(input), j)

		val, err := j.Value()
		assert.NoError(t, err)
		assert.Equal(t, input, val)
	})

	t.Run("invalid type", func(t *testing.T) {
		var j JSONB
		err := j.Scan(123)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type assertion to []byte failed")
	})
}

func TestFloat32Vector_ScanAndValue(t *testing.T) {
	tests := []struct {
		name          string
		scanInput     interface{}
		expectedScan  Float32Vector
		expectedValue interface{}
		scanError     bool
	}{
		{
			name:          "nil vector",
			scanInput:     nil,
			expectedScan:  nil,
			expectedValue: nil,
		},
		{
			name:          "empty vector string",
			scanInput:     "[]",
			expectedScan:  Float32Vector{},
			expectedValue: nil,
		},
		{
			name:          "valid vector string",
			scanInput:     "[0.1,0.2,0.3]",
			expectedScan:  Float32Vector{0.1, 0.2, 0.3},
			expectedValue: "[0.1,0.2,0.3]",
		},
		{
			name:          "valid vector bytes",
			scanInput:     []byte("[1.5,-2.5]"),
			expectedScan:  Float32Vector{1.5, -2.5},
			expectedValue: "[1.5,-2.5]",
		},
		{
			name:      "invalid element format",
			scanInput: "[0.1, abc]",
			scanError: true,
		},
		{
			name:      "invalid type scan",
			scanInput: 45.6,
			scanError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var v Float32Vector
			err := v.Scan(tc.scanInput)
			if tc.scanError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedScan, v)

				val, err := v.Value()
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedValue, val)
			}
		})
	}
}
