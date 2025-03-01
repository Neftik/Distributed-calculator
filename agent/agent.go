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

	"github.com/joho/godotenv"
)

type Agent interface {
	Start()
}

type DefaultAgent struct{}

func (a *DefaultAgent) Start() {
	StartAgentLogic()
}

type Task struct {
	ID        string  `json:"id"`
	Arg1      float64 `json:"arg1"`
	Arg2      float64 `json:"arg2"`
	Operation string  `json:"operation"`
}

type Result struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}

var orchestratorURL = "http://localhost:8080/internal/task"

var (
	timeAdditionMs       int
	timeSubtractionMs    int
	timeMultiplicationMs int
	timeDivisionMs       int
)

func init() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Не удалось загрузить .env файл, использую стандартные значения")
	}

	timeAdditionMs = getEnvInt("TIME_ADDITION_MS", 500)
	timeSubtractionMs = getEnvInt("TIME_SUBTRACTION_MS", 500)
	timeMultiplicationMs = getEnvInt("TIME_MULTIPLICATIONS_MS", 700)
	timeDivisionMs = getEnvInt("TIME_DIVISIONS_MS", 1000)
}

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

func StartAgent() {
	ActiveAgent.Start()
}

var ActiveAgent Agent = &DefaultAgent{}

func StartAgentLogic() {
	log.Println("Агент запущен и ожидает задачи...")

	power := getEnvInt("COMPUTING_POWER", 4)
	log.Printf("Используется COMPUTING_POWER = %d", power)

	taskQueue := make(chan Task, power)

	for i := 0; i < power; i++ {
		go worker(taskQueue)
	}

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

		taskQueue <- task
	}
}

func fetchTask() (Task, error) {
	resp, err := http.Get(orchestratorURL)
	if err != nil {
		log.Printf("Ошибка получения задачи: %v", err)
		return Task{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.Println("Сервер ответил: задач нет (404)")
		return Task{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка: сервер вернул статус %d", resp.StatusCode)
		return Task{}, fmt.Errorf("ошибка: %d", resp.StatusCode)
	}

	var response Task
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Printf("Ошибка декодирования задачи: %v", err)
		return Task{}, err
	}

	log.Printf("Получена задача: %+v", response)
	return response, nil
}

func sendResult(result Result) error {
	log.Printf("Отправка результата: %+v", result)

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	resp, err := http.Post(orchestratorURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Ошибка отправки результата: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка: сервер вернул статус %d", resp.StatusCode)
		return fmt.Errorf("Ошибка: сервер вернул статус %d", resp.StatusCode)
	}

	log.Println("Результат успешно отправлен!")
	return nil
}

func worker(queue chan Task) {
	for task := range queue {
		log.Printf("Обработка задачи: %f %s %f", task.Arg1, task.Operation, task.Arg2)

		delay := getOperationDelay(task.Operation)
		log.Printf("Ожидание %d мс перед выполнением операции %s", delay, task.Operation)
		time.Sleep(time.Duration(delay) * time.Millisecond)

		value, err := compute(task.Arg1, task.Arg2, task.Operation)
		if err != nil {
			log.Printf("Ошибка вычисления: %v", err)
			continue
		}

		res := Result{ID: task.ID, Result: value}
		log.Printf("Результат вычисления: %f", value)

		if err := sendResult(res); err != nil {
			log.Printf("Ошибка отправки результата: %v", err)
		} else {
			log.Println("Результат успешно отправлен!")
		}
	}
}

func compute(arg1, arg2 float64, op string) (float64, error) {
	log.Printf("Вычисление: %f %s %f", arg1, op, arg2)

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
		return 500
	}
}
