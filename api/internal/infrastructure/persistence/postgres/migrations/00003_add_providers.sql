-- +migrate Up
CREATE TABLE user_auth_providers (
    user_id UUID,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY(user_id),
    UNIQUE(provider, provider_user_id)
);

-- +migrate Down
DROP TABLE IF EXISTS user_auth_providers CASCADE;
 