-- Добавление поля email_verified в таблицу users
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT FALSE NOT NULL;

-- Создание индекса для быстрого поиска неподтвержденных пользователей
CREATE INDEX IF NOT EXISTS idx_users_email_verified ON users(email_verified);
