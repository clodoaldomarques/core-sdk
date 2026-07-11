package expression

import (
	"github.com/expr-lang/expr"
	"github.com/shopspring/decimal"
)

type Env struct {
	Amounts map[string]decimal.Decimal
	Fees    map[string]decimal.Decimal
	Sub     func(a, b interface{}) decimal.Decimal
	Add     func(a, b interface{}) decimal.Decimal
	Mul     func(a, b interface{}) decimal.Decimal
	Div     func(a, b interface{}) decimal.Decimal
	Sum     func(m map[string]decimal.Decimal) decimal.Decimal
}

// toDecimal converte interface{} para decimal.Decimal
func toDecimal(v interface{}) decimal.Decimal {
	switch val := v.(type) {
	case decimal.Decimal:
		return val
	case int:
		return decimal.NewFromInt(int64(val))
	case int64:
		return decimal.NewFromInt(val)
	case float64:
		return decimal.NewFromFloat(val)
	case float32:
		return decimal.NewFromFloat(float64(val))
	default:
		// Se for outro tipo, tenta converter via string (ex: string)
		return decimal.Zero
	}
}

func Calculate(expression string, amounts, fees map[string]decimal.Decimal) (decimal.Decimal, error) {
	options := []expr.Option{
		expr.Env(Env{}),
		expr.Operator("+", "Add"),
		expr.Operator("-", "Sub"),
		expr.Operator("*", "Mul"),
		expr.Operator("/", "Div"),
	}

	program, err := expr.Compile(expression, options...)
	if err != nil {
		return decimal.Decimal{}, ErrInvalidExpression{expression}
	}

	env := Env{
		Amounts: amounts,
		Fees:    fees,
		Add: func(a, b interface{}) decimal.Decimal {
			return toDecimal(a).Add(toDecimal(b))
		},
		Sub: func(a, b interface{}) decimal.Decimal {
			return toDecimal(a).Sub(toDecimal(b))
		},
		Mul: func(a, b interface{}) decimal.Decimal {
			return toDecimal(a).Mul(toDecimal(b))
		},
		Div: func(a, b interface{}) decimal.Decimal {
			return toDecimal(a).Div(toDecimal(b))
		},
		Sum: func(m map[string]decimal.Decimal) decimal.Decimal {
			total := decimal.Zero
			for _, v := range m {
				total = total.Add(v)
			}
			return total
		},
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return decimal.Decimal{}, ErrCalculateExpression{err.Error()}
	}

	return output.(decimal.Decimal), nil
}
