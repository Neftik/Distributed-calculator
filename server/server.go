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

type Server interface {
	Start()
}

type DefaultServer struct{}

func (s *DefaultServer) Start() {
	StartServerLogic()
}
func StartServer() {
	ActiveServer.Start()
}

var ActiveServer Server = &DefaultServer{}

type Expression struct {
	ID     string   `json:"id"`
	Expr   string   `json:"expression"`
	Status string   `json:"status"`
	Result *float64 `json:"result,omitempty"`
	Tasks  []Task   `json:"tasks,omitempty"`
}

type Task struct {
	ID        string  `json:"id"`
	Arg1      float64 `json:"arg1"`
	Arg2      float64 `json:"arg2"`
	Operation string  `json:"operation"`
}

var (
	store = make(map[string]Expression)
	tasks = []Task{}
	mutex sync.Mutex
)

func generateID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func StartServerLogic() {
	http.HandleFunc("/api/v1/calculate", addExpression)
	http.HandleFunc("/api/v1/expressions", getAllExpressions)
	http.HandleFunc("/api/v1/expressions/", getExpression)
	http.HandleFunc("/internal/task", internalTaskHandler)
	log.Println("Сервер запущен на порту 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func parseExpressionIntoTasks(expressionID string, expression string) ([]Task, error) {
	tokens := strings.Fields(expression)
	if len(tokens) < 3 {
		return nil, fmt.Errorf("выражение должно содержать минимум 3 элемента")
	}

	var tasks []Task
	var values []float64
	var operators []string
	fmt.Println("Список задач:", tasks)

	for _, token := range tokens {
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			values = append(values, num)
		} else if token == "+" || token == "-" || token == "*" || token == "/" {
			operators = append(operators, token)
		} else {
			return nil, fmt.Errorf("неизвестный токен: %s", token)
		}
	}

	var newValues []float64
	var newOperators []string
	newValues = append(newValues, values[0])

	for i, op := range operators {
		if op == "*" || op == "/" {
			arg1 := newValues[len(newValues)-1]
			arg2 := values[i+1]
			var result float64
			if op == "*" {
				result = arg1 * arg2
			} else {
				if arg2 == 0 {
					return nil, fmt.Errorf("деление на ноль")
				}
				result = arg1 / arg2
			}
			newValues[len(newValues)-1] = result
		} else {
			newOperators = append(newOperators, op)
			newValues = append(newValues, values[i+1])
		}
	}

	finalValue := newValues[0]
	for i, op := range newOperators {
		if op == "+" {
			finalValue += newValues[i+1]
		} else {
			finalValue -= newValues[i+1]
		}
	}
	task := Task{
		ID:        expressionID,
		Arg1:      finalValue,
		Arg2:      0,
		Operation: "done",
	}

	return []Task{task}, nil
}

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
	tasksList, err := parseExpressionIntoTasks(id, req.Expression)
	if err != nil {
		http.Error(w, "Ошибка обработки выражения: "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Созданные задачи:", tasksList)

	expr := Expression{
		ID:     id,
		Expr:   req.Expression,
		Status: "pending",
		Tasks:  tasksList,
	}

	mutex.Lock()
	store[id] = expr
	tasks = append(tasks, tasksList...)
	fmt.Println("Общее количество задач в очереди после добавления:", len(tasks))
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func getAllExpressions(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	var exprs []Expression
	for _, expr := range store {
		exprs = append(exprs, expr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": exprs})
}

func getExpression(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
	mutex.Lock()
	expr, exists := store[id]
	mutex.Unlock()
	if !exists {
		http.Error(w, `{"error": "Expression not found"}`, http.StatusNotFound)
		return
	}

	response := struct {
		ID         string   `json:"id"`
		Expression string   `json:"expression"`
		Status     string   `json:"status"`
		Result     *float64 `json:"result,omitempty"`
	}{
		ID:         expr.ID,
		Expression: expr.Expr,
		Status:     expr.Status,
		Result:     expr.Result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"expression": response})
}

func getTask(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	fmt.Println("Запрос задачи. Количество в очереди:", len(tasks))

	if len(tasks) == 0 {
		fmt.Println("Очередь пуста!")
		http.Error(w, `{"error": "No tasks available"}`, http.StatusNotFound)
		return
	}

	task := tasks[0]
	tasks = tasks[1:]

	fmt.Println("Отправлена задача:", task)

	response := map[string]interface{}{
		"id":             task.ID,
		"arg1":           task.Arg1,
		"arg2":           task.Arg2,
		"operation":      task.Operation,
		"operation_time": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func completeTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusUnprocessableEntity)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	fmt.Printf("Получен результат задачи: ID=%s, Result=%f\n", req.ID, req.Result)

	for exprID, expr := range store {
		for i, task := range expr.Tasks {
			if task.ID == req.ID {
				expr.Tasks[i].Operation = "done"
				expr.Tasks[i].Arg1 = req.Result
				taskIDFloat, err := strconv.ParseFloat(task.ID, 64)
				if err != nil {
					fmt.Printf("Ошибка преобразования ID задачи %s в float64: %v\n", task.ID, err)
					continue
				}

				if i+1 < len(expr.Tasks) {
					if expr.Tasks[i+1].Arg1 == taskIDFloat {
						expr.Tasks[i+1].Arg1 = req.Result
					} else if expr.Tasks[i+1].Arg2 == taskIDFloat {
						expr.Tasks[i+1].Arg2 = req.Result
					}
				}
				break
			}
		}

		allDone := true
		for _, t := range expr.Tasks {
			if t.Operation != "done" {
				allDone = false
				break
			}
		}

		if allDone {
			fmt.Println("✅ Все задачи завершены, выполняем финальный расчёт...")

			var finalResult float64
			for _, t := range expr.Tasks {
				if t.Operation == "done" {
					finalResult = t.Arg1
				}
			}

			expr.Status = "done"
			expr.Result = &finalResult
			fmt.Printf("🎯 Итоговый результат выражения ID=%s: %f\n", exprID, finalResult)

			store[exprID] = expr
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "done"})
		return
	}

	fmt.Printf("⚠️ Ошибка: Задача с ID=%s не найдена\n", req.ID)
	http.Error(w, `{"error": "Task not found"}`, http.StatusNotFound)
}

func internalTaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTask(w, r)
	case http.MethodPost:
		completeTask(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
