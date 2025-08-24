-- +goose Up
-- +goose StatementBegin
ALTER TABLE room_restrictions
  ADD CONSTRAINT fk_room_restrictions_room
    FOREIGN KEY (room_id) REFERENCES rooms(id)
    ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT fk_room_restrictions_reservation
    FOREIGN KEY (reservation_id) REFERENCES reservations(id)
    ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT fk_room_restrictions_restriction
    FOREIGN KEY (restriction_id) REFERENCES restrictions(id)
    ON DELETE CASCADE ON UPDATE CASCADE;

CREATE INDEX IF NOT EXISTS idx_room_restrictions_room_id        ON room_restrictions(room_id);
CREATE INDEX IF NOT EXISTS idx_room_restrictions_reservation_id ON room_restrictions(reservation_id);
CREATE INDEX IF NOT EXISTS idx_room_restrictions_restriction_id ON room_restrictions(restriction_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE room_restrictions
  DROP CONSTRAINT IF EXISTS fk_room_restrictions_room,
  DROP CONSTRAINT IF EXISTS fk_room_restrictions_reservation,
  DROP CONSTRAINT IF EXISTS fk_room_restrictions_restriction;

DROP INDEX IF EXISTS idx_room_restrictions_room_id;
DROP INDEX IF EXISTS idx_room_restrictions_reservation_id;
DROP INDEX IF EXISTS idx_room_restrictions_restriction_id;
-- +goose StatementEnd
