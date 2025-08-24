-- +goose Up
-- +goose StatementBegin
INSERT INTO rooms (room_name)
VALUES ('Golden Haybeam Loft');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM rooms
WHERE room_name = 'Golden Haybeam Loft';
-- +goose StatementEnd
