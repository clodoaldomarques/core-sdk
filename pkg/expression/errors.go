package expression

import "fmt"

type ErrInvalidExpression struct {
	expr string
}

func (e ErrInvalidExpression) Error() string {
	if e.expr != "" {
		return fmt.Sprintf("invalid expression: %s", e.expr)
	}
	return "invalid expression"
}

type ErrCalculateExpression struct {
	msg string
}

func (e ErrCalculateExpression) Error() string {
	if e.msg != "" {
		return fmt.Sprintf("error on calculate: %s", e.msg)
	}
	return "error on calculate"
}
