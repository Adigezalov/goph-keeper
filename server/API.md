# GophKeeper API Documentation

## Base URL
```
http://localhost:8080/api
```

## Authentication
Все endpoints секретов требуют авторизации через Bearer token в заголовке:
```
Authorization: Bearer <access_token>
```

---

## Health Check

### Check Server Health
```http
PATCH /api/v1/health
```

**Response** `200 OK`:
```json
{
  "status": "ok"
}
```

---

## User Endpoints

### Register User
```http
POST /api/v1/user/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response** `200 OK`:
```json
{
  "message": "verification_code_sent"
}
```
*После регистрации на email отправляется 6-значный код подтверждения. Токены не выдаются до подтверждения email.*

### Verify Email
```http
POST /api/v1/user/verify-email
Content-Type: application/json

{
  "email": "user@example.com",
  "code": "123456"
}
```

**Response** `200 OK`:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs..."
}
```
*Refresh token устанавливается в cookie*

### Resend Verification Code
```http
POST /api/v1/user/resend-code
Content-Type: application/json

{
  "email": "user@example.com"
}
```

**Response** `200 OK`:
```json
{
  "message": "verification_code_sent"
}
```
*Отправляет новый код подтверждения на email. Старые коды становятся недействительными.*

### Login User
```http
POST /api/v1/user/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response** `200 OK`:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs..."
}
```
*Refresh token устанавливается в cookie*

**Важно**: Вход возможен только для пользователей с подтвержденным email.

### Refresh Token
```http
GET /api/v1/user/refresh
Cookie: refresh_token=...
```

**Response** `200 OK`:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Logout
```http
GET /api/v1/user/logout
Cookie: refresh_token=...
```

**Response** `200 OK`

### Logout All Devices
```http
GET /api/v1/user/logout-all
Authorization: Bearer <access_token>
```

**Response** `200 OK`

---

## Secrets Endpoints

### Create Secret
```http
POST /api/v1/secrets
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "login": "encrypted_login_base64",
  "password": "encrypted_password_base64",
  "metadata": {
    "app": "github",
    "fileName": "secret.txt"
  },
  "binary_data": "base64_encoded_encrypted_file"
}
```

**Response** `201 Created`:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "login": "encrypted_login_base64",
  "password": "encrypted_password_base64",
  "metadata": {
    "app": "github",
    "fileName": "secret.txt"
  },
  "binary_data": "base64_encoded_encrypted_file",
  "version": 1,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

### Get All Secrets
```http
GET /api/v1/secrets
Authorization: Bearer <access_token>
```

**Response** `200 OK`:
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "login": "encrypted_login_base64",
    "password": "encrypted_password_base64",
    "metadata": {
      "app": "github"
    },
    "version": 1,
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  }
]
```

### Get Secret by ID
```http
GET /api/v1/secrets/{id}
Authorization: Bearer <access_token>
```

**Response** `200 OK`:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "login": "encrypted_login_base64",
  "password": "encrypted_password_base64",
  "metadata": {
    "app": "github"
  },
  "version": 1,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

**Errors**:
- `404 Not Found` - секрет не найден

### Update Secret
```http
PUT /api/v1/secrets/{id}
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "login": "new_encrypted_login_base64",
  "password": "new_encrypted_password_base64",
  "metadata": {
    "app": "gitlab"
  },
  "version": 1
}
```

**Response** `200 OK`:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "login": "new_encrypted_login_base64",
  "password": "new_encrypted_password_base64",
  "metadata": {
    "app": "gitlab"
  },
  "version": 2,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:05:00Z"
}
```

**Errors**:
- `404 Not Found` - секрет не найден
- `409 Conflict` - конфликт версий (секрет был изменен на другом устройстве)

### Delete Secret
```http
DELETE /api/v1/secrets/{id}
Authorization: Bearer <access_token>
```

**Response** `204 No Content`

**Errors**:
- `404 Not Found` - секрет не найден

### Sync Secrets
Получает все секреты для синхронизации. Ключевой endpoint для offline-first архитектуры.

#### Первая синхронизация (все секреты)
```http
GET /api/v1/secrets/sync
Authorization: Bearer <access_token>
```

#### Инкрементальная синхронизация (только измененные)
```http
GET /api/v1/secrets/sync?since=2024-01-15T10:00:00Z
Authorization: Bearer <access_token>
```

**Response** `200 OK`:
```json
{
  "secrets": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "login": "encrypted_login_base64",
      "password": "encrypted_password_base64",
      "metadata": {
        "app": "github"
      },
      "version": 2,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:05:00Z"
    },
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "deleted_at": "2024-01-15T10:10:00Z"
    }
  ],
  "server_time": "2024-01-15T10:20:00Z"
}
```

**Параметры**:
- `since` (optional) - RFC3339 timestamp. Возвращает только секреты, измененные после этого времени.

**Примечания**:
- Без параметра `since` возвращает все активные секреты (первая синхронизация)
- С параметром `since` возвращает созданные, обновленные и удаленные секреты
- `server_time` используется для следующего запроса синхронизации
- Удаленные секреты имеют поле `deleted_at`

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Неверный формат запроса"
}
```

### 401 Unauthorized
```json
{
  "error": "Ошибка авторизации"
}
```

### 404 Not Found
```json
{
  "error": "Секрет не найден"
}
```

### 409 Conflict
```json
{
  "error": "Конфликт версий: секрет был изменен на другом устройстве"
}
```

### 500 Internal Server Error
```json
{
  "error": "Внутренняя ошибка сервера"
}
```

---

## Data Encryption

### Client-Side Encryption
- Все чувствительные данные (`login`, `password`, `binary_data`) шифруются **на клиенте**
- Сервер хранит зашифрованные данные и не имеет доступа к ключу шифрования
- Используется AES-GCM 256-bit шифрование
- Каждое поле шифруется отдельно с уникальным IV

### Unencrypted Fields
- `metadata` - НЕ шифруется, используется для поиска и фильтрации
- `version`, `created_at`, `updated_at` - служебные поля

---

## Synchronization Strategy

### First Sync
```
Client → GET /api/v1/secrets/sync
Server → Returns all active secrets
Client → Saves to IndexedDB
Client → Stores server_time for next sync
```

### Incremental Sync
```
Client → GET /api/v1/secrets/sync?since=<last_server_time>
Server → Returns created/updated/deleted secrets since timestamp
Client → Merges changes to IndexedDB
Client → Updates last_server_time
```

### Conflict Resolution
При конфликте версий (`409 Conflict`):
1. Сервер возвращает актуальную версию
2. Клиент сохраняет обе версии (Keep Both)
3. Пользователь разрешает конфликт вручную

---

## Examples

### Complete Flow

```bash
# 1. Register
curl -X POST http://localhost:8080/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "pass123"}' \
  -c cookies.txt

# 2. Create Secret
curl -X POST http://localhost:8080/api/v1/secrets \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "login": "encrypted_login",
    "password": "encrypted_password",
    "metadata": {"app": "github"}
  }'

# 3. Get All Secrets
curl -X GET http://localhost:8080/api/v1/secrets \
  -H "Authorization: Bearer <access_token>"

# 4. Sync (first time)
curl -X GET http://localhost:8080/api/v1/secrets/sync \
  -H "Authorization: Bearer <access_token>"

# 5. Sync (incremental)
curl -X GET "http://localhost:8080/api/v1/secrets/sync?since=2024-01-15T10:00:00Z" \
  -H "Authorization: Bearer <access_token>"

# 6. Update Secret
curl -X PUT http://localhost:8080/api/v1/secrets/{id} \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "login": "new_encrypted_login",
    "password": "new_encrypted_password",
    "version": 1
  }'

# 7. Delete Secret
curl -X DELETE http://localhost:8080/api/v1/secrets/{id} \
  -H "Authorization: Bearer <access_token>"
```

