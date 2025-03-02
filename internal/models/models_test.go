package models

import (
	"encoding/json"
	"testing"
)

func TestExpressionJSON(t *testing.T) {
	// Тест сериализации Expression
	expr := Expression{
		ID:     "1",
		Input:  "2+2",
		Status: StatusCompleted,
	}

	result := 4.0
	expr.Result = &result

	data, err := json.Marshal(expr)
	if err != nil {
		t.Fatalf("Failed to marshal Expression: %v", err)
	}

	var unmarshalled Expression
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal Expression: %v", err)
	}

	if unmarshalled.ID != expr.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshalled.ID, expr.ID)
	}

	if unmarshalled.Status != expr.Status {
		t.Errorf("Status mismatch: got %s, want %s", unmarshalled.Status, expr.Status)
	}

	if unmarshalled.Result == nil || *unmarshalled.Result != *expr.Result {
		t.Errorf("Result mismatch: got %v, want %v", unmarshalled.Result, expr.Result)
	}
}

func TestTaskJSON(t *testing.T) {
	// Тест сериализации Task
	task := Task{
		ID:            "1",
		Arg1:          "2",
		Arg2:          "3",
		Operation:     "+",
		OperationTime: 1000,
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Failed to marshal Task: %v", err)
	}

	var unmarshalled Task
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal Task: %v", err)
	}

	if unmarshalled.ID != task.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshalled.ID, task.ID)
	}

	if unmarshalled.Arg1 != task.Arg1 {
		t.Errorf("Arg1 mismatch: got %s, want %s", unmarshalled.Arg1, task.Arg1)
	}

	if unmarshalled.Arg2 != task.Arg2 {
		t.Errorf("Arg2 mismatch: got %s, want %s", unmarshalled.Arg2, task.Arg2)
	}

	if unmarshalled.Operation != task.Operation {
		t.Errorf("Operation mismatch: got %s, want %s", unmarshalled.Operation, task.Operation)
	}

	if unmarshalled.OperationTime != task.OperationTime {
		t.Errorf("OperationTime mismatch: got %d, want %d", unmarshalled.OperationTime, task.OperationTime)
	}
}

func TestTaskResultJSON(t *testing.T) {
	result := TaskResult{
		ID:     "1",
		Result: 5.0,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal TaskResult: %v", err)
	}

	var unmarshalled TaskResult
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal TaskResult: %v", err)
	}

	if unmarshalled.ID != result.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshalled.ID, result.ID)
	}

	if unmarshalled.Result != result.Result {
		t.Errorf("Result mismatch: got %f, want %f", unmarshalled.Result, result.Result)
	}
}

func TestExpressionsResponseJSON(t *testing.T) {
	expr1 := Expression{
		ID:     "1",
		Input:  "2+2",
		Status: StatusCompleted,
	}
	result1 := 4.0
	expr1.Result = &result1

	expr2 := Expression{
		ID:     "2",
		Input:  "3*3",
		Status: StatusProcessing,
	}

	response := ExpressionsResponse{
		Expressions: []Expression{expr1, expr2},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal ExpressionsResponse: %v", err)
	}

	var unmarshalled ExpressionsResponse
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal ExpressionsResponse: %v", err)
	}

	if len(unmarshalled.Expressions) != len(response.Expressions) {
		t.Errorf("Expressions length mismatch: got %d, want %d", len(unmarshalled.Expressions), len(response.Expressions))
	}

	if unmarshalled.Expressions[0].ID != expr1.ID {
		t.Errorf("Expression ID mismatch: got %s, want %s", unmarshalled.Expressions[0].ID, expr1.ID)
	}

	if unmarshalled.Expressions[1].ID != expr2.ID {
		t.Errorf("Expression ID mismatch: got %s, want %s", unmarshalled.Expressions[1].ID, expr2.ID)
	}
}

func TestExpressionResponseJSON(t *testing.T) {
	expr := Expression{
		ID:     "1",
		Input:  "2+2",
		Status: StatusCompleted,
	}
	result := 4.0
	expr.Result = &result

	response := ExpressionResponse{
		Expression: expr,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal ExpressionResponse: %v", err)
	}

	var unmarshalled ExpressionResponse
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal ExpressionResponse: %v", err)
	}

	if unmarshalled.Expression.ID != expr.ID {
		t.Errorf("Expression ID mismatch: got %s, want %s", unmarshalled.Expression.ID, expr.ID)
	}

	if unmarshalled.Expression.Status != expr.Status {
		t.Errorf("Expression Status mismatch: got %s, want %s", unmarshalled.Expression.Status, expr.Status)
	}

	if unmarshalled.Expression.Result == nil || *unmarshalled.Expression.Result != *expr.Result {
		t.Errorf("Expression Result mismatch: got %v, want %v", unmarshalled.Expression.Result, expr.Result)
	}
}
