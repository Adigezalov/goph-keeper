-- Удаление индекса
DROP INDEX IF EXISTS idx_users_email_verified;

-- Удаление поля email_verified из таблицы users
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
