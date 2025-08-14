package currency

import (
	"fmt"
	"github.com/shopspring/decimal"
)

// Currency represents a monetary value with proper decimal precision
type Currency struct {
	value decimal.Decimal
}

// NewFromFloat creates a Currency from a float64 (use carefully!)
// This should only be used when reading from legacy DBF files
func NewFromFloat(f float64) Currency {
	// Round to 2 decimal places for currency
	d := decimal.NewFromFloat(f).Round(2)
	return Currency{value: d}
}

// NewFromString creates a Currency from a string representation
// This is the preferred way to create Currency values
func NewFromString(s string) (Currency, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Currency{}, err
	}
	// Round to 2 decimal places for currency
	return Currency{value: d.Round(2)}, nil
}

// NewFromCents creates a Currency from an integer number of cents
func NewFromCents(cents int64) Currency {
	d := decimal.NewFromInt(cents).Div(decimal.NewFromInt(100))
	return Currency{value: d}
}

// Zero returns a zero Currency value
func Zero() Currency {
	return Currency{value: decimal.Zero}
}

// Add adds two Currency values
func (c Currency) Add(other Currency) Currency {
	return Currency{value: c.value.Add(other.value)}
}

// Sub subtracts a Currency value from another
func (c Currency) Sub(other Currency) Currency {
	return Currency{value: c.value.Sub(other.value)}
}

// Mul multiplies a Currency by a number
func (c Currency) Mul(multiplier decimal.Decimal) Currency {
	return Currency{value: c.value.Mul(multiplier).Round(2)}
}

// Div divides a Currency by a number
func (c Currency) Div(divisor decimal.Decimal) Currency {
	return Currency{value: c.value.Div(divisor).Round(2)}
}

// Neg returns the negative of a Currency value
func (c Currency) Neg() Currency {
	return Currency{value: c.value.Neg()}
}

// Abs returns the absolute value of a Currency
func (c Currency) Abs() Currency {
	return Currency{value: c.value.Abs()}
}

// IsPositive returns true if the Currency is positive
func (c Currency) IsPositive() bool {
	return c.value.IsPositive()
}

// IsNegative returns true if the Currency is negative
func (c Currency) IsNegative() bool {
	return c.value.IsNegative()
}

// IsZero returns true if the Currency is zero
func (c Currency) IsZero() bool {
	return c.value.IsZero()
}

// GreaterThan returns true if c > other
func (c Currency) GreaterThan(other Currency) bool {
	return c.value.GreaterThan(other.value)
}

// LessThan returns true if c < other
func (c Currency) LessThan(other Currency) bool {
	return c.value.LessThan(other.value)
}

// Equal returns true if c == other
func (c Currency) Equal(other Currency) bool {
	return c.value.Equal(other.value)
}

// ToCents returns the Currency value as integer cents
func (c Currency) ToCents() int64 {
	return c.value.Mul(decimal.NewFromInt(100)).IntPart()
}

// ToFloat64 returns the Currency value as a float64
// Use with caution - this can introduce precision errors
func (c Currency) ToFloat64() float64 {
	f, _ := c.value.Float64()
	return f
}

// ToString returns the Currency value as a string with 2 decimal places
func (c Currency) ToString() string {
	return c.value.StringFixed(2)
}

// ToJSON returns the Currency value suitable for JSON marshaling
// Returns a string to preserve precision
func (c Currency) ToJSON() string {
	return c.value.StringFixed(2)
}

// String implements the Stringer interface
func (c Currency) String() string {
	return fmt.Sprintf("$%s", c.value.StringFixed(2))
}

// ParseFromDBF parses a value from a DBF file into Currency
// Handles various types that DBF files might use for numeric values
func ParseFromDBF(value interface{}) Currency {
	if value == nil {
		return Zero()
	}
	
	switch v := value.(type) {
	case float64:
		return NewFromFloat(v)
	case float32:
		return NewFromFloat(float64(v))
	case int:
		return NewFromCents(int64(v * 100))
	case int64:
		return NewFromCents(v * 100)
	case string:
		// Try to parse as decimal
		if c, err := NewFromString(v); err == nil {
			return c
		}
		return Zero()
	default:
		return Zero()
	}
}

// SumCurrencies sums a slice of Currency values
func SumCurrencies(values []Currency) Currency {
	sum := Zero()
	for _, v := range values {
		sum = sum.Add(v)
	}
	return sum
}