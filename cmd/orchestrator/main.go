package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/superlogarifm/goCalc-v2/internal/calculator"
	"github.com/superlogarifm/goCalc-v2/internal/models"
)

type Orchestrator struct {
	taskManager *calculator.TaskManager
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		taskManager: calculator.NewTaskManager(),
	}
}

func (o *Orchestrator) handleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusUnprocessableEntity)
		return
	}

	id, err := o.taskManager.CreateExpression(req.Expression)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (o *Orchestrator) handleGetExpressions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	expressions := o.taskManager.GetAllExpressions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ExpressionsResponse{Expressions: expressions})
}

func (o *Orchestrator) handleGetExpression(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
	if expr, ok := o.taskManager.GetExpression(id); ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.ExpressionResponse{Expression: *expr})
		return
	}

	http.Error(w, "Expression not found", http.StatusNotFound)
}

func (o *Orchestrator) handleGetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if task, ok := o.taskManager.GetNextTask(); ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.TaskResponse{Task: *task})
		return
	}

	http.Error(w, "No tasks available", http.StatusNotFound)
}

func (o *Orchestrator) handleTaskResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var result models.TaskResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusUnprocessableEntity)
		return
	}

	if err := o.taskManager.UpdateTaskResult(result); err != nil {
		if err.Error() == "task not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	o := NewOrchestrator()
	mux := http.NewServeMux()

	// Публичные API endpoints
	mux.HandleFunc("/api/v1/calculate", o.handleCalculate)
	mux.HandleFunc("/api/v1/expressions", o.handleGetExpressions)
	mux.HandleFunc("/api/v1/expressions/", o.handleGetExpression)

	// Внутренние endpoints для агентов
	mux.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			o.handleGetTask(w, r)
		case http.MethodPost:
			o.handleTaskResult(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	handler := loggingMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	log.Printf("Starting orchestrator on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
