package uuid

import (
	"database/sql/driver"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUUIDArray_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    UUIDArray
		wantErr bool
	}{
		{
			name:    "nil value",
			input:   nil,
			want:    UUIDArray{},
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    UUIDArray{},
			wantErr: false,
		},
		{
			name:    "valid UUID array",
			input:   "{550e8400-e29b-41d4-a716-446655440000,550e8400-e29b-41d4-a716-446655440001,71f5659f-b2dc-4219-8fd7-4a4f03e744c3}",
			want:    UUIDArray{uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), uuid.MustParse("71f5659f-b2dc-4219-8fd7-4a4f03e744c3")},
			wantErr: false,
		},
		{
			name:    "invalid UUID",
			input:   "{invalid-uuid}",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "unsupported type",
			input:   123,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ua UUIDArray
			err := ua.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("UUIDArray.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, ua)
		})
	}
}

func BenchmarkUUIDArray_Scan(b *testing.B) {
	var ua UUIDArray
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ua.Scan("{550e8400-e29b-41d4-a716-446655440000,550e8400-e29b-41d4-a716-446655440001,71f5659f-b2dc-4219-8fd7-4a4f03e744c3,550e8400-e29b-41d4-a716-446655440002,550e8400-e29b-41d4-a716-446655440003,550e8400-e29b-41d4-a716-446655440004,550e8400-e29b-41d4-a716-446655440005,550e8400-e29b-41d4-a716-446655440006,550e8400-e29b-41d4-a716-446655440007}")
	}
}

func TestUUIDArray_Value(t *testing.T) {
	tests := []struct {
		name    string
		input   UUIDArray
		want    driver.Value
		wantErr bool
	}{
		{
			name:    "nil array",
			input:   nil,
			want:    "{}",
			wantErr: false,
		},
		{
			name:    "empty array",
			input:   UUIDArray{},
			want:    "{}",
			wantErr: false,
		},
		{
			name:    "non-empty array",
			input:   UUIDArray{uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")},
			want:    `{"550e8400-e29b-41d4-a716-446655440000","550e8400-e29b-41d4-a716-446655440001"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("UUIDArray.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func BenchmarkUUIDArray_Value(b *testing.B) {
	// Create a larger dataset for benchmarking
	ua := UUIDArray{
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"),
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"),
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440006"),
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440007"),
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440008"),
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440009"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ua.Value()
	}
}

func TestUUIDArray_GormDataType(t *testing.T) {
	var ua UUIDArray
	assert.Equal(t, "uuid[]", ua.GormDataType())
}
