-- +goose Up
-- +goose StatementBegin
CREATE TABLE room_restrictions (
    id SERIAL PRIMARY KEY,
    start_date DATE,
    end_date DATE NOT NULL,
    room_id INTEGER,
    reservation_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    restriction_id INTEGER
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE room_restrictions;
-- +goose StatementEnd
