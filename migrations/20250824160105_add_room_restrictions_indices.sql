-- +goose NO TRANSACTION

-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_room_restrictions_start_end
  ON room_restrictions (start_date, end_date);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_room_restrictions_room_id
  ON room_restrictions (room_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_room_restrictions_reservation_id
  ON room_restrictions (reservation_id);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_room_restrictions_reservation_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_room_restrictions_room_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_room_restrictions_start_end;
