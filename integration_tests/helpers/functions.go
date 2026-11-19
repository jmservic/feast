package helpers

import (
	"fmt"
	"github.com/shopspring/decimal"
)

func CreateDecimal(value any) (decimal.Decimal, error) {
	switch v := value.(type) {
	case int:
		return decimal.NewFromInt(int64(v)), nil
	case int32:
		return decimal.NewFromInt32(v), nil
	case int64:
		return decimal.NewFromInt(v), nil
	case string:
		return decimal.NewFromString(v)
	case float32:
		return decimal.NewFromFloat32(v), nil
	case float64:
		return decimal.NewFromFloat(v), nil
	case decimal.Decimal:
		return v, nil
	default:
		return decimal.Zero, fmt.Errorf("unknown type %T", v)
	}
}
