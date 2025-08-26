-- +goose Up
-- +goose StatementBegin
INSERT INTO restrictions (restriction_name)
SELECT v.name
FROM (VALUES ('Reservation'), ('Owner Block')) AS v(name)
WHERE NOT EXISTS (
  SELECT 1 FROM restrictions r WHERE r.restriction_name = v.name
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM restrictions
WHERE restriction_name IN ('Reservation','Owner Block');
-- +goose StatementEnd
