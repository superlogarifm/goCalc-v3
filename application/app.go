package app

//
import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	calculate "github.com/superlogarifm/goCalc-v2/internal/calculator"
)

type Config struct {
	Host string
	Port string
}

func loadConfig() Config {
	os.Setenv("HOST", "127.0.0.1")
	os.Setenv("PORT", "8080")
	return Config{
		Host: os.Getenv("HOST"),
		Port: os.Getenv("PORT"),
	}
}

type App struct {
	config Config
}

func NewApp() *App {
	return &App{
		config: loadConfig(),
	}
}

type Request struct {
	Expression string `json:"expression"`
}

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	request := Request{}
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		fmt.Fprintf(w, `{"error": "method not allowed", "status": %d}`, http.StatusMethodNotAllowed)
		return
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	result, err := calculate.Calc(request.Expression)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, `{"error": "%s"}`, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"result": %f}`, result)
}

func (a App) StartServer() {
	http.HandleFunc("/api/v1/calculate", CalculateHandler)
	http.ListenAndServe(a.config.Host+":"+a.config.Port, nil)
}
