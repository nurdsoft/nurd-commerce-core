package decimal

import (
	"testing"

	"github.com/shopspring/decimal"
)

func Test_MaxDecimal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values []decimal.Decimal
		want   decimal.Decimal
	}{
		{
			name:   "success",
			values: []decimal.Decimal{decimal.NewFromFloat(1.0), decimal.NewFromFloat(2.0), decimal.NewFromFloat(3.0)},
			want:   decimal.NewFromFloat(3.0),
		},
		{
			name:   "success with negative values",
			values: []decimal.Decimal{decimal.NewFromFloat(-1.0), decimal.NewFromFloat(-2.0), decimal.NewFromFloat(-3.0)},
			want:   decimal.NewFromFloat(-1.0),
		},
		{
			name:   "success with mixed values",
			values: []decimal.Decimal{decimal.NewFromFloat(-1.0), decimal.NewFromFloat(2.0), decimal.NewFromFloat(-3.0)},
			want:   decimal.NewFromFloat(2.0),
		},
		{
			name:   "success with zero values",
			values: []decimal.Decimal{decimal.NewFromFloat(0.0), decimal.NewFromFloat(0.0), decimal.NewFromFloat(0.0)},
			want:   decimal.NewFromFloat(0.0),
		},
		{
			name:   "success with single value",
			values: []decimal.Decimal{decimal.NewFromFloat(1.0)},
			want:   decimal.NewFromFloat(1.0),
		},
		{
			name:   "success with empty values",
			values: []decimal.Decimal{},
			want:   decimal.Decimal{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaxDecimal(tt.values...)

			if !got.Equal(tt.want) {
				t.Errorf("MaxDecimal() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_SumDecimals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values []decimal.Decimal
		want   decimal.Decimal
	}{
		{
			name:   "success",
			values: []decimal.Decimal{decimal.NewFromFloat(1.0), decimal.NewFromFloat(2.0), decimal.NewFromFloat(3.0)},
			want:   decimal.NewFromFloat(6.0),
		},
		{
			name:   "success with negative values",
			values: []decimal.Decimal{decimal.NewFromFloat(-1.0), decimal.NewFromFloat(-2.0), decimal.NewFromFloat(-3.0)},
			want:   decimal.NewFromFloat(-6.0),
		},
		{
			name:   "success with mixed values",
			values: []decimal.Decimal{decimal.NewFromFloat(-1.0), decimal.NewFromFloat(2.0), decimal.NewFromFloat(-3.0)},
			want:   decimal.NewFromFloat(-2.0),
		},
		{
			name:   "success with zero values",
			values: []decimal.Decimal{decimal.NewFromFloat(0.0), decimal.NewFromFloat(0.0), decimal.NewFromFloat(0.0)},
			want:   decimal.NewFromFloat(0.0),
		},
		{
			name:   "success with single value",
			values: []decimal.Decimal{decimal.NewFromFloat(1.0)},
			want:   decimal.NewFromFloat(1.0),
		},
		{
			name: "suceess with mixed positive values",
			values: []decimal.Decimal{
				decimal.NewFromFloat(1.0),
				decimal.NewFromFloat(2.3),
				decimal.NewFromFloat(3.0),
			},
			want: decimal.NewFromFloat(6.3),
		},
		{
			name:   "success with empty values",
			values: []decimal.Decimal{},
			want:   decimal.Decimal{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SumDecimals(tt.values...)

			if !got.Equal(tt.want) {
				t.Errorf("SumDecimals() got = %v, want %v", got, tt.want)
			}
		})
	}
}
