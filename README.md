# Распределённый вычислитель арифметических выражений

⚠️ **ВНИМАНИЕ:** Выражения должны содержать пробелы между числами и операторами, например: `2 + 3 * 4`, а не `2+3*4`.

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

---

## Возможные ошибки и их решения

### 1. Ошибка: "Ошибка: сервер вернул статус 500"
#### Причина:
- Возможна ошибка при обработке выражения или вычислениях.
- Ошибка деления на 0.

#### Решение:
- Проверьте корректность входного выражения.
- Убедитесь, что все операции валидны (`+`, `-`, `*`, `/`).

### 2. Ошибка: "Ошибка: сервер вернул статус 404 (No tasks available)"
#### Причина:
- Сервер не имеет доступных задач для агентов.
- Агенты могут запрашивать задачи слишком быстро.

#### Решение:
- Проверьте, были ли добавлены задачи (`POST /api/v1/calculate`).
- Добавьте больше выражений для обработки.
- Убедитесь, что сервер и агент запущены корректно.

### 3. Ошибка: "Ошибка декодирования JSON"
#### Причина:
- Некорректный JSON в теле запроса.

#### Решение:
- Убедитесь, что JSON передаётся в правильном формате.
- Используйте `Content-Type: application/json` в заголовке запроса.

### 4. Ошибка: "Ошибка получения задачи: 500 Internal Server Error"
#### Причина:
- Сервер мог неожиданно завершить работу.

#### Решение:
- Проверьте логи сервера (`server.log`).
- Перезапустите сервер (`go run server.go`).

### 5. Ошибка: "Ошибка вычисления: неизвестная операция"
#### Причина:
- Агент получил оператор, который не поддерживается.

#### Решение:
- Проверьте, что выражения содержат только поддерживаемые операции (`+`, `-`, `*`, `/`).

---

## Как проверить систему

### 1. Добавить выражение
```bash
curl -X POST http://localhost:8080/api/v1/calculate -H "Content-Type: application/json" -d '{"expression": "2 + 3 * 4"}'
```

### 2. Получить все выражения
```bash
curl http://localhost:8080/api/v1/expressions
```

### 3. Получить задачу для агента
```bash
curl http://localhost:8080/internal/task
```

### 4. Завершить задачу
```bash
curl -X POST http://localhost:8080/internal/task -H "Content-Type: application/json" -d '{"id": "task-id", "result": 14.0}'
```

---

## Как это работает

1. **Сервер** получает математическое выражение от пользователя, разбивает его на отдельные задачи (например, операции сложения, вычитания и т. д.).
2. **Агент** забирает задачи с сервера, выполняет вычисления (производит операции) и отправляет результат обратно на сервер.
3. Сервер отслеживает статус выполнения задач, и как только все задачи в выражении выполнены, вычисляется итоговый результат.

