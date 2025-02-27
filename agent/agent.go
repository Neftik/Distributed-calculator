package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Интерфейс агента для мокирования в тестах
type Agent interface {
	Start()
}

type DefaultAgent struct{}

func (a *DefaultAgent) Start() {
	StartAgentLogic()
}

// Глобальная переменная, позволяющая подменять агента в тестах
var ActiveAgent Agent = &DefaultAgent{}

// Task — структура задачи, получаемая от сервера
type Task struct {
	ID        string  `json:"id"`
	Arg1      float64 `json:"arg1"`
	Arg2      float64 `json:"arg2"`
	Operation string  `json:"operation"`
}

// Result — структура результата, отправляемая на сервер
type Result struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}

// URL оркестратора
var orchestratorURL = "http://localhost:8080/internal/task"

// Тайминги выполнения операций
var (
	timeAdditionMs      = getEnvInt("TIME_ADDITION_MS", 500)
	timeSubtractionMs   = getEnvInt("TIME_SUBTRACTION_MS", 500)
	timeMultiplicationMs = getEnvInt("TIME_MULTIPLICATIONS_MS", 700)
	timeDivisionMs      = getEnvInt("TIME_DIVISIONS_MS", 1000)
)

// **StartAgent** — точка входа, вызываемая из main
func StartAgent() {
	ActiveAgent.Start()
}

// **Основная логика агента**
func StartAgentLogic() {
	log.Println("Агент запущен и ожидает задачи...")

	power := getEnvInt("COMPUTING_POWER", 4)
	log.Printf("Используется COMPUTING_POWER = %d", power)

	taskQueue := make(chan Task, power)

	// Запускаем воркеры
	for i := 0; i < power; i++ {
		go worker(taskQueue)
	}

	// Цикл получения задач
	for {
		task, err := fetchTask()
		if err != nil {
			log.Printf("Ошибка получения задачи: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if task.ID == "" {
			log.Println("Ожидание задач от сервера...")
			time.Sleep(2 * time.Second)
			continue
		}

		// Отправляем задачу в очередь
		taskQueue <- task
	}
}

// **Получение задачи от оркестратора**
func fetchTask() (Task, error) {
	resp, err := http.Get(orchestratorURL)
	if err != nil {
		return Task{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Task{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return Task{}, fmt.Errorf("ошибка: %d", resp.StatusCode)
	}

	var response struct {
		Task Task `json:"task"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return Task{}, err
	}

	return response.Task, nil
}

// **Отправка результата обратно на сервер**
func sendResult(result Result) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	resp, err := http.Post(orchestratorURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ошибка: сервер вернул статус %d", resp.StatusCode)
	}

	log.Println("Результат успешно отправлен!")
	return nil
}

// **Воркеры-вычислители**
func worker(queue chan Task) {
	for task := range queue {
		log.Printf("Обработка: %f %s %f", task.Arg1, task.Operation, task.Arg2)

		// Добавляем задержку перед вычислением
		delay := getOperationDelay(task.Operation)
		log.Printf("Ожидание %d мс перед выполнением операции %s", delay, task.Operation)
		time.Sleep(time.Duration(delay) * time.Millisecond)

		value, err := compute(task.Arg1, task.Arg2, task.Operation)

		if err != nil {
			log.Printf("Ошибка вычисления: %v", err)
			continue
		}

		res := Result{ID: task.ID, Result: value}

		if err := sendResult(res); err != nil {
			log.Printf("Ошибка отправки результата: %v", err)
		} else {
			log.Printf("Результат отправлен: %f", value)
		}
	}
}

// **Функция вычисления**
func compute(arg1, arg2 float64, op string) (float64, error) {
	switch op {
	case "+":
		return arg1 + arg2, nil
	case "-":
		return arg1 - arg2, nil
	case "*":
		return arg1 * arg2, nil
	case "/":
		if arg2 == 0 {
			return 0, fmt.Errorf("деление на 0")
		}
		return arg1 / arg2, nil
	default:
		return 0, fmt.Errorf("неизвестная операция: %s", op)
	}
}

// **Функция получения задержки для операции**
func getOperationDelay(op string) int {
	switch op {
	case "+":
		return timeAdditionMs
	case "-":
		return timeSubtractionMs
	case "*":
		return timeMultiplicationMs
	case "/":
		return timeDivisionMs
	default:
		return 500 // Значение по умолчанию
	}
}

// **Функция чтения переменных окружения**
func getEnvInt(key string, defaultValue int) int {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
