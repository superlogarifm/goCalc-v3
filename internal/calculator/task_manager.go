package calculator

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/superlogarifm/goCalc-v3/internal/models"
)

type TaskManager struct {
	tasks          sync.Map
	expressions    sync.Map
	expressionASTs map[string]*Node
	taskQueue      chan models.Task
	nextID         int64
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		taskQueue:      make(chan models.Task, 100),
		nextID:         1,
		expressionASTs: make(map[string]*Node),
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

func (tm *TaskManager) CreateExpression(exprStr string) (string, error) {
	id := tm.generateID()

	ast, err := ParseExpression(exprStr)
	if err != nil {
		return "", err
	}
	tm.expressionASTs[id] = ast

	expression := models.Expression{
		ID:     id,
		Input:  exprStr,
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
			ExpressionID:  exprID,
		}

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

	if result.Error != nil {
		task.Error = result.Error
		task.Result = nil
		tm.tasks.Store(result.ID, task)

		if task.ExpressionID != "" {
			if exprVal, ok := tm.expressions.Load(task.ExpressionID); ok {
				exprToUpdate := exprVal.(models.Expression)
				if exprToUpdate.Status != models.StatusCompleted && exprToUpdate.Status != models.StatusError {
					exprToUpdate.Status = models.StatusError
					exprToUpdate.ErrorMsg = *result.Error
					tm.expressions.Store(task.ExpressionID, exprToUpdate)
					log.Printf("Expression %s failed due to task %s error: %s", task.ExpressionID, result.ID, *result.Error)
				}
			}
		}
		return nil
	}

	task.Result = &result.Result
	task.Error = nil
	tm.tasks.Store(result.ID, task)

	tm.checkAndUpdateExpressions()

	return nil
}

func (tm *TaskManager) checkAndUpdateExpressions() {
	var expressionsToComplete []string

	tm.expressions.Range(func(exprKey, exprValue interface{}) bool {
		exprID := exprKey.(string)
		expr := exprValue.(models.Expression)

		if expr.Status == models.StatusCompleted || expr.Status == models.StatusError {
			return true
		}

		astRootNode, astExists := tm.expressionASTs[exprID]
		if !astExists {
			log.Printf("AST not found for expression %s during checkAndUpdateExpressions, cannot check completion status.", exprID)
			return true
		}

		finalCalcResult, calcErr := tm.evaluateAST(astRootNode)

		if calcErr != nil {
			if calcErr.Error() == "task_not_ready" {
			} else {
				if expr.Status != models.StatusError {
					expr.Status = models.StatusError
					expr.ErrorMsg = calcErr.Error()
					tm.expressions.Store(exprID, expr)
					log.Printf("Expression %s marked as ERROR during check: %s", exprID, calcErr.Error())
				}
			}
		} else if finalCalcResult != nil {
			if expr.Status != models.StatusError {
				expr.Result = finalCalcResult
				expr.Status = models.StatusCompleted
				tm.expressions.Store(exprID, expr)
				expressionsToComplete = append(expressionsToComplete, exprID)
			}
		}
		return true
	})

	for _, exprID := range expressionsToComplete {
		if exprVal, ok := tm.expressions.Load(exprID); ok {
			expr := exprVal.(models.Expression)
			if expr.Status == models.StatusCompleted && expr.Result != nil {
				log.Printf("Expression %s COMPLETED with result: %f", exprID, *expr.Result)
			}
		}
	}
}

func (tm *TaskManager) evaluateAST(node *Node) (*float64, error) {
	if node.Token.Type == Number {
		val, err := strconv.ParseFloat(node.Token.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number in AST: %s", node.Token.Value)
		}
		return &val, nil
	}

	if node.Token.Type == Operator {
		taskInterface, taskExists := tm.tasks.Load(node.TaskID)
		if !taskExists {
			return nil, fmt.Errorf("task_not_ready")
		}
		task := taskInterface.(models.Task)

		if task.Error != nil {
			return nil, fmt.Errorf(*task.Error)
		}

		if task.Result == nil {
			return nil, fmt.Errorf("task_not_ready")
		}
		return task.Result, nil
	}
	return nil, fmt.Errorf("unknown node type in AST: %v", node.Token.Type)
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

func (tm *TaskManager) StartInternalWorker() {
	log.Println("Starting TaskManager internal worker...")
	go func() {
		for {
			taskFromQueue, ok := tm.GetNextTask()
			if !ok {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			task := *taskFromQueue

			log.Printf("Internal worker picked up task ID: %s, ExprID: %s (%s %s %s)", task.ID, task.ExpressionID, task.Arg1, task.Operation, task.Arg2)

			arg1, err1 := strconv.ParseFloat(task.Arg1, 64)
			arg2, err2 := strconv.ParseFloat(task.Arg2, 64)

			taskResult := models.TaskResult{ID: task.ID}

			if err1 != nil || err2 != nil {
				errMsg := fmt.Sprintf("Error parsing arguments for task %s (ExprID: %s): %v, %v.", task.ID, task.ExpressionID, err1, err2)
				log.Printf(errMsg)
				errorStr := errMsg
				taskResult.Error = &errorStr
			} else {
				var resultValue float64
				var calcError error

				switch task.Operation {
				case "+":
					resultValue = arg1 + arg2
				case "-":
					resultValue = arg1 - arg2
				case "*":
					resultValue = arg1 * arg2
				case "/":
					if arg2 == 0 {
						calcError = fmt.Errorf("division by zero")
					} else {
						resultValue = arg1 / arg2
					}
				default:
					calcError = fmt.Errorf("unknown operation: %s", task.Operation)
				}

				if calcError != nil {
					log.Printf("Error calculating task %s (ExprID: %s): %v", task.ID, task.ExpressionID, calcError)
					errorStr := calcError.Error()
					taskResult.Error = &errorStr
				} else {
					taskResult.Result = resultValue
				}
			}

			if err := tm.UpdateTaskResult(taskResult); err != nil {
				log.Printf("Error updating task result for task %s (ExprID: %s) in internal worker: %v", task.ID, task.ExpressionID, err)
			}
		}
	}()
	log.Println("TaskManager internal worker started.")
}
