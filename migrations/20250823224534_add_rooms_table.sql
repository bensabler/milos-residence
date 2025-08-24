-- +goose Up
-- +goose StatementBegin
CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    room_name VARCHAR(255) DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE rooms;
-- +goose StatementEnd
