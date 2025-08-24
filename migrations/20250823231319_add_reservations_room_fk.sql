-- +goose Up
-- +goose StatementBegin
ALTER TABLE reservations
  ADD CONSTRAINT fk_reservations_room
  FOREIGN KEY (room_id)
  REFERENCES rooms(id)
  ON DELETE CASCADE
  ON UPDATE CASCADE;

-- index the referencing column for faster deletes/updates on rooms
CREATE INDEX IF NOT EXISTS idx_reservations_room_id ON reservations(room_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE reservations
  DROP CONSTRAINT IF EXISTS fk_reservations_room;

DROP INDEX IF EXISTS idx_reservations_room_id;
-- +goose StatementEnd
