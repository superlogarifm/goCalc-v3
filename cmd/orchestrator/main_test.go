package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/superlogarifm/goCalc-v3/internal/models"
)

func TestHandleCalculate(t *testing.T) {
	o := NewOrchestrator()

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "валидное выражение",
			body:       `{"expression": "2+2"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "невалидное выражение",
			body:       `{"expression": "2+"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "пустое выражение",
			body:       `{"expression": ""}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "невалидный JSON",
			body:       `{"expression": 2+2}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(o.handleCalculate)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusCreated {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}

				if _, ok := response["id"]; !ok {
					t.Errorf("Response does not contain id field")
				}
			}
		})
	}
}

func TestHandleGetExpressions(t *testing.T) {
	o := NewOrchestrator()

	// Создаем выражение
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(`{"expression": "2+2"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	http.HandlerFunc(o.handleCalculate).ServeHTTP(rr, req)

	// Получаем список выражений
	req, _ = http.NewRequest("GET", "/api/v1/expressions", nil)
	rr = httptest.NewRecorder()
	http.HandlerFunc(o.handleGetExpressions).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response models.ExpressionsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if len(response.Expressions) != 1 {
		t.Errorf("Expected 1 expression, got %d", len(response.Expressions))
	}
}

func TestHandleGetExpression(t *testing.T) {
	o := NewOrchestrator()

	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(`{"expression": "2+2"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	http.HandlerFunc(o.handleCalculate).ServeHTTP(rr, req)

	var createResponse map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &createResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	id := createResponse["id"]
	req, _ = http.NewRequest("GET", "/api/v1/expressions/"+id, nil)
	rr = httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/api/v1/expressions/" + id
		o.handleGetExpression(w, r)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response models.ExpressionResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response.Expression.ID != id {
		t.Errorf("Expected expression ID %s, got %s", id, response.Expression.ID)
	}
}

func TestHandleTaskOperations(t *testing.T) {
	o := NewOrchestrator()
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(`{"expression": "2+2"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	http.HandlerFunc(o.handleCalculate).ServeHTTP(rr, req)

	req, _ = http.NewRequest("GET", "/internal/task", nil)
	rr = httptest.NewRecorder()
	http.HandlerFunc(o.handleGetTask).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var taskResponse models.TaskResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &taskResponse); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	taskResult := models.TaskResult{
		ID:     taskResponse.Task.ID,
		Result: 4,
	}

	taskResultJSON, _ := json.Marshal(taskResult)
	req, _ = http.NewRequest("POST", "/internal/task", bytes.NewBuffer(taskResultJSON))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	http.HandlerFunc(o.handleTaskResult).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что выражение завершено
	var createResponse map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &createResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	req, _ = http.NewRequest("GET", "/api/v1/expressions", nil)
	rr = httptest.NewRecorder()
	http.HandlerFunc(o.handleGetExpressions).ServeHTTP(rr, req)

	var expressionsResponse models.ExpressionsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &expressionsResponse); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	completed := false
	for _, expr := range expressionsResponse.Expressions {
		if expr.Status == models.StatusCompleted {
			completed = true
			break
		}
	}

	if !completed {
		t.Errorf("No completed expressions found")
	}
}
