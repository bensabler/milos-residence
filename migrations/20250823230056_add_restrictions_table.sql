-- +goose Up
-- +goose StatementBegin
CREATE TABLE restrictions (
    id SERIAL PRIMARY KEY,
    restriction_name VARCHAR(255) DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE restrictions;
-- +goose StatementEnd
