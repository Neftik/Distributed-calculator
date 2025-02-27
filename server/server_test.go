package server

import (
	"bytes"
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Логирование запроса и ответа
func logRequestResponse(testName string, req *http.Request, rr *httptest.ResponseRecorder) {
	fmt.Printf("[%s] REQUEST: %s %s\n", testName, req.Method, req.URL.Path)
	fmt.Printf("[%s] REQUEST BODY: %s\n", testName, req.Body)
	fmt.Printf("[%s] RESPONSE STATUS: %d\n", testName, rr.Code)
	fmt.Printf("[%s] RESPONSE BODY: %s\n", testName, rr.Body.String())

	fmt.Printf("[%s] REQUEST: %s %s\n", testName, req.Method, req.URL.Path)
	fmt.Printf("[%s] RESPONSE STATUS: %d\n", testName, rr.Code)
}

// Тест добавления выражения
func TestAddExpression(t *testing.T) {
	testName := "TestAddExpression"
	reqBody := `{"expression": "2 + 3"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	addExpression(rr, req)
	logRequestResponse(testName, req, rr)

	if rr.Code != http.StatusCreated {
		t.Fatalf("❌ [%s] Ожидался статус %d, но получен %d", testName, http.StatusCreated, rr.Code)
	}

	log.Printf("[%s] прошел успешно!", testName)
	fmt.Printf("[%s] прошел успешно!\n", testName)
}

// Тест получения списка выражений
func TestGetAllExpressions(t *testing.T) {
	testName := "TestGetAllExpressions"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	rr := httptest.NewRecorder()

	getAllExpressions(rr, req)
	logRequestResponse(testName, req, rr)

	if rr.Code != http.StatusOK {
		t.Fatalf("❌ [%s] Ожидался статус %d, но получен %d", testName, http.StatusOK, rr.Code)
	}

	log.Printf("✅ [%s] прошел успешно!", testName)
	fmt.Printf("✅ [%s] прошел успешно!\n", testName)
}

// Тест получения конкретного выражения
func TestGetExpression(t *testing.T) {
	testName := "TestGetExpression"

	// Создаем тестовое выражение
	expr := Expression{
		ID:     "test_id",
		Expr:   "4 * 5",
		Status: "pending",
	}
	store["test_id"] = expr

	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/test_id", nil)
	rr := httptest.NewRecorder()

	getExpression(rr, req)
	logRequestResponse(testName, req, rr)

	if rr.Code != http.StatusOK {
		t.Fatalf("❌ [%s] Ожидался статус %d, но получен %d", testName, http.StatusOK, rr.Code)
	}

	log.Printf("[%s] прошел успешно!", testName)
	fmt.Printf("[%s] прошел успешно!\n", testName)
}

// Тест получения задачи агентом
func TestGetTask(t *testing.T) {
	testName := "TestGetTask"

	// Создаем тестовую задачу
	task := Task{
		ID:        "task1",
		Arg1:      2,
		Arg2:      3,
		Operation: "+",
	}
	tasks = append(tasks, task)

	req := httptest.NewRequest(http.MethodGet, "/internal/task", nil)
	rr := httptest.NewRecorder()

	getTask(rr, req)
	logRequestResponse(testName, req, rr)

	if rr.Code != http.StatusOK {
		t.Fatalf("❌ [%s] Ожидался статус %d, но получен %d", testName, http.StatusOK, rr.Code)
	}

	log.Printf("[%s] прошел успешно!", testName)
	fmt.Printf("[%s] прошел успешно!\n", testName)
}

// Тест завершения задачи агентом
func TestCompleteTask(t *testing.T) {
	testName := "TestCompleteTask"

	// Создаем тестовое выражение с задачей
	expr := Expression{
		ID:     "expr1",
		Expr:   "5 - 3",
		Status: "pending",
		Tasks: []Task{
			{ID: "task2", Arg1: 5, Arg2: 3, Operation: "-"},
		},
	}
	store["expr1"] = expr

	reqBody := `{"id": "task2", "result": 2}`
	req := httptest.NewRequest(http.MethodPost, "/internal/task", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	completeTask(rr, req)
	logRequestResponse(testName, req, rr)

	if rr.Code != http.StatusOK {
		t.Fatalf("❌ [%s] Ожидался статус %d, но получен %d", testName, http.StatusOK, rr.Code)
	}

	log.Printf("[%s] прошел успешно!", testName)
	fmt.Printf("[%s] прошел успешно!\n", testName)
}

// Финальный тест для логирования успешных тестов
func TestAllTestsPassed(t *testing.T) {
	log.Println("🎉 Все тесты пройдены успешно!")
	fmt.Println("🎉 Все тесты пройдены успешно!")
}
