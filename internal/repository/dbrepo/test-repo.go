// internal/repository/dbrepo/test-repo.go

package dbrepo

import (
	"errors"
	"time"

	"github.com/bensabler/milos-residence/internal/models"
)

// Test-only toggles used to force specific error paths in handlers.
// Keep these simple & explicit so tests can flip them safely.
var (
	ForceAllReservationsErr        bool // AllReservations() returns error
	ForceAllNewReservationsErr     bool // AllNewReservations() returns error
	ForceUpdateReservationErr      bool // UpdateReservation(...) returns error
	ForceProcessedUpdateErr        bool // UpdateProcessedForReservation(...) returns error
	ForceAllRoomsErr               bool // AllRooms() returns error
	ForceGetReservationErr         bool // GetReservationByID(...) returns error
	ForceRestrictionsErr           bool // GetRestrictionsForRoomByDate(...) returns error
	ForceSearchAvailabilityErrOn   int  // if non-zero & roomID==this, SearchAvailabilityByDatesByRoomID returns error
	ForceHasReservationRestriction bool // cause GetRestrictionsForRoomByDate to return at least one Reservation restriction
	ForceInsertBlockErr            bool // force errors in block add paths for AdminPostReservationsCalendar
	ForceDeleteBlockErr            bool // force errors in block remove paths for AdminPostReservationsCalendar

)

func (m *testDBRepo) AllUsers() bool { return true }

// InsertReservation returns an error for RoomID == 2 to cover that branch.
func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	if res.RoomID == 2 {
		return 0, errors.New("insert reservation error")
	}
	return 1, nil
}

// InsertRoomRestriction returns an error for a valid room ID so we can reach
// the restriction insert branch in PostReservation after passing GetRoomByID.
func (m *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	if r.RoomID == 3 {
		return errors.New("insert restriction error")
	}
	return nil
}

// SearchAvailabilityByDatesByRoomID supports three branches:
// 1) Forced DB error when ForceSearchAvailabilityErrOn matches the roomID (if set)
// 2) DB error when roomID == 2 (legacy switch kept for backward compatibility)
// 3) Availability true for year 2101, false otherwise (2100 used in tests for "no")
func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	if ForceSearchAvailabilityErrOn != 0 && roomID == ForceSearchAvailabilityErrOn {
		return false, errors.New("db error")
	}
	if roomID == 2 {
		return false, errors.New("db error")
	}
	if start.Year() == 2101 {
		return true, nil
	}
	return false, nil
}

// SearchAvailabilityForAllRooms returns:
// - error if ForceAllRoomsErr is set
// - a single room when start.Year()==2101 (used to render choose-room)
// - empty slice otherwise (used to trigger "No availability")
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	if ForceAllRoomsErr {
		return nil, errors.New("all rooms error")
	}
	if start.Year() == 2101 {
		return []models.Room{{ID: 1, RoomName: "Golden Haybeam Loft"}}, nil
	}
	return []models.Room{}, nil
}

// GetRoomByID returns a stub room unless id > 3, which triggers the "not found" path.
func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	if id > 3 {
		return models.Room{}, errors.New("room not found")
	}
	return models.Room{ID: id, RoomName: "Room"}, nil
}

func (m *testDBRepo) GetUserByID(id int) (models.User, error) { return models.User{}, nil }
func (m *testDBRepo) UpdateUser(u models.User) error          { return nil }

// Authenticate returns an error for a known-bad email to exercise the auth failure branch.
func (m *testDBRepo) Authenticate(email, _ string) (int, string, error) {
	if email == "badlogin@example.com" {
		return 0, "", errors.New("invalid credentials")
	}
	return 1, "", nil
}

// AllReservations returns a small list unless forced to error.
func (m *testDBRepo) AllReservations() ([]models.Reservation, error) {
	if ForceAllReservationsErr {
		return nil, errors.New("all reservations error")
	}
	return []models.Reservation{{ID: 1, FirstName: "A", LastName: "B"}}, nil
}

// AllNewReservations returns a small list unless forced to error.
func (m *testDBRepo) AllNewReservations() ([]models.Reservation, error) {
	if ForceAllNewReservationsErr {
		return nil, errors.New("all new reservations error")
	}
	return []models.Reservation{{ID: 2, FirstName: "C", LastName: "D"}}, nil
}

// GetReservationByID returns an error when forced, else a minimal reservation.
func (m *testDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	if ForceGetReservationErr {
		return models.Reservation{}, errors.New("get reservation error")
	}
	return models.Reservation{ID: id}, nil
}

// UpdateReservation returns an error when forced to exercise the 500 path.
func (m *testDBRepo) UpdateReservation(u models.Reservation) error {
	if ForceUpdateReservationErr {
		return errors.New("update reservation error")
	}
	return nil
}

func (m *testDBRepo) DeleteReservation(id int) error { return nil }

// UpdateProcessedForReservation returns an error when forced, used by the process handler.
func (m *testDBRepo) UpdateProcessedForReservation(id, processed int) error {
	if ForceProcessedUpdateErr {
		return errors.New("processed update error")
	}
	return nil
}

// AllRooms returns one room unless forced to error; this seeds calendar maps during tests.
func (m *testDBRepo) AllRooms() ([]models.Room, error) {
	if ForceAllRoomsErr {
		return nil, errors.New("all rooms error")
	}
	return []models.Room{{ID: 1, RoomName: "Golden Haybeam Loft"}}, nil
}

// GetRestrictionsForRoomByDate returns one block by default so add/remove loops have data to scan.
// Can be forced to error to cover that branch.
func (m *testDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	if ForceRestrictionsErr {
		return nil, errors.New("restrictions error")
	}
	// Always include a block so delete/keep loops have work.
	res := []models.RoomRestriction{
		{ID: 11, StartDate: start.AddDate(0, 0, 4), EndDate: start.AddDate(0, 0, 4), RoomID: roomID, ReservationID: 0},
	}
	// Optionally include an actual reservation restriction to hit reservationMap code path.
	if ForceHasReservationRestriction {
		res = append(res, models.RoomRestriction{
			ID:            42,
			StartDate:     start.AddDate(0, 0, 1),
			EndDate:       start.AddDate(0, 0, 3),
			RoomID:        roomID,
			ReservationID: 777,
			RestrictionID: 1,
		})
	}
	return res, nil
}

func (m *testDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	if ForceInsertBlockErr {
		return errors.New("insert block error")
	}
	return nil
}

func (m *testDBRepo) DeleteBlockByID(id int) error {
	if ForceDeleteBlockErr {
		return errors.New("delete block error")
	}
	return nil
}
