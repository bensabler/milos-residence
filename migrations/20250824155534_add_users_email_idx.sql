-- +goose NO TRANSACTION
-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_users_email;
