package main

import (
	"log"
	"os"
	"sync"
	"testing"
	"time"
	"project2/server"
	"project2/agent"

	"github.com/joho/godotenv"
)

var serverStarted bool
var agentStarted bool
var mu sync.Mutex

type MockServer struct{}

func (m *MockServer) Start() {
	log.Println("[MockServer] Start() вызван")
	mu.Lock()
	serverStarted = true
	mu.Unlock()
	log.Println("[MockServer] serverStarted установлен в true")
}

type MockAgent struct{}

func (m *MockAgent) Start() {
	log.Println("[MockAgent] Start() вызван")
	mu.Lock()
	agentStarted = true
	mu.Unlock()
	log.Println("[MockAgent] agentStarted установлен в true")
}

// TestMainFunction тестирует запуск main() функции, которая должна вызвать Start() для сервера и агента.
func TestMainFunction(t *testing.T) {
	log.Println("===============================")
	log.Println("🚀 Запуск теста TestMainFunction")
	log.Println("===============================")

	// Загружаем переменные из .env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Не удалось загрузить .env файл, будут использоваться системные переменные окружения")
	}

	// Логируем значения из окружения
	log.Println("📌 Значения переменных из .env:")
	log.Printf("SERVER_PORT: %s", os.Getenv("SERVER_PORT"))
	log.Printf("AGENT_WORKERS: %s", os.Getenv("AGENT_WORKERS"))
	log.Printf("TIME_ADDITION_MS: %s", os.Getenv("TIME_ADDITION_MS"))
	log.Printf("TIME_SUBTRACTION_MS: %s", os.Getenv("TIME_SUBTRACTION_MS"))
	log.Printf("TIME_MULTIPLICATIONS_MS: %s", os.Getenv("TIME_MULTIPLICATIONS_MS"))
	log.Printf("TIME_DIVISIONS_MS: %s", os.Getenv("TIME_DIVISIONS_MS"))

	// Подменяем объекты
	log.Println("[Setup] Сохранение оригинальных объектов")
	originalServer := server.ActiveServer
	originalAgent := agent.ActiveAgent

	server.ActiveServer = &MockServer{}
	agent.ActiveAgent = &MockAgent{}
	log.Println("[Setup] Объекты подменены Mock-реализациями")

	// Восстановление оригинальных объектов после выполнения теста
	defer func() {
		log.Println("[Teardown] Восстановление оригинальных объектов")
		server.ActiveServer = originalServer
		agent.ActiveAgent = originalAgent
	}()

	// Канал для ожидания завершения выполнения main()
	done := make(chan bool)

	log.Println("[Execution] Запуск main() в отдельной горутине")
	go func() {
		main() // Запуск вашей функции main
		done <- true // Отправка сигнала о завершении
	}()

	// Ожидание выполнения 100ms
	log.Println("[Execution] Ожидание выполнения 100ms")
	time.Sleep(100 * time.Millisecond)

	// Проверка статусов serverStarted и agentStarted
	mu.Lock()
	log.Println("[Verification] Проверка статусов serverStarted и agentStarted")
	log.Printf("[Status] serverStarted=%v, agentStarted=%v", serverStarted, agentStarted)

	if !serverStarted {
		t.Error("❌ Ошибка: server.StartServer() не был вызван")
		log.Println("❌ Ошибка: server.StartServer() не был вызван")
	} else {
		log.Println("✅ Тест пройден: server.StartServer() был вызван")
	}

	if !agentStarted {
		t.Error("❌ Ошибка: agent.StartAgent() не был вызван")
		log.Println("❌ Ошибка: agent.StartAgent() не был вызван")
	} else {
		log.Println("✅ Тест пройден: agent.StartAgent() был вызван")
	}

	mu.Unlock()

	log.Println("===============================")
	log.Println("🎉 ✅ Все тесты успешно пройдены! 🎉")
	log.Println("===============================")
}
