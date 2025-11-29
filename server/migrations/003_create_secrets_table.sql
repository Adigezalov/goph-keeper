-- Включаем расширение для работы с UUID (если еще не включено)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Создание таблицы секретов
CREATE TABLE IF NOT EXISTS secrets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    login TEXT NOT NULL,                     -- Логин (base64)
    password TEXT NOT NULL,                  -- Пароль (base64)
    metadata JSONB,                          -- Метаданные (fileName, app и т.д.)
    binary_data BYTEA,                       -- Бинарные данные (файлы)
    version INTEGER NOT NULL DEFAULT 1,      -- Версия для разрешения конфликтов синхронизации
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE      -- Soft delete (NULL = активная запись)
);

-- Создание индексов для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_secrets_user_id ON secrets(user_id);
CREATE INDEX IF NOT EXISTS idx_secrets_deleted_at ON secrets(deleted_at);
CREATE INDEX IF NOT EXISTS idx_secrets_user_id_deleted_at ON secrets(user_id, deleted_at);

-- Индекс для поиска по метаданным (например, по типу файла или приложению)
CREATE INDEX IF NOT EXISTS idx_secrets_metadata ON secrets USING GIN (metadata);

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_secrets_updated_at 
    BEFORE UPDATE ON secrets 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

