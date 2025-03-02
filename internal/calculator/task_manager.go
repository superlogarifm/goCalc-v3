package calculator

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/superlogarifm/goCalc-v2/internal/models"
)

type TaskManager struct {
	tasks       sync.Map
	expressions sync.Map
	taskQueue   chan models.Task
	nextID      int64 // Счетчик для генерации ID
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		taskQueue: make(chan models.Task, 100),
		nextID:    1,
	}
}

func (tm *TaskManager) generateID() string {
	return fmt.Sprintf("%d", atomic.AddInt64(&tm.nextID, 1))
}

func getOperationTime(operation string) int64 {
	var envVar string
	switch operation {
	case "+":
		envVar = "TIME_ADDITION_MS"
	case "-":
		envVar = "TIME_SUBTRACTION_MS"
	case "*":
		envVar = "TIME_MULTIPLICATIONS_MS"
	case "/":
		envVar = "TIME_DIVISIONS_MS"
	default:
		return 1000 // default 1 second
	}

	if val := os.Getenv(envVar); val != "" {
		if ms, err := strconv.ParseInt(val, 10, 64); err == nil {
			return ms
		}
	}
	return 1000
}

func (tm *TaskManager) CreateExpression(expr string) (string, error) {
	id := tm.generateID()

	ast, err := ParseExpression(expr)
	if err != nil {
		return "", err
	}

	expression := models.Expression{
		ID:     id,
		Input:  expr,
		Status: models.StatusProcessing,
	}
	tm.expressions.Store(id, expression)

	tm.createTasks(ast, id)

	return id, nil
}

func (tm *TaskManager) createTasks(node *Node, exprID string) {
	if node == nil {
		return
	}

	tm.createTasks(node.Left, exprID)
	tm.createTasks(node.Right, exprID)

	if node.Token.Type == Operator {
		taskID := tm.generateID()
		node.TaskID = taskID

		task := models.Task{
			ID:            taskID,
			Operation:     node.Token.Value,
			OperationTime: getOperationTime(node.Token.Value),
		}

		// Определяем аргументы задачи
		if node.Left.Token.Type == Number {
			task.Arg1 = node.Left.Token.Value
		} else {
			task.Arg1 = fmt.Sprintf("task:%s", node.Left.TaskID)
		}

		if node.Right.Token.Type == Number {
			task.Arg2 = node.Right.Token.Value
		} else {
			task.Arg2 = fmt.Sprintf("task:%s", node.Right.TaskID)
		}

		tm.tasks.Store(taskID, task)
		tm.taskQueue <- task
	}
}

func (tm *TaskManager) GetNextTask() (*models.Task, bool) {
	select {
	case task := <-tm.taskQueue:
		if strings.HasPrefix(task.Arg1, "task:") || strings.HasPrefix(task.Arg2, "task:") {
			tm.taskQueue <- task
			return nil, false
		}
		return &task, true
	default:
		return nil, false
	}
}

func (tm *TaskManager) UpdateTaskResult(result models.TaskResult) error {
	taskInterface, exists := tm.tasks.Load(result.ID)
	if !exists {
		return fmt.Errorf("task not found: %s", result.ID)
	}

	task := taskInterface.(models.Task)
	task.Result = &result.Result
	tm.tasks.Store(result.ID, task)

	var readyTasks []models.Task
	tm.tasks.Range(func(taskKey, taskValue interface{}) bool {
		t := taskValue.(models.Task)
		updated := false

		if t.Result != nil {
			return true
		}

		if strings.HasPrefix(t.Arg1, fmt.Sprintf("task:%s", result.ID)) {
			t.Arg1 = fmt.Sprintf("%f", result.Result)
			updated = true
		}
		if strings.HasPrefix(t.Arg2, fmt.Sprintf("task:%s", result.ID)) {
			t.Arg2 = fmt.Sprintf("%f", result.Result)
			updated = true
		}

		if updated {
			tm.tasks.Store(taskKey, t)
			if !strings.HasPrefix(t.Arg1, "task:") && !strings.HasPrefix(t.Arg2, "task:") {
				readyTasks = append(readyTasks, t)
			}
		}
		return true
	})

	for _, t := range readyTasks {
		select {
		case tm.taskQueue <- t:
		default:
			// Очередь полна
		}
	}

	var completedExpressions []string
	tm.expressions.Range(func(key, value interface{}) bool {
		exprID := key.(string)
		expr := value.(models.Expression)

		if expr.Status == models.StatusCompleted {
			return true
		}

		var rootTask *models.Task
		var allTasksCompleted = true

		tm.tasks.Range(func(taskKey, taskValue interface{}) bool {
			t := taskValue.(models.Task)

			if t.Result == nil {
				allTasksCompleted = false
				return false
			}
			isRoot := true
			tm.tasks.Range(func(_, otherTask interface{}) bool {
				other := otherTask.(models.Task)
				if strings.HasPrefix(other.Arg1, fmt.Sprintf("task:%s", taskKey)) ||
					strings.HasPrefix(other.Arg2, fmt.Sprintf("task:%s", taskKey)) {
					isRoot = false
					return false
				}
				return true
			})

			if isRoot {
				rootTask = &t
			}

			return true
		})

		if allTasksCompleted && rootTask != nil {
			expr.Status = models.StatusCompleted
			expr.Result = rootTask.Result
			tm.expressions.Store(exprID, expr)
			completedExpressions = append(completedExpressions, exprID)
		}

		return true
	})

	for _, exprID := range completedExpressions {
		if expr, ok := tm.expressions.Load(exprID); ok {
			e := expr.(models.Expression)
			log.Printf("Expression %s completed with result: %f", exprID, *e.Result)
		}
	}

	return nil
}

func (tm *TaskManager) GetExpression(id string) (*models.Expression, bool) {
	if expr, ok := tm.expressions.Load(id); ok {
		expression := expr.(models.Expression)
		return &expression, true
	}
	return nil, false
}

func (tm *TaskManager) GetAllExpressions() []models.Expression {
	var expressions []models.Expression
	tm.expressions.Range(func(key, value interface{}) bool {
		expressions = append(expressions, value.(models.Expression))
		return true
	})
	return expressions
}
