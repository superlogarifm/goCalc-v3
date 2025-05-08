package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/superlogarifm/goCalc-v3/internal/models"
)

type Agent struct {
	orchestratorURL string
	client          *http.Client
}

func NewAgent(orchestratorURL string) *Agent {
	return &Agent{
		orchestratorURL: orchestratorURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  true,
				DisableKeepAlives:   false,
				MaxConnsPerHost:     100,
				MaxIdleConnsPerHost: 100,
			},
		},
	}
}

func (a *Agent) getTask() (*models.Task, error) {
	resp, err := a.client.Get(a.orchestratorURL + "/internal/task")
	if err != nil {
		if os.IsTimeout(err) || isConnectionRefused(err) {
			log.Printf("Оркестратор недоступен, ожидание...")
			time.Sleep(5 * time.Second)
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var taskResp models.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return nil, err
	}

	return &taskResp.Task, nil
}

func isConnectionRefused(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "context deadline exceeded"))
}

func (a *Agent) submitResult(result models.TaskResult) error {
	body, err := json.Marshal(result)
	if err != nil {
		return err
	}

	resp, err := a.client.Post(
		a.orchestratorURL+"/internal/task",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		if os.IsTimeout(err) || isConnectionRefused(err) {
			log.Printf("Оркестратор недоступен при отправке результата, повторная попытка...")
			time.Sleep(5 * time.Second)
			return a.submitResult(result)
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (a *Agent) processTask(task models.Task) error {
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond) // Имитация длительного вычисления

	taskResult := models.TaskResult{ID: task.ID}

	if strings.HasPrefix(task.Arg1, "task:") || strings.HasPrefix(task.Arg2, "task:") {
		log.Printf("Task %s (ExprID: %s) has unresolved dependencies, returning to queue", task.ID, task.ExpressionID)
		// В этом случае агент не должен отправлять результат, т.к. задача не его.
		// Оркестратор должен сам перевыставить задачу, когда зависимости разрешатся.
		// Однако, для консистентности, можно было бы отправить ошибку "unresolved_dependencies".
		// Но текущая логика оркестратора в GetNextTask уже обрабатывает это, возвращая задачу в очередь.
		// Поэтому здесь просто выходим, не отправляя результат.
		return nil
	}

	arg1, err1 := strconv.ParseFloat(task.Arg1, 64)
	arg2, err2 := strconv.ParseFloat(task.Arg2, 64)

	if err1 != nil || err2 != nil {
		errMsg := fmt.Sprintf("invalid arguments for task %s (ExprID: %s): arg1='%s', arg2='%s', err1=%v, err2=%v", task.ID, task.ExpressionID, task.Arg1, task.Arg2, err1, err2)
		log.Printf("Error processing task: %s", errMsg)
		taskResult.Error = &errMsg
		return a.submitResult(taskResult)
	}

	var operationResult float64
	var calcError error

	switch task.Operation {
	case "+":
		operationResult = arg1 + arg2
	case "-":
		operationResult = arg1 - arg2
	case "*":
		operationResult = arg1 * arg2
	case "/":
		if arg2 == 0 {
			calcError = fmt.Errorf("division by zero")
		} else {
			operationResult = arg1 / arg2
		}
	default:
		calcError = fmt.Errorf("unknown operation: %s", task.Operation)
	}

	if calcError != nil {
		errMsg := fmt.Sprintf("error calculating task %s (ExprID: %s): %v", task.ID, task.ExpressionID, calcError)
		log.Printf("Error processing task: %s", errMsg)
		taskResult.Error = &errMsg
		return a.submitResult(taskResult)
	}

	taskResult.Result = operationResult
	log.Printf("Task %s (ExprID: %s) completed: %f %s %f = %f", task.ID, task.ExpressionID, arg1, task.Operation, arg2, operationResult)
	return a.submitResult(taskResult)
}

func (a *Agent) worker(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		task, err := a.getTask()
		if err != nil {
			log.Printf("Error getting task: %v", err)
			continue
		}

		if task == nil {
			time.Sleep(time.Second)
			continue
		}

		if err := a.processTask(*task); err != nil {
			log.Printf("Error submitting result for task %s (ExprID: %s): %v", task.ID, task.ExpressionID, err)
		}
	}
}

func main() {
	orchestratorURL := os.Getenv("ORCHESTRATOR_URL")
	if orchestratorURL == "" {
		orchestratorURL = "http://localhost:8080"
	}

	computingPower := 4
	if cp := os.Getenv("COMPUTING_POWER"); cp != "" {
		if val, err := strconv.Atoi(cp); err == nil {
			computingPower = val
		}
	}

	agent := NewAgent(orchestratorURL)
	var wg sync.WaitGroup

	log.Printf("Starting agent with %d workers, connecting to %s", computingPower, orchestratorURL)
	for i := 0; i < computingPower; i++ {
		wg.Add(1)
		go agent.worker(&wg)
	}

	wg.Wait()
}
