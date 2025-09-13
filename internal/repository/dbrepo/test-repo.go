// Package dbrepo provides database repository implementations for Milo's Residence.
// This file contains the test repository implementation that provides controlled,
// predictable responses for unit testing without requiring an actual database connection.
//
// The test repository uses a toggle-based system where global boolean variables
// can be set to force specific error conditions, allowing comprehensive testing
// of error handling paths in the application's handler and business logic layers.
//
// Design Pattern: Test Double (specifically a Test Stub) - provides canned responses
// Design Pattern: Configurable Stub - behavior can be modified via toggle variables
package dbrepo

import (
	"errors"
	"time"

	"github.com/bensabler/milos-residence/internal/models"
)

// Global toggle variables control test repository behavior to enable comprehensive error path testing.
// These variables should be set to true before tests that need to exercise specific error conditions,
// and reset to false afterward to ensure test isolation and prevent interference between tests.
//
// Usage pattern in tests:
//
//	dbrepo.ForceAllReservationsErr = true
//	defer func() { dbrepo.ForceAllReservationsErr = false }()
//
// This approach enables testing of error handling without complex mocking frameworks
// while maintaining simplicity and readability in test code.
var (
	// ForceAllReservationsErr causes AllReservations() to return an error.
	// Used to test error handling in administrative reservation listing functionality.
	ForceAllReservationsErr bool

	// ForceAllNewReservationsErr causes AllNewReservations() to return an error.
	// Used to test error handling when retrieving unprocessed reservations for staff review.
	ForceAllNewReservationsErr bool

	// ForceUpdateReservationErr causes UpdateReservation() to return an error.
	// Used to test error handling during reservation modification operations.
	ForceUpdateReservationErr bool

	// ForceProcessedUpdateErr causes UpdateProcessedForReservation() to return an error.
	// Used to test error handling when marking reservations as processed/unprocessed.
	ForceProcessedUpdateErr bool

	// ForceAllRoomsErr causes AllRooms() to return an error.
	// Used to test error handling in room listing and calendar functionality.
	ForceAllRoomsErr bool

	// ForceGetReservationErr causes GetReservationByID() to return an error.
	// Used to test error handling when retrieving specific reservation details.
	ForceGetReservationErr bool

	// ForceRestrictionsErr causes GetRestrictionsForRoomByDate() to return an error.
	// Used to test error handling in calendar and availability checking functionality.
	ForceRestrictionsErr bool

	// ForceSearchAvailabilityErrOn specifies a room ID that will trigger errors in SearchAvailabilityByDatesByRoomID.
	// When non-zero and the roomID parameter matches this value, the method returns an error.
	// Used to test error handling in room-specific availability checking.
	ForceSearchAvailabilityErrOn int

	// ForceHasReservationRestriction causes GetRestrictionsForRoomByDate to return reservation restrictions.
	// When true, the method includes reservation-type restrictions (ReservationID > 0) in its results,
	// enabling testing of reservation vs. owner-block distinction in calendar displays.
	ForceHasReservationRestriction bool

	// ForceInsertBlockErr causes InsertBlockForRoom() to return an error.
	// Used to test error handling when administrators add room blocks through the calendar interface.
	ForceInsertBlockErr bool

	// ForceDeleteBlockErr causes DeleteBlockByID() to return an error.
	// Used to test error handling when administrators remove room blocks through the calendar interface.
	ForceDeleteBlockErr bool
)

// AllUsers is a placeholder method that always returns true for basic connectivity testing.
// This method was implemented during development for simple database interaction verification
// and currently serves as a minimal health check operation in the test environment.
//
// In production repository implementations, this would typically return actual user data,
// user counts, or perform more meaningful user-related operations.
func (m *testDBRepo) AllUsers() bool {
	return true
}

// InsertReservation creates a mock reservation and returns a predictable ID.
// This method simulates the database insertion process by returning a consistent ID value
// while providing controlled error scenarios for comprehensive testing of reservation creation workflows.
//
// Test behavior patterns:
//   - RoomID 2: Returns error to test reservation insertion failure handling
//   - All other RoomIDs: Returns ID 1 to simulate successful creation
//
// The error condition (RoomID 2) enables testing of:
//   - Database connection failure recovery
//   - Transaction rollback scenarios
//   - User error messaging for reservation creation failures
//   - Workflow cleanup when reservation creation fails
//
// Parameters:
//   - res: Reservation model containing guest and booking details
//
// Returns:
//   - int: Mock reservation ID (1) for successful operations, 0 for errors
//   - error: Simulated database error when res.RoomID == 2, nil otherwise
func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// Simulate insertion failure for specific room ID to enable error path testing
	if res.RoomID == 2 {
		return 0, errors.New("insert reservation error")
	}

	// Return consistent success response for all other cases
	return 1, nil
}

// InsertRoomRestriction creates a mock room restriction record with controlled error scenarios.
// This method simulates the database insertion of room restrictions (reservations and owner blocks)
// while providing specific error conditions for testing restriction creation workflows.
//
// Room restrictions are critical for availability management as they prevent double-booking
// and coordinate room availability across reservation and administrative blocking systems.
//
// Test behavior patterns:
//   - RoomID 3: Returns error to test restriction insertion failure after successful reservation creation
//   - All other RoomIDs: Returns nil to simulate successful restriction creation
//
// The error condition (RoomID 3) specifically enables testing of the scenario where
// reservation insertion succeeds but restriction creation fails, requiring proper
// error handling and potential rollback of the reservation record.
//
// Parameters:
//   - r: RoomRestriction model containing date range, room, and restriction details
//
// Returns:
//   - error: Simulated database error when r.RoomID == 3, nil otherwise
func (m *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	// Simulate restriction insertion failure to test partial success scenarios
	if r.RoomID == 3 {
		return errors.New("insert restriction error")
	}

	return nil
}

// SearchAvailabilityByDatesByRoomID simulates room availability checking with multiple test scenarios.
// This method provides controlled availability responses and error conditions to enable comprehensive
// testing of room booking workflows, availability validation, and error handling patterns.
//
// The method implements three distinct test scenarios based on input parameters:
//
//  1. **Forced Database Errors**: When ForceSearchAvailabilityErrOn is set to a non-zero room ID
//     and matches the roomID parameter, simulates database connectivity or query failures.
//
//  2. **Legacy Error Condition**: RoomID 2 always returns a database error, maintained for
//     backward compatibility with existing tests that rely on this specific error pattern.
//
// 3. **Date-Based Availability Logic**: Uses the start date year to determine availability:
//   - Year 2101: Returns true (available) - used for testing successful booking flows
//   - All other years: Returns false (unavailable) - used for testing "no availability" scenarios
//
// This tri-modal approach enables testing of:
//   - Database error recovery and user messaging
//   - "Room unavailable" workflows and alternative suggestions
//   - Successful availability confirmation and booking progression
//   - Real-time availability checking via JSON API endpoints
//
// Parameters:
//   - start, end: Date range for availability checking (end date not currently used in logic)
//   - roomID: Specific room identifier for availability checking
//
// Returns:
//   - bool: true if room is available, false if unavailable or error occurred
//   - error: Simulated database error for specific test conditions, nil for normal operation
func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	// Check for dynamically configured error condition via toggle system
	if ForceSearchAvailabilityErrOn != 0 && roomID == ForceSearchAvailabilityErrOn {
		return false, errors.New("db error")
	}

	// Legacy error condition maintained for existing test compatibility
	if roomID == 2 {
		return false, errors.New("db error")
	}

	// Date-based availability logic for predictable test scenarios
	if start.Year() == 2101 {
		return true, nil // Available - triggers successful booking workflows
	}

	return false, nil // Unavailable - triggers "no availability" workflows
}

// SearchAvailabilityForAllRooms simulates comprehensive availability search across all rooms.
// This method supports testing of the main availability search functionality where users
// input desired dates and receive a list of available rooms for selection.
//
// The method implements controlled responses based on global toggles and date patterns:
//
//  1. **Error Simulation**: When ForceAllRoomsErr is true, returns a database error
//     to test error handling in availability search workflows.
//
//  2. **Rooms Available**: When start date year is 2101, returns a single room
//     (Golden Haybeam Loft) to simulate successful availability search results.
//     This enables testing of room selection interfaces and booking progression.
//
//  3. **No Availability**: For all other date combinations, returns an empty slice
//     to simulate scenarios where no rooms are available for the requested dates.
//     This enables testing of "no availability" messaging and alternative suggestions.
//
// The predictable room response (Golden Haybeam Loft, ID: 1) provides consistency
// for tests that need to verify room selection and booking workflows without
// dependency on specific room data or database state.
//
// Parameters:
//   - start, end: Date range for availability search across all rooms
//
// Returns:
//   - []models.Room: Available rooms list (empty, single room, or error based on test conditions)
//   - error: Simulated database error when ForceAllRoomsErr is true, nil otherwise
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	// Check for forced error condition via toggle system
	if ForceAllRoomsErr {
		return nil, errors.New("all rooms error")
	}

	// Return available room for specific test scenario (year 2101)
	if start.Year() == 2101 {
		return []models.Room{{ID: 1, RoomName: "Golden Haybeam Loft"}}, nil
	}

	// Return empty availability for all other scenarios
	return []models.Room{}, nil
}

// GetRoomByID retrieves room information with controlled error scenarios for testing.
// This method simulates database room lookup operations while providing predictable
// responses for both successful retrieval and "room not found" error conditions.
//
// Test behavior patterns:
//   - ID > 3: Returns "room not found" error to test invalid room ID handling
//   - ID 1-3: Returns mock room data with the provided ID and generic name
//
// The "room not found" condition (ID > 3) enables testing of error handling
// throughout the application stack, including:
//   - User-friendly error messages for invalid room requests
//   - Graceful degradation when room data is unavailable
//   - Validation of room IDs before proceeding with booking workflows
//   - Error recovery and alternative suggestion systems
//
// Parameters:
//   - id: Room identifier to retrieve
//
// Returns:
//   - models.Room: Mock room data with provided ID and generic name
//   - error: "room not found" error when id > 3, nil otherwise
func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	// Simulate "room not found" for IDs beyond test data range
	if id > 3 {
		return models.Room{}, errors.New("room not found")
	}

	// Return mock room data with provided ID
	return models.Room{ID: id, RoomName: "Room"}, nil
}

// GetUserByID is a placeholder method that returns an empty User model.
// This method is implemented to satisfy the DatabaseRepo interface requirements
// but provides minimal functionality in the test environment since user management
// testing is not the primary focus of the current test suite.
//
// In a more comprehensive test environment, this method would provide controlled
// user data responses and error scenarios similar to other test repository methods.
//
// Parameters:
//   - id: User identifier (not used in current implementation)
//
// Returns:
//   - models.User: Empty user model
//   - error: Always nil in current implementation
func (m *testDBRepo) GetUserByID(id int) (models.User, error) {
	return models.User{}, nil
}

// UpdateUser is a placeholder method that always succeeds.
// This method is implemented to satisfy the DatabaseRepo interface requirements
// but provides minimal functionality for user update operations in the test environment.
//
// Parameters:
//   - u: User model with updated information (not processed in current implementation)
//
// Returns:
//   - error: Always nil in current implementation
func (m *testDBRepo) UpdateUser(u models.User) error {
	return nil
}

// Authenticate simulates user authentication with controlled success and failure scenarios.
// This method enables testing of login workflows, authentication error handling,
// and session management without requiring actual user accounts or password hashing.
//
// Test behavior patterns:
//   - Email "badlogin@example.com": Returns authentication error to test login failure handling
//   - All other emails: Returns successful authentication with user ID 1
//
// The controlled failure scenario enables testing of:
//   - Invalid credential handling and user messaging
//   - Login attempt rate limiting and security measures
//   - Authentication error logging and monitoring
//   - Failed login redirect and retry workflows
//
// Parameters:
//   - email: Email address used for authentication attempt
//   - _: Password parameter (ignored in test implementation)
//
// Returns:
//   - int: User ID (1) for successful authentication, 0 for failures
//   - string: Empty password hash (not used in test scenarios)
//   - error: Authentication error for "badlogin@example.com", nil for success
func (m *testDBRepo) Authenticate(email, _ string) (int, string, error) {
	// Simulate authentication failure for specific test email
	if email == "badlogin@example.com" {
		return 0, "", errors.New("invalid credentials")
	}

	// Return successful authentication for all other emails
	return 1, "", nil
}

// AllReservations retrieves all reservations with controlled error scenarios for administrative testing.
// This method simulates the comprehensive reservation listing functionality used in administrative
// interfaces while providing error conditions for testing database connectivity and error handling.
//
// When operating normally, returns a minimal reservation list with basic guest information
// sufficient for testing reservation display, sorting, and administrative workflow functionality.
//
// Error scenarios (when ForceAllReservationsErr is true) enable testing of:
//   - Database connectivity failure recovery
//   - Administrative interface error handling and user messaging
//   - Graceful degradation when reservation data is unavailable
//   - Error logging and monitoring in administrative systems
//
// Returns:
//   - []models.Reservation: Single mock reservation for testing or nil if error forced
//   - error: Simulated database error when ForceAllReservationsErr is true, nil otherwise
func (m *testDBRepo) AllReservations() ([]models.Reservation, error) {
	// Check for forced error condition via toggle system
	if ForceAllReservationsErr {
		return nil, errors.New("all reservations error")
	}

	// Return minimal reservation data for successful testing scenarios
	return []models.Reservation{{ID: 1, FirstName: "A", LastName: "B"}}, nil
}

// AllNewReservations retrieves unprocessed reservations with controlled error scenarios.
// This method simulates the new reservation queue functionality used by administrative staff
// to review, validate, and process incoming guest bookings.
//
// When operating normally, returns a minimal unprocessed reservation with different guest
// information than AllReservations() to enable testing of reservation processing workflows
// and staff interface functionality.
//
// Error scenarios (when ForceAllNewReservationsErr is true) enable testing of:
//   - Database connectivity failure during staff processing workflows
//   - New reservation queue error handling and recovery
//   - Staff notification systems when reservation data is unavailable
//   - Administrative workflow continuity during database issues
//
// Returns:
//   - []models.Reservation: Single mock unprocessed reservation or nil if error forced
//   - error: Simulated database error when ForceAllNewReservationsErr is true, nil otherwise
func (m *testDBRepo) AllNewReservations() ([]models.Reservation, error) {
	// Check for forced error condition via toggle system
	if ForceAllNewReservationsErr {
		return nil, errors.New("all new reservations error")
	}

	// Return minimal unprocessed reservation data for testing
	return []models.Reservation{{ID: 2, FirstName: "C", LastName: "D"}}, nil
}

// GetReservationByID retrieves specific reservation details with controlled error scenarios.
// This method simulates individual reservation lookup operations used throughout administrative
// interfaces for detailed reservation display, editing, and processing workflows.
//
// When operating normally, returns a minimal reservation model with the provided ID,
// sufficient for testing reservation detail interfaces and modification workflows.
//
// Error scenarios (when ForceGetReservationErr is true) enable testing of:
//   - Database connectivity failure during reservation detail access
//   - Administrative interface error handling for missing or inaccessible reservations
//   - Graceful degradation when specific reservation data cannot be retrieved
//   - Error recovery and alternative workflows in administrative systems
//
// Parameters:
//   - id: Reservation identifier for retrieval
//
// Returns:
//   - models.Reservation: Mock reservation with provided ID or empty if error forced
//   - error: Simulated database error when ForceGetReservationErr is true, nil otherwise
func (m *testDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	// Check for forced error condition via toggle system
	if ForceGetReservationErr {
		return models.Reservation{}, errors.New("get reservation error")
	}

	// Return minimal reservation data with provided ID
	return models.Reservation{ID: id}, nil
}

// UpdateReservation modifies reservation information with controlled error scenarios.
// This method simulates reservation update operations used in administrative interfaces
// for guest information correction, contact detail updates, and reservation modifications.
//
// Error scenarios (when ForceUpdateReservationErr is true) enable testing of:
//   - Database connectivity failure during reservation modification
//   - Administrative interface error handling for failed update operations
//   - Transaction rollback scenarios when updates cannot be completed
//   - User messaging and retry workflows for failed reservation updates
//
// Parameters:
//   - u: Reservation model with updated information (not processed in test implementation)
//
// Returns:
//   - error: Simulated database error when ForceUpdateReservationErr is true, nil otherwise
func (m *testDBRepo) UpdateReservation(u models.Reservation) error {
	// Check for forced error condition via toggle system
	if ForceUpdateReservationErr {
		return errors.New("update reservation error")
	}

	return nil
}

// DeleteReservation is a placeholder method that always succeeds.
// This method simulates reservation deletion operations but provides minimal
// functionality in the current test environment.
//
// Parameters:
//   - id: Reservation identifier for deletion (not processed in current implementation)
//
// Returns:
//   - error: Always nil in current implementation
func (m *testDBRepo) DeleteReservation(id int) error {
	return nil
}

// UpdateProcessedForReservation modifies reservation processing status with controlled error scenarios.
// This method simulates the reservation processing workflow where administrative staff mark
// reservations as reviewed, validated, and ready for guest communication and service delivery.
//
// Error scenarios (when ForceProcessedUpdateErr is true) enable testing of:
//   - Database connectivity failure during reservation processing workflows
//   - Administrative interface error handling for failed status updates
//   - Staff notification systems when processing status cannot be updated
//   - Workflow recovery when reservation processing operations fail
//
// Parameters:
//   - id: Reservation identifier for processing status update
//   - processed: New processing status (typically 0 for unprocessed, 1 for processed)
//
// Returns:
//   - error: Simulated database error when ForceProcessedUpdateErr is true, nil otherwise
func (m *testDBRepo) UpdateProcessedForReservation(id, processed int) error {
	// Check for forced error condition via toggle system
	if ForceProcessedUpdateErr {
		return errors.New("processed update error")
	}

	return nil
}

// AllRooms retrieves comprehensive room information with controlled error scenarios.
// This method simulates room listing operations used throughout the application for
// availability checking, administrative calendar displays, and room selection interfaces.
//
// When operating normally, returns a single room (Golden Haybeam Loft, ID: 1) which
// provides sufficient data for testing room-based functionality while maintaining
// consistency with other test methods that reference room ID 1.
//
// Error scenarios (when ForceAllRoomsErr is true) enable testing of:
//   - Database connectivity failure during room data retrieval
//   - Calendar interface error handling when room information is unavailable
//   - Graceful degradation of availability checking when room data cannot be accessed
//   - Administrative system error recovery for room management operations
//
// The single room response pattern enables testing of calendar and availability systems
// by providing predictable room data for session storage and administrative interface
// functionality without requiring complex test data setup.
//
// Returns:
//   - []models.Room: Single room (Golden Haybeam Loft) or nil if error forced
//   - error: Simulated database error when ForceAllRoomsErr is true, nil otherwise
func (m *testDBRepo) AllRooms() ([]models.Room, error) {
	// Check for forced error condition via toggle system
	if ForceAllRoomsErr {
		return nil, errors.New("all rooms error")
	}

	// Return consistent single room data for testing
	return []models.Room{{ID: 1, RoomName: "Golden Haybeam Loft"}}, nil
}

// GetRestrictionsForRoomByDate retrieves room restrictions with comprehensive test scenario support.
// This method simulates the complex room restriction query operations used by calendar interfaces
// and availability checking systems to determine room booking conflicts and administrative blocks.
//
// The method provides multiple types of test data to support comprehensive testing:
//
//  1. **Default Block Restriction**: Always includes one owner block (ReservationID = 0)
//     positioned 4 days after the start date. This simulates administrative room blocks
//     used for maintenance, personal use, or other non-guest restrictions.
//
//  2. **Optional Reservation Restriction**: When ForceHasReservationRestriction is true,
//     adds a reservation restriction (ReservationID = 777) spanning days 1-3 of the query period.
//     This simulates actual guest reservations and enables testing of reservation vs. block
//     distinction in calendar displays and availability calculations.
//
//  3. **Error Scenarios**: When ForceRestrictionsErr is true, returns database errors
//     to test error handling in calendar and availability systems.
//
// The dual restriction approach (blocks + reservations) enables comprehensive testing of:
//   - Calendar color coding and visual distinction between restriction types
//   - Availability calculation logic that considers both guest bookings and administrative blocks
//   - Administrative interfaces that allow editing blocks but not reservation restrictions
//   - Session storage of restriction maps for calendar form processing workflows
//
// Parameters:
//   - roomID: Room identifier for restriction query
//   - start, end: Date range for finding overlapping restrictions
//
// Returns:
//   - []models.RoomRestriction: List of restrictions (block + optional reservation) or nil if error
//   - error: Simulated database error when ForceRestrictionsErr is true, nil otherwise
func (m *testDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	// Check for forced error condition via toggle system
	if ForceRestrictionsErr {
		return nil, errors.New("restrictions error")
	}

	// Always include an owner block restriction to provide data for delete/keep testing loops
	// Positioned 4 days after start date to simulate realistic administrative blocking patterns
	res := []models.RoomRestriction{
		{
			ID:            11,
			StartDate:     start.AddDate(0, 0, 4),
			EndDate:       start.AddDate(0, 0, 4),
			RoomID:        roomID,
			ReservationID: 0, // Owner block (no associated reservation)
		},
	}

	// Optionally include a reservation restriction when toggle is enabled
	// This enables testing of reservation vs. block distinction in calendar systems
	if ForceHasReservationRestriction {
		res = append(res, models.RoomRestriction{
			ID:            42,
			StartDate:     start.AddDate(0, 0, 1),
			EndDate:       start.AddDate(0, 0, 3),
			RoomID:        roomID,
			ReservationID: 777, // Guest reservation ID
			RestrictionID: 1,   // Reservation type restriction
		})
	}

	return res, nil
}

// InsertBlockForRoom creates room blocks with controlled error scenarios for calendar testing.
// This method simulates the administrative block creation functionality used in calendar
// interfaces where staff can click dates to create owner blocks for maintenance, personal use,
// or other non-guest restrictions.
//
// Error scenarios (when ForceInsertBlockErr is true) enable testing of:
//   - Database connectivity failure during administrative calendar operations
//   - Calendar interface error handling for failed block creation attempts
//   - Error logging and user notification when blocks cannot be created
//   - Administrative workflow recovery when block operations fail
//
// Parameters:
//   - id: Room identifier for block creation
//   - startDate: Date to create block for (not processed in test implementation)
//
// Returns:
//   - error: Simulated database error when ForceInsertBlockErr is true, nil otherwise
func (m *testDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	// Check for forced error condition via toggle system
	if ForceInsertBlockErr {
		return errors.New("insert block error")
	}

	return nil
}

// DeleteBlockByID removes room blocks with controlled error scenarios for calendar testing.
// This method simulates the administrative block removal functionality used in calendar
// interfaces where staff can uncheck blocked dates to remove owner blocks and restore
// room availability for guest bookings.
//
// Error scenarios (when ForceDeleteBlockErr is true) enable testing of:
//   - Database connectivity failure during administrative calendar operations
//   - Calendar interface error handling for failed block removal attempts
//   - Error logging and user notification when blocks cannot be removed
//   - Administrative workflow recovery when block deletion operations fail
//
// Parameters:
//   - id: Block restriction identifier for removal (not processed in test implementation)
//
// Returns:
//   - error: Simulated database error when ForceDeleteBlockErr is true, nil otherwise
func (m *testDBRepo) DeleteBlockByID(id int) error {
	// Check for forced error condition via toggle system
	if ForceDeleteBlockErr {
		return errors.New("delete block error")
	}

	return nil
}
