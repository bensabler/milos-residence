-- +goose Up
-- +goose StatementBegin
CREATE TABLE reservations (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(255) DEFAULT '',
    last_name VARCHAR(255) DEFAULT '',
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(255) DEFAULT '',
    start_date DATE,
    end_date DATE NOT NULL,
    room_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    processed INTEGER DEFAULT 0,
    CONSTRAINT chk_dates CHECK (start_date < end_date)

);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE reservations;
-- +goose StatementEnd
