-- +migrate Up
CREATE TABLE user_api_keys (
    id UUID DEFAULT gen_random_uuid(),
    user_id UUID,
    provider VARCHAR(20) NOT NULL,
    encrypted_key TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    CHECK (provider IN ('openai', 'gemini', 'claude')),
    UNIQUE(provider, user_id)
);
CREATE INDEX idx_user_api_keys_user_id ON user_api_keys(user_id);

-- +migrate Down
DROP TABLE IF EXISTS user_api_keys CASCADE;
 