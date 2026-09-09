-- +migrate Up
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    sandbox_id VARCHAR(255) NOT NULL DEFAULT '',
    preview_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_projects_user_updated
    ON projects (user_id, updated_at DESC, id DESC);

-- +migrate Down
DROP TABLE IF EXISTS projects CASCADE;
