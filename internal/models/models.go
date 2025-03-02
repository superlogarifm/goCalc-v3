package models

// статус вычисления выражения
type ExpressionStatus string

const (
	StatusPending    ExpressionStatus = "pending"
	StatusProcessing ExpressionStatus = "processing"
	StatusCompleted  ExpressionStatus = "completed"
	StatusError      ExpressionStatus = "error"
)

// арифметическое выражение
type Expression struct {
	ID       string           `json:"id"`
	Input    string           `json:"expression,omitempty"`
	Status   ExpressionStatus `json:"status"`
	Result   *float64         `json:"result,omitempty"`
	ErrorMsg string           `json:"error,omitempty"`
}

//запрос на вычисление
type CalculateRequest struct {
	Expression string `json:"expression" binding:"required"`
}

// вычислительная задача
type Task struct {
	ID            string   `json:"id"`
	Arg1          string   `json:"arg1"`
	Arg2          string   `json:"arg2"`
	Operation     string   `json:"operation"`
	OperationTime int64    `json:"operation_time"`
	Result        *float64 `json:"result,omitempty"`
}

// результат выполнения задачи
type TaskResult struct {
	ID     string  `json:"id" binding:"required"`
	Result float64 `json:"result" binding:"required"`
}

// ответ со списком выражений
type ExpressionsResponse struct {
	Expressions []Expression `json:"expressions"`
}

// ответ с одним выражением
type ExpressionResponse struct {
	Expression Expression `json:"expression"`
}

// ответ с задачей
type TaskResponse struct {
	Task Task `json:"task"`
}
