-- +migrate Up
CREATE EXTENSION IF NOT EXISTS "citext";

-- +migrate Down
DROP EXTENSION IF EXISTS "citext";