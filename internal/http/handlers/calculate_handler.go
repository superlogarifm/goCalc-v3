package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/superlogarifm/goCalc-v3/internal/calculator"
	"github.com/superlogarifm/goCalc-v3/internal/http/middleware"
	"github.com/superlogarifm/goCalc-v3/internal/models"
)

type CalculateRequest struct {
	Expression string `json:"expression"`
}

type CalculateResponse struct {
	ExpressionID string `json:"expression_id"`
}

type CalculateHandler struct {
	taskManager *calculator.TaskManager
}

func NewCalculateHandler(tm *calculator.TaskManager) *CalculateHandler {
	return &CalculateHandler{taskManager: tm}
}

func (h *CalculateHandler) HandleCalculate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		log.Println("Error: User ID not found in context after authentication middleware")
		http.Error(w, "Internal Server Error: User context missing", http.StatusInternalServerError)
		return
	}
	log.Printf("Received calculate request from UserID: %d\n", userID)

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	log.Printf("Raw request body: %s\n", string(bodyBytes))

	var req CalculateRequest
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding JSON request body: %v\n", err)
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Expression == "" {
		http.Error(w, `{"error": "Expression cannot be empty"}`, http.StatusBadRequest)
		return
	}

	expressionID, err := h.taskManager.CreateExpression(req.Expression)
	if err != nil {
		http.Error(w, `{"error": "invalid expression"}`, http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CalculateResponse{ExpressionID: expressionID})
}

func (h *CalculateHandler) HandleGetExpressions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		log.Println("Error: User ID not found in context for GetExpressions")
		http.Error(w, "Internal Server Error: User context missing", http.StatusInternalServerError)
		return
	}
	log.Printf("Received get all expressions request from UserID: %d\n", userID)

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	expressions := h.taskManager.GetAllExpressions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ExpressionsResponse{Expressions: expressions})
}

func (h *CalculateHandler) HandleGetExpressionByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		log.Println("Error: User ID not found in context for GetExpressionByID")
		http.Error(w, "Internal Server Error: User context missing", http.StatusInternalServerError)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
	if id == "" {
		http.Error(w, `{"error": "Expression ID is required in the path"}`, http.StatusBadRequest)
		return
	}
	log.Printf("Received get expression by ID request from UserID: %d for ExpressionID: %s\n", userID, id)

	expression, found := h.taskManager.GetExpression(id)
	if !found {
		http.Error(w, `{"error": "Expression not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ExpressionResponse{Expression: *expression})
}
