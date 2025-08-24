-- +goose Up
-- +goose StatementBegin
INSERT INTO restrictions (restriction_name)
SELECT 'Reservation'
WHERE NOT EXISTS (
  SELECT 1 FROM restrictions WHERE restriction_name = 'Reservation'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM restrictions
WHERE restriction_name = 'Reservation';
-- +goose StatementEnd
