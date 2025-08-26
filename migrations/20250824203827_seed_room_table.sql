-- +goose Up
-- +goose StatementBegin
INSERT INTO rooms (room_name)
VALUES 
    ('Golden Haybeam Loft'),
    ('Window Perch Theater'),
    ('Laundry-Basket Nook');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM rooms
WHERE room_name IN ('Golden Haybeam Loft', 'Window Perch Theater', 'Laundry-Basket Nook');
-- +goose StatementEnd
