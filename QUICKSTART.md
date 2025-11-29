# 🚀 Быстрый старт Goph-Keeper

Запустите проект за **5 минут**! Этот гайд покажет самый быстрый способ начать работу.

## Предварительные требования

- Go 1.25+
- PostgreSQL 14+
- Node.js 20+ или Bun
- Git

## 1️⃣ Клонирование и установка (1 мин)

```bash
# Клонируем репозиторий
git clone https://github.com/Adigezalov/goph-keeper.git
cd goph-keeper
```

## 2️⃣ Настройка базы данных (2 мин)

```bash
# Запускаем PostgreSQL (если еще не запущен)
# macOS
brew services start postgresql@14

# Linux
sudo systemctl start postgresql

# Создаем базу данных
psql postgres -c "CREATE USER keeper_user WITH PASSWORD 'keeper_password';"
psql postgres -c "CREATE DATABASE goph_keeper OWNER keeper_user;"
psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE goph_keeper TO keeper_user;"
```

## 3️⃣ Запуск сервера (1 мин)

```bash
cd server

# Устанавливаем зависимости
go mod download

# Настройка переменных окружения (одной командой)
export DATABASE_URI="postgres://keeper_user:keeper_password@localhost:5432/goph_keeper?sslmode=disable"
export JWT_SECRET="dev-secret-key-change-in-production-12345"

# Применяем миграции
go run cmd/migrate/main.go

# Запускаем сервер
go run cmd/goph-keeper/main.go
```

Сервер запустится на `http://localhost:8080` ✅

## 4️⃣ Запуск клиента (1 мин)

**Откройте новый терминал:**

```bash
cd client

# С bun (рекомендуется - быстрее)
bun install
bun run dev

# Или с npm
npm install
npm run dev
```

Клиент запустится на `http://localhost:3000` ✅

## 🎉 Готово!

Теперь откройте браузер:
- **Приложение:** http://localhost:3000
- **API (Swagger):** http://localhost:8080/swagger/index.html

## 📝 Первый запуск

1. Откройте http://localhost:3000
2. Нажмите **"Регистрация"**
3. Введите email и пароль
4. Смотрите логи сервера - там будет **6-значный код**:
   ```
   [Email] SMTP не настроен. Код подтверждения для test@example.com: 123456
   ```
5. Введите этот код в форму
6. Готово! Вы в системе 🎉

## ⚙️ Опциональная настройка SMTP

Без SMTP коды выводятся в логи (удобно для dev). Для production настройте SMTP:

```bash
export SMTP_HOST="smtp.yandex.ru"
export SMTP_PORT="465"
export SMTP_USERNAME="your-email@yandex.ru"
export SMTP_PASSWORD="your-app-password"
export SMTP_FROM="your-email@yandex.ru"
```

Подробнее: [server/SMTP_SETUP.md](SMTP_SETUP.md)

## 🐛 Проблемы?

### Порт 8080 занят?
```bash
# Используйте другой порт
export RUN_ADDRESS=":8081"
go run cmd/goph-keeper/main.go
```

### PostgreSQL не подключается?
```bash
# Проверьте статус
brew services list  # macOS
systemctl status postgresql  # Linux

# Проверьте подключение
psql -U keeper_user -d goph_keeper -h localhost
```

### Ошибка миграций?
```bash
# Убедитесь, что база создана
psql postgres -c "\l" | grep goph_keeper

# Проверьте права
psql postgres -c "\du" | grep keeper_user
```

## 📚 Что дальше?

- 📖 [Полная документация](README.md)
- 📧 [Настройка SMTP](SMTP_SETUP.md)
- 🧪 [Тестирование](server/TESTING.md)
- 🤝 [Contributing](CONTRIBUTING.md)

## 🎯 Основные команды

### Сервер
```bash
cd server

# Запуск
go run cmd/goph-keeper/main.go

# Тесты
go test ./...

# Тесты с coverage
go test ./... -cover

# Миграции вверх
go run cmd/migrate/main.go

# Swagger обновление
swag init -g cmd/goph-keeper/main.go --output ./docs
```

### Клиент
```bash
cd client

# Разработка
bun run dev
# или
npm run dev

# Сборка production
bun run build
# или
npm run build

# Линтер
bun run lint
# или
npm run lint
```

## 🔑 Тестовые данные

Для быстрого тестирования:
- **Email:** test@example.com
- **Password:** password123
- **Код из логов:** смотрите в консоли сервера

## 🚀 Production деплой

Для production не забудьте:

1. **Изменить JWT_SECRET** (минимум 32 символа):
   ```bash
   openssl rand -base64 32
   ```

2. **Настроить SMTP** для email (см. [SMTP_SETUP.md](SMTP_SETUP.md))

3. **Использовать HTTPS**

4. **PostgreSQL с SSL:**
   ```bash
   export DATABASE_URI="postgres://user:pass@host:5432/db?sslmode=require"
   ```

5. **Настроить CORS** для вашего домена

## ⏱️ Таймлайн быстрого старта

- ⏰ **0-1 мин:** Клонирование репозитория
- ⏰ **1-3 мин:** Настройка PostgreSQL
- ⏰ **3-4 мин:** Запуск сервера
- ⏰ **4-5 мин:** Запуск клиента
- ✅ **5 мин:** Работает!

## 💡 Полезные ссылки

- 🌐 **Frontend:** http://localhost:3000
- 🔌 **Backend API:** http://localhost:8080
- 📚 **Swagger UI:** http://localhost:8080/swagger/index.html
- 🏥 **Health Check:** http://localhost:8080/api/v1/health

---

**Приятного использования Goph-Keeper!** 🎉

Возникли вопросы? Создайте [issue на GitHub](https://github.com/Adigezalov/goph-keeper/issues)

