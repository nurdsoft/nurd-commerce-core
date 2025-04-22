package decimal

import (
	"github.com/shopspring/decimal"
)

// MaxDecimal returns the maximum value from the provided decimal.Decimal arguments.
func MaxDecimal(values ...decimal.Decimal) decimal.Decimal {
	if len(values) > 0 {
		max := values[0]
		for _, value := range values[1:] {
			if value.GreaterThan(max) {
				max = value
			}
		}
		return max
	}

	return decimal.Decimal{}
}

// SumDecimals returns the sum of the provided decimal.Decimal arguments.
func SumDecimals(values ...decimal.Decimal) decimal.Decimal {
	sum := decimal.Zero
	for _, value := range values {
		sum = sum.Add(value)
	}
	return sum
}
