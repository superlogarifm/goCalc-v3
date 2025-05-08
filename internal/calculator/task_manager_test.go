package calculator

import (
	"testing"
	"time"

	"github.com/superlogarifm/goCalc-v3/internal/models"
)

func TestTaskManager_CreateExpression(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{
			name:    "простое выражение",
			expr:    "2+2",
			wantErr: false,
		},
		{
			name:    "сложное выражение",
			expr:    "2+2*2",
			wantErr: false,
		},
		{
			name:    "выражение со скобками",
			expr:    "(2+2)*2",
			wantErr: false,
		},
		{
			name:    "пустое выражение",
			expr:    "",
			wantErr: true,
		},
		{
			name:    "некорректное выражение",
			expr:    "2+",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTaskManager()
			id, err := tm.CreateExpression(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id == "" {
				t.Errorf("CreateExpression() id is empty")
			}
		})
	}
}

func TestTaskManager_GetNextTask(t *testing.T) {
	tm := NewTaskManager()
	_, err := tm.CreateExpression("2+2")
	if err != nil {
		t.Fatalf("Failed to create expression: %v", err)
	}

	task, ok := tm.GetNextTask()
	if !ok {
		t.Fatalf("GetNextTask() returned no task")
	}

	if task.Operation != "+" {
		t.Errorf("Task operation = %v, want +", task.Operation)
	}

	if task.Arg1 != "2" || task.Arg2 != "2" {
		t.Errorf("Task arguments = %v, %v, want 2, 2", task.Arg1, task.Arg2)
	}
}

func TestTaskManager_UpdateTaskResult(t *testing.T) {
	tm := NewTaskManager()

	id, err := tm.CreateExpression("2+2")
	if err != nil {
		t.Fatalf("Failed to create expression: %v", err)
	}

	task, ok := tm.GetNextTask()
	if !ok {
		t.Fatalf("GetNextTask() returned no task")
	}

	result := models.TaskResult{
		ID:     task.ID,
		Result: 4,
	}

	err = tm.UpdateTaskResult(result)
	if err != nil {
		t.Fatalf("UpdateTaskResult() error = %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	expr, ok := tm.GetExpression(id)
	if !ok {
		t.Fatalf("GetExpression() returned no expression")
	}

	if expr.Status != models.StatusCompleted {
		t.Errorf("Expression status = %v, want %v", expr.Status, models.StatusCompleted)
	}

	if expr.Result == nil || *expr.Result != 4 {
		t.Errorf("Expression result = %v, want 4", expr.Result)
	}
}

func TestTaskManager_ComplexExpression(t *testing.T) {
	tm := NewTaskManager()
	id, err := tm.CreateExpression("2+2*2")
	if err != nil {
		t.Fatalf("Failed to create expression: %v", err)
	}

	for i := 0; i < 2; i++ {
		task, ok := tm.GetNextTask()
		if !ok {
			// Возможно, задача еще не готова
			time.Sleep(100 * time.Millisecond)
			task, ok = tm.GetNextTask()
			if !ok {
				t.Fatalf("GetNextTask() returned no task on iteration %d", i)
			}
		}

		// Выполняем задачу
		var result float64
		switch task.Operation {
		case "*":
			result = 4 // 2*2
		case "+":
			result = 6 // 2+4
		}

		err = tm.UpdateTaskResult(models.TaskResult{
			ID:     task.ID,
			Result: result,
		})
		if err != nil {
			t.Fatalf("UpdateTaskResult() error = %v", err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	expr, ok := tm.GetExpression(id)
	if !ok {
		t.Fatalf("GetExpression() returned no expression")
	}

	if expr.Status != models.StatusCompleted {
		t.Errorf("Expression status = %v, want %v", expr.Status, models.StatusCompleted)
	}

	if expr.Result == nil || *expr.Result != 6 {
		t.Errorf("Expression result = %v, want 6", expr.Result)
	}
}
