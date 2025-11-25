# Настройка SMTP для Email Верификации

## Обзор

Система регистрации использует email верификацию. После регистрации пользователю отправляется 6-значный код подтверждения, который необходимо ввести для завершения регистрации и получения токенов доступа.

## Параметры

- **Формат кода**: 6-значный цифровой (например, `123456`)
- **Время жизни кода**: 10 минут
- **Повторная отправка**: Доступна через UI, старые коды автоматически аннулируются

## Настройка Yandex SMTP

### 1. Создание пароля приложения

1. Войдите в свой Yandex аккаунт
2. Перейдите на страницу: https://id.yandex.ru/security/app-passwords
3. Нажмите "Создать пароль приложения"
4. Выберите "Почта" или "Другое приложение"
5. Скопируйте сгенерированный пароль

### 2. Настройка переменных окружения

Создайте файл `.env` (или обновите существующий) со следующими параметрами:

```bash
# SMTP настройки для Yandex
SMTP_HOST=smtp.yandex.ru
SMTP_PORT=465
SMTP_USERNAME=your-email@yandex.ru
SMTP_PASSWORD=your-app-password-from-step-1
SMTP_FROM=your-email@yandex.ru
```

### 3. Загрузка переменных окружения

```bash
# Linux/Mac
export $(cat .env | xargs)

# Или просто установите переменные напрямую:
export SMTP_USERNAME="your-email@yandex.ru"
export SMTP_PASSWORD="your-app-password"
export SMTP_FROM="your-email@yandex.ru"
```

## Альтернативные SMTP сервисы

### Gmail

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=465
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password  # Получить в: https://myaccount.google.com/apppasswords
SMTP_FROM=your-email@gmail.com
```

### Mail.ru

```bash
SMTP_HOST=smtp.mail.ru
SMTP_PORT=465
SMTP_USERNAME=your-email@mail.ru
SMTP_PASSWORD=your-password
SMTP_FROM=your-email@mail.ru
```

## Тестирование без SMTP

Если SMTP не настроен (переменные окружения отсутствуют), код подтверждения будет выводиться в логи сервера:

```
[Email] SMTP не настроен. Код подтверждения для user@example.com: 123456
```

Это полезно для разработки и тестирования.

## Запуск миграций

Перед первым запуском примените миграции базы данных:

```bash
cd server
go run cmd/migrate/main.go
```

## Запуск сервера

```bash
cd server
go run cmd/goph-keeper/main.go
```

## Проверка настройки

При запуске сервер выводит информацию о конфигурации:

```
SMTP Host: smtp.yandex.ru:465
Verification Code TTL: 10m0s
```

## Структура таблиц

### users
- `email_verified` - флаг подтверждения email (по умолчанию `false`)

### email_verification_codes
- `user_id` - ID пользователя
- `code` - 6-значный код
- `expires_at` - время истечения кода
- `used` - флаг использования кода

## API Endpoints

### POST /api/v1/user/register
Создает пользователя и отправляет код верификации

### POST /api/v1/user/verify-email
Проверяет код и выдает токены

### POST /api/v1/user/resend-code
Отправляет новый код (аннулирует старые)

### POST /api/v1/user/login
Вход возможен только с подтвержденным email

## Troubleshooting

### Письма не приходят

1. Проверьте правильность SMTP настроек
2. Проверьте, что используется пароль приложения, а не основной пароль
3. Проверьте логи сервера на наличие ошибок
4. Проверьте папку "Спам" в почтовом ящике

### Ошибка "535 5.7.8 Error: authentication failed"

- Используйте пароль приложения вместо основного пароля аккаунта
- Убедитесь, что `SMTP_USERNAME` совпадает с `SMTP_FROM`

### Ошибка подключения к SMTP

- Проверьте, что порт 465 не заблокирован firewall
- Для Gmail может потребоваться включить "Ненадежные приложения"

## Безопасность

1. **Никогда не коммитьте** файлы с реальными паролями в git
2. Используйте **пароли приложений**, а не основные пароли
3. В production используйте **переменные окружения** или секреты (Vault, AWS Secrets Manager)
4. Настройте **rate limiting** для endpoints отправки email

