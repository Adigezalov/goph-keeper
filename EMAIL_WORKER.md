# Email Worker - Асинхронная отправка email

## Обзор

Email Worker - это отдельный сервис, работающий в отдельной горутине для асинхронной обработки задач отправки email. Это решение обеспечивает неблокирующую отправку писем и повышает производительность основного приложения.

## Архитектура

```
┌─────────────┐
│   User      │
│  Service    │
└─────┬───────┘
      │ SendEmail(to, code)
      │ (неблокирующий вызов)
      ▼
┌─────────────────────┐
│  EmailWorker        │
│   Service           │
└─────┬───────────────┘
      │ добавление в очередь
      ▼
┌─────────────────────┐
│   Job Queue         │ ◄───── Буферизированный канал (100 задач)
│   (channel)         │
└─────┬───────────────┘
      │
      ▼
┌─────────────────────┐
│   Worker            │ ◄───── Отдельная горутина
│   (goroutine)       │
└─────┬───────────────┘
      │ обработка задач
      ▼
┌─────────────────────┐
│   SMTP Service      │ ◄───── Реальная отправка email
│   (Yandex)          │
└─────────────────────┘
```

## Компоненты

### 1. EmailWorker (`internal/emailworker/worker.go`)

Основной воркер, управляющий очередью и обработкой задач.

**Параметры:**
- `queueSize` - размер буфера очереди (по умолчанию 100)
- `maxRetries` - максимальное количество попыток (по умолчанию 3)
- `retryDelay` - задержка между попытками (по умолчанию 5 секунд)

**Методы:**
- `Start()` - запускает воркер в отдельной горутине
- `Stop()` - корректно останавливает воркер, обрабатывая оставшиеся задачи
- `SendEmail(to, code)` - добавляет задачу в очередь (неблокирующий)
- `QueueLength()` - возвращает текущую длину очереди

### 2. EmailWorkerService (`internal/emailworker/service.go`)

Адаптер для интеграции с `user.Service`, реализует интерфейс `EmailService`.

**Методы:**
- `GenerateVerificationCode()` - генерирует 6-значный код
- `SendEmail(toEmail, code)` - отправляет задачу в воркер

### 3. SMTP Service (`internal/email/service.go`)

Низкоуровневый сервис для фактической отправки email через SMTP.

## Поведение

### Нормальная работа

```go
// 1. Регистрация пользователя
user.RegisterUser(req)
  ↓
// 2. Генерация кода
code := emailService.GenerateVerificationCode()
  ↓
// 3. Сохранение кода в БД
verificationRepo.CreateVerificationCode(code)
  ↓
// 4. Асинхронная отправка (неблокирующая)
emailService.SendEmail(email, code) // возвращается сразу
  ↓
// 5. Ответ пользователю
return success
```

### Обработка в воркере (отдельная горутина)

```go
// Воркер получает задачу из канала
job := <-jobQueue
  ↓
// Попытка отправки
err := smtpService.SendVerificationCode(job.To, job.Code)
  ↓
// Если ошибка и есть попытки
if err && job.Attempt < maxRetries {
    // Ждем retryDelay
    time.Sleep(5 * time.Second)
    // Повторная отправка
    SendEmail(job.To, job.Code)
}
```

### Graceful Shutdown

При остановке сервера:

1. Вызывается `emailWorker.Stop()`
2. Закрывается канал очереди (новые задачи не принимаются)
3. Обрабатываются все оставшиеся задачи в очереди
4. Завершается работа горутины
5. Возвращается управление

## Преимущества

### ✅ Неблокирующая работа
Регистрация пользователя не ждет отправки email:
```go
// Быстро возвращает результат
user.RegisterUser() // ~50ms (без ожидания SMTP)
// vs старый способ
user.RegisterUser() // ~500-1000ms (ждет SMTP)
```

### ✅ Retry механизм
Автоматические повторные попытки при сбоях:
```
Попытка 1: Ошибка подключения → Повтор через 5 секунд
Попытка 2: Таймаут → Повтор через 5 секунд
Попытка 3: Успешно отправлено
```

### ✅ Масштабируемость
Можно легко увеличить количество воркеров:
```go
// Запускаем 5 воркеров для обработки
for i := 0; i < 5; i++ {
    worker := emailworker.NewWorker(smtpService, 100, 3, 5*time.Second)
    worker.Start()
}
```

### ✅ Изоляция
SMTP сбои не влияют на основное приложение:
- Даже если SMTP сервер недоступен, регистрация работает
- Задачи остаются в очереди
- Логируются ошибки для мониторинга

### ✅ Graceful Shutdown
Корректная остановка без потери задач:
- Обрабатываются все задачи в очереди
- Не прерываются отправки в процессе
- Контролируемое завершение горутин

## Конфигурация

В `main.go`:

```go
// Создаем email worker
// Параметры: queueSize=100, maxRetries=3, retryDelay=5s
emailWorker := emailworker.NewWorker(smtpService, 100, 3, 5*time.Second)
emailWorker.Start()
defer emailWorker.Stop()

// Создаем сервис для интеграции
emailService := emailworker.NewEmailWorkerService(emailWorker)
```

### Настройка параметров

```go
// Большая очередь для высоконагруженных систем
emailWorker := emailworker.NewWorker(smtpService, 1000, 5, 10*time.Second)

// Быстрые повторные попытки
emailWorker := emailworker.NewWorker(smtpService, 100, 10, 1*time.Second)

// Без повторов (только одна попытка)
emailWorker := emailworker.NewWorker(smtpService, 100, 0, 0)
```

## Логирование

Worker логирует все важные события:

```
[EmailWorker] Worker запущен
[EmailWorker] Задача добавлена в очередь: user@example.com
[Email] Код подтверждения отправлен на user@example.com
[EmailWorker] Email успешно отправлен на user@example.com
```

При ошибках:

```
[EmailWorker] Ошибка отправки email на user@example.com (попытка 1): connection timeout
[EmailWorker] Повторная попытка через 5s
[EmailWorker] Email успешно отправлен на user@example.com
```

При превышении попыток:

```
[EmailWorker] Ошибка отправки email на user@example.com (попытка 3): connection refused
[EmailWorker] Превышено максимальное количество попыток для user@example.com
```

## Мониторинг

### Получение длины очереди

```go
queueLen := emailWorker.QueueLength()
logger.Infof("Задач в очереди: %d", queueLen)
```

### Добавление метрик (опционально)

```go
// В worker.go можно добавить счетчики
type Worker struct {
    // ...
    successCount int64
    failureCount int64
}

// И методы для мониторинга
func (w *Worker) GetStats() (success, failure int64) {
    return atomic.LoadInt64(&w.successCount), 
           atomic.LoadInt64(&w.failureCount)
}
```

## Тестирование

### Mock для тестов

```go
type MockEmailSender struct {
    Calls []struct{ To, Code string }
}

func (m *MockEmailSender) SendVerificationCode(to, code string) error {
    m.Calls = append(m.Calls, struct{ To, Code string }{to, code})
    return nil
}

// В тестах
mockSender := &MockEmailSender{}
worker := emailworker.NewWorker(mockSender, 10, 3, 100*time.Millisecond)
worker.Start()
defer worker.Stop()

// Отправляем задачу
service := emailworker.NewEmailWorkerService(worker)
service.SendEmail("test@example.com", "123456")

// Ждем обработки
time.Sleep(200 * time.Millisecond)

// Проверяем
assert.Equal(t, 1, len(mockSender.Calls))
assert.Equal(t, "test@example.com", mockSender.Calls[0].To)
```

## Troubleshooting

### Очередь переполнена

```
[EmailWorker] Очередь заполнена, задача отклонена
```

**Решение:** Увеличьте размер очереди:
```go
emailWorker := emailworker.NewWorker(smtpService, 1000, 3, 5*time.Second)
```

### Много неудачных попыток

```
[EmailWorker] Превышено максимальное количество попыток для ...
```

**Решение:** Проверьте SMTP настройки и увеличьте retryDelay:
```go
emailWorker := emailworker.NewWorker(smtpService, 100, 3, 10*time.Second)
```

### Медленная обработка

**Решение:** Запустите несколько воркеров (будет в следующей версии) или оптимизируйте SMTP подключение (используйте connection pooling).

## Будущие улучшения

- [ ] Поддержка нескольких воркеров (worker pool)
- [ ] Приоритетная очередь
- [ ] Персистентная очередь (Redis/DB) для надежности
- [ ] Метрики и мониторинг (Prometheus)
- [ ] Rate limiting
- [ ] Batch отправка
- [ ] Circuit breaker для SMTP

