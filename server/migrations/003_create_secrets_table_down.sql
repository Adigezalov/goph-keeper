-- Откат миграции: удаление таблицы секретов
DROP TRIGGER IF EXISTS update_secrets_updated_at ON secrets;
DROP INDEX IF EXISTS idx_secrets_metadata;
DROP INDEX IF EXISTS idx_secrets_user_id_deleted_at;
DROP INDEX IF EXISTS idx_secrets_deleted_at;
DROP INDEX IF EXISTS idx_secrets_user_id;
DROP TABLE IF EXISTS secrets;

