package calculator

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var (
	ErrInvalidExpression = fmt.Errorf("invalid expression")
)

type TokenType int

const (
	Number TokenType = iota
	Operator
	LeftParen
	RightParen
)

type Token struct {
	Type  TokenType
	Value string
}

type Node struct {
	Left     *Node
	Right    *Node
	Token    Token
	TaskID   string
	Computed bool
	Result   *float64
}

func tokenize(expr string) ([]Token, error) {
	expr = strings.ReplaceAll(expr, " ", "")

	var tokens []Token
	var current strings.Builder

	for i := 0; i < len(expr); i++ {
		ch := rune(expr[i])

		if unicode.IsDigit(ch) || ch == '.' {
			current.WriteRune(ch)
			continue
		}

		if current.Len() > 0 {
			tokens = append(tokens, Token{Type: Number, Value: current.String()})
			current.Reset()
		}

		switch ch {
		case '+', '-', '*', '/':
			tokens = append(tokens, Token{Type: Operator, Value: string(ch)})
		case '(':
			tokens = append(tokens, Token{Type: LeftParen})
		case ')':
			tokens = append(tokens, Token{Type: RightParen})
		default:
			return nil, fmt.Errorf("unexpected character: %c", ch)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, Token{Type: Number, Value: current.String()})
	}

	if len(tokens) == 0 {
		return nil, ErrInvalidExpression
	}

	if tokens[0].Type == Operator || tokens[len(tokens)-1].Type == Operator {
		return nil, ErrInvalidExpression
	}

	return tokens, nil
}

func getPrecedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}

func buildAST(tokens []Token) (*Node, error) {
	var output []*Node
	var operators []*Node

	for _, token := range tokens {
		switch token.Type {
		case Number:
			val, err := strconv.ParseFloat(token.Value, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", token.Value)
			}
			output = append(output, &Node{Token: token, Result: &val})

		case Operator:
			node := &Node{Token: token}
			for len(operators) > 0 {
				top := operators[len(operators)-1]
				if top.Token.Type == LeftParen {
					break
				}
				if getPrecedence(top.Token.Value) >= getPrecedence(token.Value) {
					operators = operators[:len(operators)-1]
					right := output[len(output)-1]
					output = output[:len(output)-1]
					left := output[len(output)-1]
					output = output[:len(output)-1]
					top.Left = left
					top.Right = right
					output = append(output, top)
				} else {
					break
				}
			}
			operators = append(operators, node)

		case LeftParen:
			operators = append(operators, &Node{Token: token})

		case RightParen:
			for len(operators) > 0 {
				top := operators[len(operators)-1]
				operators = operators[:len(operators)-1]
				if top.Token.Type == LeftParen {
					break
				}
				right := output[len(output)-1]
				output = output[:len(output)-1]
				left := output[len(output)-1]
				output = output[:len(output)-1]
				top.Left = left
				top.Right = right
				output = append(output, top)
			}
		}
	}

	for len(operators) > 0 {
		top := operators[len(operators)-1]
		operators = operators[:len(operators)-1]
		right := output[len(output)-1]
		output = output[:len(output)-1]
		left := output[len(output)-1]
		output = output[:len(output)-1]
		top.Left = left
		top.Right = right
		output = append(output, top)
	}

	if len(output) != 1 {
		return nil, fmt.Errorf("invalid expression")
	}

	return output[0], nil
}

func ParseExpression(expr string) (*Node, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return nil, err
	}

	return buildAST(tokens)
}

func Calc(expression string) (float64, error) {
	node, err := ParseExpression(expression)
	if err != nil {
		return 0, err
	}

	if node.Token.Type == Number {
		val, err := strconv.ParseFloat(node.Token.Value, 64)
		if err != nil {
			return 0, err
		}
		return val, nil
	}

	if node.Token.Type == Operator {
		if node.Left == nil || node.Right == nil {
			return 0, ErrInvalidExpression
		}

		leftVal, err := Calc(node.Left.Token.Value)
		if err != nil {
			return 0, err
		}

		rightVal, err := Calc(node.Right.Token.Value)
		if err != nil {
			return 0, err
		}

		switch node.Token.Value {
		case "+":
			return leftVal + rightVal, nil
		case "-":
			return leftVal - rightVal, nil
		case "*":
			return leftVal * rightVal, nil
		case "/":
			if rightVal == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return leftVal / rightVal, nil
		default:
			return 0, fmt.Errorf("unknown operator: %s", node.Token.Value)
		}
	}

	return 0, ErrInvalidExpression
}
