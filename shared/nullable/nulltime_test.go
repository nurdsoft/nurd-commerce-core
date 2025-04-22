package nullable

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNullTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    NullTime
		wantErr bool
	}{
		{
			name:    "valid time",
			data:    `"2023-10-10T10:10:10Z"`,
			want:    NewNullTime(time.Date(2023, 10, 10, 10, 10, 10, 0, time.UTC)),
			wantErr: false,
		},
		{
			name:    "null time",
			data:    `null`,
			want:    NullTime{},
			wantErr: false,
		},
		{
			name:    "empty string",
			data:    `""`,
			want:    NullTime{},
			wantErr: false,
		},
		{
			name:    "invalid format",
			data:    `"invalid-time"`,
			want:    NullTime{},
			wantErr: true,
		},
		{
			name:    "zero time",
			data:    `"0001-01-01T00:00:00Z"`,
			want:    NullTime{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nt NullTime
			err := json.Unmarshal([]byte(tt.data), &nt)
			if (err != nil) != tt.wantErr {
				t.Errorf("NullTime.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, nt)
		})
	}
}
