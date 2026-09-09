-- +migrate Up
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL,
    mode VARCHAR(20) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX idx_messages_project_created
    ON messages (project_id, created_at ASC, id ASC);

-- +migrate Down
DROP TABLE IF EXISTS messages CASCADE;
