-- Создание таблицы кодов подтверждения email
CREATE TABLE IF NOT EXISTS email_verification_codes (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(6) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE NOT NULL
);

-- Создание индекса для быстрого поиска по user_id
CREATE INDEX IF NOT EXISTS idx_verification_codes_user_id ON email_verification_codes(user_id);

-- Создание индекса для быстрого поиска по коду
CREATE INDEX IF NOT EXISTS idx_verification_codes_code ON email_verification_codes(code);

-- Создание индекса для быстрого поиска активных кодов
CREATE INDEX IF NOT EXISTS idx_verification_codes_active ON email_verification_codes(user_id, used, expires_at);

