package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompute(t *testing.T) {
	tests := []struct {
		arg1, arg2 float64
		op         string
		expected   float64
		expectErr  bool
	}{
		{2, 3, "+", 5, false},
		{10, 4, "-", 6, false},
		{6, 7, "*", 42, false},
		{8, 2, "/", 4, false},
		{5, 0, "/", 0, true},
		{2, 3, "unknown", 0, true}, 
	}

	for _, tt := range tests {
		result, err := compute(tt.arg1, tt.arg2, tt.op)
		if (err != nil) != tt.expectErr {
			t.Errorf("compute(%f, %f, %s) ожидает ошибку: %v, получено: %v", tt.arg1, tt.arg2, tt.op, tt.expectErr, err)
		}
		if result != tt.expected {
			t.Errorf("compute(%f, %f, %s) = %f, ожидается %f", tt.arg1, tt.arg2, tt.op, result, tt.expected)
		}
	}
}

func TestFetchTask(t *testing.T) {
	task := Task{
		ID:        "123",
		Arg1:      10,
		Arg2:      2,
		Operation: "+",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]Task{"task": task})
	}))
	defer server.Close()

	orchestratorURL = server.URL

	receivedTask, err := fetchTask()
	if err != nil {
		t.Fatalf("fetchTask() вернул ошибку: %v", err)
	}
	if receivedTask.ID != task.ID || receivedTask.Arg1 != task.Arg1 || receivedTask.Arg2 != task.Arg2 || receivedTask.Operation != task.Operation {
		t.Errorf("fetchTask() получено %v, ожидается %v", receivedTask, task)
	}
}

func TestSendResult(t *testing.T) {
	result := Result{
		ID:     "123",
		Result: 42.0,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var received Result
		err := json.NewDecoder(r.Body).Decode(&received)
		if err != nil {
			t.Fatalf("Ошибка при декодировании JSON: %v", err)
		}

		if received.ID != result.ID || received.Result != result.Result {
			t.Errorf("sendResult() отправил %v, ожидается %v", received, result)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	orchestratorURL = server.URL

	err := sendResult(result)
	if err != nil {
		t.Fatalf("sendResult() вернул ошибку: %v", err)
	}
}

func TestWorker(t *testing.T) {
	task := Task{
		ID:        "task-1",
		Arg1:      6,
		Arg2:      7,
		Operation: "*",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var received Result
		err := json.NewDecoder(r.Body).Decode(&received)
		if err != nil {
			t.Fatalf("Ошибка при декодировании JSON: %v", err)
		}

		if received.Result != 42 {
			t.Errorf("worker() отправил %v, ожидается %v", received.Result, 42)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	orchestratorURL = server.URL

	taskQueue := make(chan Task, 1)
	go worker(taskQueue)

	taskQueue <- task
	close(taskQueue)
}
