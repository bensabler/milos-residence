package dbrepo

import (
	"errors"
	"time"

	"github.com/bensabler/milos-residence/internal/models"
)

// Global toggles exclusively for tests.
var (
	ForceAllReservationsErr    bool
	ForceAllNewReservationsErr bool
	ForceUpdateReservationErr  bool
	ForceProcessedUpdateErr    bool
	ForceAllRoomsErr           bool
)

func (m *testDBRepo) AllUsers() bool { return true }

func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// Simulate DB insert error for room 2
	if res.RoomID == 2 {
		return 0, errors.New("insert reservation error")
	}
	return 1, nil
}

func (m *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	// Simulate restriction insert error for room 1000
	if r.RoomID == 1000 {
		return errors.New("insert restriction error")
	}
	return nil
}

func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	// room 2 => force DB error
	if roomID == 2 {
		return false, errors.New("db error")
	}
	// If year=2101, pretend available; otherwise false
	if start.Year() == 2101 {
		return true, nil
	}
	return false, nil
}

func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	if ForceAllRoomsErr {
		return nil, errors.New("all rooms error")
	}
	// Year 2101 => one room available, else none
	if start.Year() == 2101 {
		return []models.Room{{ID: 1, RoomName: "Golden Haybeam Loft"}}, nil
	}
	return []models.Room{}, nil
}

func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	// id > 3 => not found
	if id > 3 {
		return models.Room{}, errors.New("room not found")
	}
	return models.Room{ID: id, RoomName: "Room"}, nil
}

func (m *testDBRepo) GetUserByID(id int) (models.User, error) { return models.User{}, nil }
func (m *testDBRepo) UpdateUser(u models.User) error          { return nil }

func (m *testDBRepo) Authenticate(email, _ string) (int, string, error) {
	// Make auth fail deterministically
	if email == "badlogin@example.com" {
		return 0, "", errors.New("invalid credentials")
	}
	return 1, "", nil
}

func (m *testDBRepo) AllReservations() ([]models.Reservation, error) {
	if ForceAllReservationsErr {
		return nil, errors.New("all reservations error")
	}
	return []models.Reservation{
		{ID: 1, FirstName: "A", LastName: "B"},
	}, nil
}

func (m *testDBRepo) AllNewReservations() ([]models.Reservation, error) {
	if ForceAllNewReservationsErr {
		return nil, errors.New("all new reservations error")
	}
	return []models.Reservation{
		{ID: 2, FirstName: "C", LastName: "D"},
	}, nil
}

func (m *testDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	// Return empty for unknowns to match your handlers’ "OK even when empty" paths.
	return models.Reservation{ID: id}, nil
}

func (m *testDBRepo) UpdateReservation(u models.Reservation) error {
	if ForceUpdateReservationErr {
		return errors.New("update reservation error")
	}
	return nil
}

func (m *testDBRepo) DeleteReservation(id int) error { return nil }

func (m *testDBRepo) UpdateProcessedForReservation(id, processed int) error {
	if ForceProcessedUpdateErr {
		return errors.New("processed update error")
	}
	return nil
}

func (m *testDBRepo) AllRooms() ([]models.Room, error) {
	if ForceAllRoomsErr {
		return nil, errors.New("all rooms error")
	}
	// Ensure calendar tests have at least one room.
	return []models.Room{{ID: 1, RoomName: "Golden Haybeam Loft"}}, nil
}

func (m *testDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	// Keep empty (no reservations), one block to exercise remove loop if present
	return []models.RoomRestriction{
		{ID: 11, StartDate: start.AddDate(0, 0, 4), EndDate: start.AddDate(0, 0, 4), RoomID: roomID, ReservationID: 0},
	}, nil
}

func (m *testDBRepo) InsertBlockForRoom(id int, startDate time.Time) error { return nil }
func (m *testDBRepo) DeleteBlockByID(id int) error                         { return nil }
