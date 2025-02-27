package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Интерфейс сервера для удобства тестирования и замены в будущем
type Server interface {
	Start()
}

type DefaultServer struct{}

func (s *DefaultServer) Start() {
	StartServerLogic()
}

// Глобальная переменная для подмены сервера в тестах
var ActiveServer Server = &DefaultServer{}

// Expression — структура математического выражения
type Expression struct {
	ID      string   `json:"id"`
	Expr    string   `json:"expression"`
	Status  string   `json:"status"`
	Result  *float64 `json:"result,omitempty"`
	Tasks   []Task   `json:"tasks"`
}

// Task — структура задачи
type Task struct {
	ID        string  `json:"id"`
	Arg1      float64 `json:"arg1"`
	Arg2      float64 `json:"arg2"`
	Operation string  `json:"operation"`
}

var (
	store  = make(map[string]Expression) // Хранилище выражений
	tasks  = []Task{}                    // Очередь задач
	mutex  sync.Mutex                     // Мьютекс для синхронизации
)

// Генерация уникального ID
func generateID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

// **Запуск сервера**
func StartServer() {
	ActiveServer.Start()
}

// **Основная логика сервера**
func StartServerLogic() {
	http.HandleFunc("/api/v1/calculate", addExpression)  // POST /api/v1/calculate
	http.HandleFunc("/api/v1/expressions", getAllExpressions) // GET /api/v1/expressions
	http.HandleFunc("/api/v1/expressions/", getExpression) // GET /api/v1/expressions/:id
	http.HandleFunc("/internal/task", internalTaskHandler) // GET & POST /internal/task
	log.Println("Сервер запущен на порту 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// **POST /api/v1/calculate** – Добавление выражения
func addExpression(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	id := generateID()
	tokens := strings.Fields(req.Expression)

	if len(tokens) < 3 {
		http.Error(w, "Invalid expression format", http.StatusBadRequest)
		return
	}

	var expr Expression
	expr.ID = id
	expr.Expr = req.Expression
	expr.Status = "pending"

	// Разбиваем выражение на задачи
	for i := 0; i < len(tokens)-2; i += 2 {
		arg1, err1 := strconv.ParseFloat(tokens[i], 64)
		arg2, err2 := strconv.ParseFloat(tokens[i+2], 64)
		operation := tokens[i+1]

		if err1 != nil || err2 != nil {
			http.Error(w, "Invalid numbers in expression", http.StatusBadRequest)
			return
		}

		// Проверка допустимых операций
		if operation != "+" && operation != "-" && operation != "*" && operation != "/" {
			http.Error(w, fmt.Sprintf("Invalid operation: %s", operation), http.StatusBadRequest)
			return
		}

		task := Task{
			ID:        generateID(),
			Arg1:      arg1,
			Arg2:      arg2,
			Operation: operation,
		}

		mutex.Lock()
		tasks = append(tasks, task)
		expr.Tasks = append(expr.Tasks, task)
		store[id] = expr
		mutex.Unlock()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// **GET /api/v1/expressions** – Получение списка выражений
func getAllExpressions(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	expressions := []Expression{}
	for _, expr := range store {
		expressions = append(expressions, expr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": expressions})
}

// **GET /api/v1/expressions/:id** – Получение выражения по ID
func getExpression(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
	mutex.Lock()
	expr, exists := store[id]
	mutex.Unlock()

	if !exists {
		http.Error(w, `{"error": "Expression not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"expression": expr})
}

// **GET /internal/task** – Получение задачи агентом
func getTask(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	if len(tasks) == 0 {
		http.Error(w, "No tasks available", http.StatusNotFound)
		return
	}

	task := tasks[0]
	tasks = tasks[1:]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
}

// **POST /internal/task** – Отправка результата задачи агентом
func completeTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for id, expr := range store {
		for i, task := range expr.Tasks {
			if task.ID == req.ID {
				expr.Tasks[i].Operation = "done"
				expr.Tasks[i].Arg1 = req.Result

				// Проверяем, все ли задачи выполнены
				allDone := true
				var finalResult float64
				for _, t := range expr.Tasks {
					if t.Operation != "done" {
						allDone = false
						break
					}
					finalResult = t.Arg1
				}

				if allDone {
					expr.Status = "done"
					expr.Result = &finalResult
				}

				store[id] = expr
				json.NewEncoder(w).Encode(map[string]string{"status": "done"})
				return
			}
		}
	}

	http.Error(w, "Task not found", http.StatusNotFound)
}

// **Обработчик для внутреннего API**
func internalTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		getTask(w, r)
	} else if r.Method == http.MethodPost {
		completeTask(w, r)
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
