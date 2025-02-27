
# Распределённый вычислитель арифметических выражений

Эта система позволяет пользователям отправлять арифметические выражения, которые затем парсятся, вычисляются, и результаты возвращаются после обработки. Система построена по архитектуре сервер-агент, где сервер управляет задачами и выражениями, а агенты выполняют вычисления асинхронно.

---

## Как запустить

### Запуск сервера

Сервер будет слушать на порту `http://localhost:8080`.

```bash
go run server.go
```

### Запуск агента(ов)

Агенты начнут свою работу и будут слушать задачи для вычислений.

```bash
go run agent.go
```

### Главная функция запускает и сервер, и агента параллельно.

```bash
go run main.go
```

---

## Использование API

### Добавить выражение (POST /api/v1/calculate)

Этот API принимает арифметическое выражение и разбивает его на задачи для вычислений агентами.

#### Запрос
- **Метод**: `POST`
- **Эндпоинт**: `/api/v1/calculate`
- **Тело**:
  ```json
  {
    "expression": "2 + 3 * 4"
  }
  ```

#### Ответ
- **Статус**: `201 Created`
- **Тело**:
  ```json
  {
    "id": "unique-expression-id"
  }
  ```

### Получить все выражения (GET /api/v1/expressions)

Этот API получает все выражения, которые хранятся на сервере.

#### Запрос
- **Метод**: `GET`
- **Эндпоинт**: `/api/v1/expressions`

#### Ответ
- **Статус**: `200 OK`
- **Тело**:
  ```json
  {
    "expressions": [
      {
        "id": "unique-expression-id",
        "expression": "2 + 3 * 4",
        "status": "done",
        "result": 14.0
      }
    ]
  }
  ```

### Получить выражение по ID (GET /api/v1/expressions/{id})

Получить конкретное выражение по его ID.

#### Запрос
- **Метод**: `GET`
- **Эндпоинт**: `/api/v1/expressions/{id}`

#### Ответ
- **Статус**: `200 OK`
- **Тело**:
  ```json
  {
    "expression": {
      "id": "unique-expression-id",
      "expression": "2 + 3 * 4",
      "status": "done",
      "result": 14.0
    }
  }
  ```

### Получить задачу (GET /internal/task)

Этот API позволяет агентам получать задачу для вычислений.

#### Запрос
- **Метод**: `GET`
- **Эндпоинт**: `/internal/task`

#### Ответ
- **Статус**: `200 OK`
- **Тело**:
  ```json
  {
    "task": {
      "id": "task-id",
      "arg1": 2,
      "arg2": 3,
      "operation": "+"
    }
  }
  ```

### Завершить задачу (POST /internal/task)

После завершения вычислений агент отправляет результат обратно на сервер.

#### Запрос
- **Метод**: `POST`
- **Эндпоинт**: `/internal/task`
- **Тело**:
  ```json
  {
    "id": "task-id",
    "result": 5.0
  }
  ```

#### Ответ
- **Статус**: `200 OK`
- **Тело**:
  ```json
  {
    "status": "done"
  }
  ```

---

## Примеры запросов

Вы можете взаимодействовать с сервером с помощью команд `curl`.

### 1. Добавить выражение
```bash
curl -X POST http://localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d '{"expression": "2 + 3 * 4"}'
```

### 2. Получить все выражения
```bash
curl http://localhost:8080/api/v1/expressions
```

### 3. Получить выражение по ID
```bash
curl http://localhost:8080/api/v1/expressions/{expression-id}
```

### 4. Получить задачу для агента
```bash
curl http://localhost:8080/internal/task
```

### 5. Завершить задачу
```bash
curl -X POST http://localhost:8080/internal/task -H "Content-Type: application/json" -d '{"id": "task-id", "result": 14.0}'
```

---

## Как это работает

1. **Сервер** получает математическое выражение от пользователя, разбивает его на отдельные задачи (например, операции сложения, вычитания и т. д.).
2. **Агент** забирает задачи с сервера, выполняет вычисления (производит операции) и отправляет результат обратно на сервер.
3. Сервер отслеживает статус выполнения задач, и как только все задачи в выражении выполнены, вычисляется итоговый результат.
