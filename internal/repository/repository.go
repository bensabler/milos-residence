// Package repository defines interfaces for data access operations.
package repository

import (
	"time"

	"github.com/bensabler/milos-residence/internal/models"
)

// DatabaseRepo defines the interface for all database operations.
// Implementations provide data access for users, reservations, rooms, and restrictions.
type DatabaseRepo interface {
	// AllUsers returns true if database connection is healthy.
	AllUsers() bool

	// InsertReservation creates a new reservation record.
	// Returns the generated reservation ID.
	InsertReservation(res models.Reservation) (int, error)

	// InsertRoomRestriction creates a room restriction record.
	InsertRoomRestriction(r models.RoomRestriction) error

	// SearchAvailabilityByDatesByRoomID checks if a specific room is available for the given dates.
	SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error)

	// SearchAvailabilityForAllRooms returns all rooms available for the given dates.
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)

	// GetRoomByID retrieves a room by its ID.
	GetRoomByID(id int) (models.Room, error)

	// GetUserByID retrieves a user by their ID.
	GetUserByID(id int) (models.User, error)

	// UpdateUser modifies an existing user record.
	UpdateUser(u models.User) error

	// Authenticate verifies user credentials.
	// Returns user ID and password hash on success.
	Authenticate(email, testPassword string) (int, string, error)

	// AllReservations retrieves all reservation records.
	AllReservations() ([]models.Reservation, error)

	// AllNewReservations retrieves unprocessed reservation records.
	AllNewReservations() ([]models.Reservation, error)

	// GetReservationByID retrieves a reservation by its ID.
	GetReservationByID(id int) (models.Reservation, error)

	// UpdateReservation modifies an existing reservation record.
	UpdateReservation(u models.Reservation) error

	// DeleteReservation removes a reservation record.
	DeleteReservation(id int) error

	// UpdateProcessedForReservation updates the processed status of a reservation.
	UpdateProcessedForReservation(id, processed int) error

	// AllRooms retrieves all room records.
	AllRooms() ([]models.Room, error)

	// GetRestrictionsForRoomByDate retrieves room restrictions overlapping the given date range.
	GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error)

	// InsertBlockForRoom creates an owner block restriction for a room.
	InsertBlockForRoom(id int, startDate time.Time) error

	// DeleteBlockByID removes a room restriction by its ID.
	DeleteBlockByID(id int) error
}
