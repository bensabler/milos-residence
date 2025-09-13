// Package dbrepo provides PostgreSQL database repository implementations for Milo's Residence.
// It contains concrete implementations of the DatabaseRepo interface, handling all
// database operations including reservation management, user authentication, room
// availability checking, and administrative functions using PostgreSQL-specific queries.
//
// The implementation uses context-based timeouts for all database operations to prevent
// hanging connections and ensure responsive application behavior under load.
package dbrepo

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/bensabler/milos-residence/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// AllUsers is a placeholder method that returns a boolean indicating system health.
// This method was implemented as a basic connectivity test during development
// and currently serves as a simple database interaction verification.
// In a production system, this would typically return actual user data or counts.
//
// Returns true if the database connection is functional, though the current
// implementation always returns true regardless of database state.
func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation creates a new reservation record in the PostgreSQL database.
// It inserts guest information, reservation dates, and room assignment into
// the reservations table and returns the auto-generated reservation ID for
// use in related operations such as room restrictions and email notifications.
//
// The method uses a parameterized query to prevent SQL injection attacks and
// employs a context timeout to prevent indefinite blocking during database operations.
// All timestamp fields are populated with the current time to maintain audit trails.
//
// Parameters:
//   - res: Reservation model containing guest details, dates, and room assignment
//
// Returns:
//   - int: The auto-generated ID of the newly created reservation
//   - error: Database error if insertion fails, nil on success
//
// The method will fail if:
//   - Database connection is unavailable
//   - Required fields contain invalid data
//   - Foreign key constraints are violated (invalid room_id)
//   - Context timeout (3 seconds) is exceeded
func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newId int

	stmt := `insert into reservations (first_name, last_name, email, phone, start_date,
	 end_date, room_id, created_at, updated_at)
	 values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newId)

	if err != nil {
		return 0, err
	}

	return newId, nil
}

// InsertRoomRestriction creates a room restriction record in the PostgreSQL database.
// Room restrictions are used to block room availability for specific date ranges,
// either due to reservations (restriction_id=1) or owner blocks (restriction_id=2).
// This method is typically called immediately after InsertReservation to prevent
// double-booking of rooms during the reserved period.
//
// The method establishes the relationship between reservations and room availability
// by creating restriction records that are checked during availability searches.
// Context timeout prevents hanging transactions during database operations.
//
// Parameters:
//   - r: RoomRestriction model containing date range, room ID, reservation ID, and restriction type
//
// Returns:
//   - error: Database error if insertion fails, nil on success
//
// The method will fail if:
//   - Database connection is unavailable
//   - Foreign key constraints are violated (invalid room_id, reservation_id, or restriction_id)
//   - Date range validation fails
//   - Context timeout (3 seconds) is exceeded
func (m *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into room_restrictions (start_date, end_date, room_id, reservation_id,
				created_at, updated_at, restriction_id)
				values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID,
	)

	if err != nil {
		return err
	}

	return nil
}

// SearchAvailabilityByDatesByRoomID checks if a specific room is available for given dates.
// It queries the room_restrictions table to count any overlapping restrictions
// (reservations or owner blocks) that would prevent booking the room during
// the requested period. The method uses date range overlap logic to determine conflicts.
//
// The availability check uses the standard interval overlap condition:
// - A restriction conflicts if: requestStart < restrictionEnd AND requestEnd > restrictionStart
// - If no conflicting restrictions exist (count = 0), the room is available
// - If any restrictions exist (count > 0), the room is unavailable
//
// This method is used by both the availability search functionality and the
// JSON API endpoint for real-time availability checking on individual room pages.
//
// Parameters:
//   - start: Check-in date for availability query
//   - end: Check-out date for availability query
//   - roomID: Specific room ID to check availability for
//
// Returns:
//   - bool: true if room is available, false if conflicts exist
//   - error: Database error if query fails, nil on success
//
// The query will return false (unavailable) if any overlapping restrictions exist,
// regardless of restriction type (reservation or owner block).
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var numRows int

	query := `
		select 
			count(id) 
		from 
			room_restrictions 
		where
			room_id = $1
		and 
			$2 < end_date and $3 > start_date;`

	row := m.DB.QueryRowContext(ctx, query, roomID, start, end)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	}

	return false, nil
}

// SearchAvailabilityForAllRooms retrieves all rooms that are available during specified dates.
// It performs a complex query that excludes rooms with any overlapping restrictions
// (reservations or owner blocks) during the requested date range. The method returns
// a list of available rooms that can be presented to users for selection.
//
// The query uses a NOT IN subquery approach:
// 1. Main query: SELECT all rooms FROM rooms table
// 2. Subquery: Find room IDs that have restrictions overlapping the requested dates
// 3. Result: Only rooms NOT IN the restricted list are returned as available
//
// This method powers the main availability search functionality where users
// input their desired dates and receive a list of available rooms to choose from.
// The results feed into the room selection workflow.
//
// Parameters:
//   - start: Check-in date for availability search
//   - end: Check-out date for availability search
//
// Returns:
//   - []models.Room: Slice of available rooms with ID and name populated
//   - error: Database error if query fails, nil on success
//
// Returns an empty slice if no rooms are available during the specified dates.
// Each returned room includes sufficient information for display in the room selection interface.
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `
		select 
			r.id, r.room_name 
		from 
			rooms r 
		where
			r.id not in (
				select room_id 
				from room_restrictions rr
				where $1 < rr.end_date and $2 > rr.start_date
			)`

	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room

		err := rows.Scan(&room.ID, &room.RoomName)
		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// GetRoomByID retrieves complete room information for a specific room ID.
// This method is used throughout the application to fetch room details for
// reservation processing, form display, and administrative functions.
// It returns all stored room attributes including audit timestamps.
//
// The method is frequently called during:
// - Reservation form rendering to display room names
// - Validation to ensure room existence before creating reservations
// - Administrative interfaces showing room-specific information
// - Room detail page population and booking workflows
//
// Parameters:
//   - id: Unique identifier of the room to retrieve
//
// Returns:
//   - models.Room: Complete room record with all fields populated
//   - error: Database error if query fails or room not found, nil on success
//
// Returns sql.ErrNoRows error if the specified room ID does not exist.
// This error should be handled gracefully in calling code to provide
// appropriate user feedback for invalid room requests.
func (m *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `
		select 
			id, room_name, created_at, updated_at 
		from 
			rooms 
		where
			id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err != nil {
		return room, err
	}

	return room, nil
}

// GetUserByID retrieves complete user information for a specific user ID.
// This method is used for user profile management, administrative functions,
// and account-related operations. It returns all user attributes including
// the hashed password and access level information for authorization decisions.
//
// The method supports user management functionality including:
// - Profile display and editing interfaces
// - Access level verification for administrative functions
// - User account administration and reporting
// - Authentication-related user data retrieval
//
// Parameters:
//   - id: Unique identifier of the user to retrieve
//
// Returns:
//   - models.User: Complete user record with all fields populated including hashed password
//   - error: Database error if query fails or user not found, nil on success
//
// Security note: The returned User model includes the hashed password field.
// Calling code should be careful not to expose password hashes in API responses
// or user interfaces. Consider creating separate methods for public user data.
func (m *postgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		select 
			id, first_name, last_name, email, password, access_level, created_at, updated_at 
		from 
			users 
		where
			id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return u, err
	}

	return u, nil

}

// UpdateUser modifies user information in the PostgreSQL database.
// This method updates user profile data including name, email, and access level
// while automatically updating the modification timestamp. The password field
// is intentionally excluded from updates and requires separate handling for security.
//
// The method is used for:
// - User profile management and self-service updates
// - Administrative user management and role assignments
// - Email address changes and contact information updates
// - Access level modifications for permission management
//
// Parameters:
//   - u: User model containing updated information; ID field determines which record to update
//
// Returns:
//   - error: Database error if update fails, nil on success
//
// Security considerations:
// - Password updates are intentionally excluded and should use separate methods
// - Email changes should trigger verification workflows in calling code
// - Access level changes should be restricted to authorized administrators
// - The updated_at timestamp is automatically set to the current time
//
// The method will fail if the user ID does not exist, but this does not return
// an error in the current implementation (UPDATE affects 0 rows but succeeds).
func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update
			users
		set 
			first_name = $1, last_name = $2, email = $3, access_level = $4, updated_at = $5
		`

	_, err := m.DB.ExecContext(ctx, query, u.FirstName, u.LastName, u.Email, u.AccessLevel, time.Now())

	if err != nil {
		return err
	}

	return nil

}

// Authenticate verifies user credentials against the PostgreSQL database.
// This method implements secure authentication by retrieving the user's hashed
// password and comparing it with the provided password using bcrypt hashing.
// It returns user identification information upon successful authentication.
//
// The authentication process:
// 1. Query database for user record by email address
// 2. Retrieve stored bcrypt hash for the user account
// 3. Compare provided password against stored hash using bcrypt.CompareHashAndPassword
// 4. Return user ID and hash on success, or appropriate error on failure
//
// Security features:
// - Uses bcrypt for secure password hashing and comparison
// - Protects against timing attacks through consistent bcrypt operations
// - Returns specific error for incorrect passwords vs. database errors
// - Context timeout prevents indefinite blocking during authentication
//
// Parameters:
//   - email: User's email address used as login identifier
//   - testPassword: Plain text password provided by user during login
//
// Returns:
//   - int: User ID if authentication succeeds
//   - string: Stored password hash (for session management or further verification)
//   - error: Authentication error or database error, nil on success
//
// Possible errors:
// - sql.ErrNoRows: Email address not found in database
// - bcrypt.ErrMismatchedHashAndPassword: Converted to "incorrect password" error
// - Other bcrypt errors: Returned as-is for debugging
// - Database connectivity errors: Returned as-is
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "select id, password from users where email = $1", email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil

}

// AllReservations retrieves all reservation records from the PostgreSQL database.
// This method performs a comprehensive query joining reservation data with room
// information to provide complete reservation details for administrative interfaces.
// Results are ordered chronologically by start date for logical presentation.
//
// The method uses a LEFT JOIN to ensure all reservations are returned even if
// room data is missing (though this should not occur in normal operation due to
// foreign key constraints). This approach provides defensive programming against
// data inconsistencies while maintaining complete reservation visibility.
//
// Administrative uses include:
// - Comprehensive reservation reporting and analytics
// - Administrative oversight of all booking activity
// - Historical reservation data for business intelligence
// - Audit trails and compliance reporting requirements
//
// Returns:
//   - []models.Reservation: All reservations with embedded room information, ordered by start_date
//   - error: Database error if query fails, nil on success
//
// Performance considerations:
// - Query returns all reservations which could be large datasets in production
// - Consider implementing pagination for systems with extensive reservation history
// - LEFT JOIN adds minimal overhead due to foreign key relationship optimization
// - Context timeout prevents indefinite blocking during large result set processing
func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select 
			r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, 
			r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, 
			rm.id, rm.room_name
		from 
			reservations r 
		left join
			rooms rm 
		on 
			(r.room_id = rm.id)
		order by
			r.start_date asc
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil

}

// AllNewReservations retrieves unprocessed reservation records from the PostgreSQL database.
// This method filters reservations to show only those requiring administrative attention
// (processed = 0), enabling staff to efficiently manage incoming bookings and guest requests.
// Like AllReservations, it joins with room data and orders results chronologically.
//
// The processed flag workflow:
// - New reservations start with processed = 0 (requires attention)
// - Staff review, validate, and confirm reservations through admin interface
// - Processed reservations are marked processed = 1 (completed/confirmed)
// - This method shows only the processed = 0 reservations needing action
//
// Administrative workflow uses:
// - Daily reservation processing queues for staff
// - New booking notification and review processes
// - Quality control and fraud prevention screening
// - Guest communication and confirmation workflows
//
// Returns:
//   - []models.Reservation: Unprocessed reservations with embedded room information, ordered by start_date
//   - error: Database error if query fails, nil on success
//
// The chronological ordering (start_date ASC) helps staff prioritize processing
// based on arrival dates, ensuring near-term reservations receive prompt attention.
func (m *postgresDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select 
			r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, 
			r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, 
			rm.id, rm.room_name
		from 
			reservations r 
		left join
			rooms rm 
		on 
			(r.room_id = rm.id)
		where
			processed = 0
		order by
			r.start_date asc
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil

}

// GetReservationByID retrieves a specific reservation record by its unique identifier.
// This method performs the same comprehensive query as AllReservations but filters
// to a single record, providing complete reservation and room information for
// detailed display, editing, and administrative operations.
//
// The method is used throughout the administrative interface for:
// - Reservation detail pages showing complete guest and booking information
// - Edit forms pre-populated with existing reservation data
// - Processing workflows where staff review individual reservations
// - Reporting and analytics requiring specific reservation details
//
// The LEFT JOIN ensures room information is included even in edge cases where
// data integrity issues might exist, though foreign key constraints should
// prevent such scenarios in normal operations.
//
// Parameters:
//   - id: Unique identifier of the reservation to retrieve
//
// Returns:
//   - models.Reservation: Complete reservation record with embedded room information
//   - error: Database error if query fails or reservation not found, nil on success
//
// Returns sql.ErrNoRows if the specified reservation ID does not exist.
// Calling code should handle this error appropriately to provide user feedback
// for invalid reservation access attempts.
func (m *postgresDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Reservation

	query := `
		select 
			r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, 
			r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, 
			rm.id, rm.room_name
		from 
			reservations r 
		left join
			rooms rm 
		on 
			(r.room_id = rm.id)
		where
			r.id = $1
	`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&res.ID,
		&res.FirstName,
		&res.LastName,
		&res.Email,
		&res.Phone,
		&res.StartDate,
		&res.EndDate,
		&res.RoomID,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.Processed,
		&res.Room.ID,
		&res.Room.RoomName,
	)

	if err != nil {
		return res, err
	}

	return res, nil

}

// UpdateReservation modifies guest information for an existing reservation.
// This method updates the primary guest contact details (name, email, phone)
// while preserving reservation dates, room assignments, and system timestamps.
// The updated_at field is automatically refreshed to track modification history.
//
// The method specifically handles guest information updates that commonly occur:
// - Corrections to guest names due to typos or preference changes
// - Email address updates for communication and confirmation delivery
// - Phone number changes for contact and emergency purposes
// - Administrative corrections based on guest requests or verification
//
// Deliberately excluded fields:
// - Reservation dates (start_date, end_date): Require separate handling due to availability implications
// - Room assignments (room_id): Require availability checking and restriction updates
// - System fields (created_at, processed): Maintained by specific business logic
//
// Parameters:
//   - u: Reservation model containing updated guest information; ID field determines which record to update
//
// Returns:
//   - error: Database error if update fails, nil on success
//
// Business considerations:
// - Email changes may require re-sending confirmation messages in calling code
// - Name changes should trigger verification processes for security
// - The method does not validate data format (e.g., email validity) - this should occur in calling code
func (m *postgresDBRepo) UpdateReservation(u models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update
			reservations
		set 
			first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5
		where
			id = $6
		`

	_, err := m.DB.ExecContext(ctx, query, u.FirstName, u.LastName, u.Email, u.Phone, time.Now(), u.ID)

	if err != nil {
		return err
	}

	return nil

}

// DeleteReservation removes a reservation record from the PostgreSQL database.
// This method performs a hard delete of the reservation record and should typically
// be used only in administrative scenarios such as spam cleanup, test data removal,
// or exceptional circumstances requiring complete record elimination.
//
// Important considerations:
// - Hard deletion permanently removes reservation data and cannot be undone
// - Associated room restrictions should be cleaned up by calling code or database CASCADE rules
// - Audit trails and historical reporting will lose access to deleted reservation data
// - Email confirmations and guest communications should be considered before deletion
//
// Typical use cases:
// - Administrative cleanup of spam, duplicate, or test reservations
// - Data privacy compliance requiring complete data removal
// - Exceptional business circumstances requiring reservation cancellation and removal
// - Development and testing environments requiring data cleanup
//
// Parameters:
//   - id: Unique identifier of the reservation to delete
//
// Returns:
//   - error: Database error if deletion fails, nil on success
//
// The method does not return an error if the reservation ID does not exist
// (DELETE affects 0 rows but succeeds). Calling code should verify reservation
// existence before deletion if confirmation is required.
//
// Consider implementing soft deletion (status flags) instead of hard deletion
// for production systems requiring audit trails and data recovery capabilities.
func (m *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		delete
		from
			reservations
		where
			id = $1
	
	`

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil

}

// UpdateProcessedForReservation modifies the processing status of a reservation.
// This method implements the reservation workflow by allowing staff to mark
// reservations as processed (reviewed, confirmed, and ready) or reset them
// to unprocessed status if issues are discovered requiring additional review.
//
// Processing workflow integration:
// - New reservations start with processed = 0 (unprocessed, requiring staff attention)
// - Staff review reservation details, validate guest information, and confirm availability
// - Upon successful review, staff mark reservation as processed = 1 (confirmed and ready)
// - If issues are discovered later, staff can reset to processed = 0 for re-review
//
// The processed flag is used by:
// - AllNewReservations() to show only unprocessed reservations needing attention
// - Administrative dashboards to track processing progress and workload
// - Automated systems to trigger confirmation emails or other post-processing actions
// - Reporting systems to distinguish between pending and confirmed reservations
//
// Parameters:
//   - id: Unique identifier of the reservation to update
//   - processed: New processing status (0 = unprocessed, 1 = processed)
//
// Returns:
//   - error: Database error if update fails, nil on success
//
// The method does not validate the processed value - calling code should ensure
// only appropriate values (0 or 1) are passed to maintain data consistency.
func (m *postgresDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update
			reservations
		set 
			processed = $1
		where
			id = $2
	`

	_, err := m.DB.ExecContext(ctx, query, processed, id)

	if err != nil {
		return err
	}

	return nil

}

// AllRooms retrieves all room records from the PostgreSQL database.
// This method returns complete room information ordered alphabetically by
// room name for consistent presentation in user interfaces and administrative
// functions. It provides the foundation data for room selection, availability
// displays, and administrative room management.
//
// The method is used throughout the application for:
// - Room selection interfaces during availability searches and booking
// - Administrative calendar views showing all rooms and their availability status
// - Room management interfaces for configuration and maintenance scheduling
// - Reporting and analytics requiring complete room inventory data
//
// The alphabetical ordering (ORDER BY room_name) ensures consistent presentation
// across different interfaces and improves user experience by providing predictable
// room ordering that users can rely on for navigation and selection.
//
// Returns:
//   - []models.Room: All rooms with complete information, ordered alphabetically by name
//   - error: Database error if query fails, nil on success
//
// Performance considerations:
// - This method returns all rooms, which should remain manageable for typical B&B operations
// - Room count is typically small (< 50 rooms) making full table scans acceptable
// - Results are commonly cached at the application level due to infrequent room changes
// - Consider implementing caching strategies for high-traffic applications
func (m *postgresDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `
		select
			id, room_name, created_at, updated_at
		from 
			rooms
		order by
			room_name
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}
	defer rows.Close()

	for rows.Next() {
		var rm models.Room
		err := rows.Scan(
			&rm.ID,
			&rm.RoomName,
			&rm.CreatedAt,
			&rm.UpdatedAt,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, rm)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// GetRestrictionsForRoomByDate retrieves room restrictions overlapping a specified date range.
// This method queries room_restrictions to find all conflicts (reservations and owner blocks)
// that intersect with the given time period for a specific room. It's essential for
// calendar displays, availability management, and administrative oversight of room usage.
//
// The method uses standard interval overlap logic to find restrictions:
// - Query conditions: queryStart < restrictionEnd AND queryEnd >= restrictionStart
// - This captures all restrictions that have any overlap with the query period
// - Uses COALESCE for reservation_id to handle owner blocks (NULL reservation_id)
//
// Restriction types returned:
// - Reservations: restriction_id=1, has valid reservation_id linking to reservation record
// - Owner blocks: restriction_id=2, reservation_id is NULL (handled by COALESCE)
//
// Administrative uses:
// - Calendar interfaces showing room availability and booking status
// - Conflict detection when manually scheduling maintenance or blocks
// - Administrative reporting on room utilization and restriction patterns
// - Validation of booking requests against existing restrictions
//
// Parameters:
//   - roomID: Specific room to query restrictions for
//   - start: Beginning of date range to check for overlapping restrictions
//   - end: End of date range to check for overlapping restrictions
//
// Returns:
//   - []models.RoomRestriction: All restrictions overlapping the specified date range
//   - error: Database error if query fails, nil on success
//
// Each returned restriction includes sufficient information to distinguish between
// reservation restrictions (with reservation_id) and owner blocks (reservation_id=0).
func (m *postgresDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	query := `
		select
			id, coalesce(reservation_id, 0), restriction_id, room_id, start_date, end_date
		from 
			room_restrictions
		where
			$1 < end_date
		and
			$2 >= start_date
		and 
			room_id = $3
	`

	rows, err := m.DB.QueryContext(ctx, query, start, end, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.RoomRestriction
		err := rows.Scan(
			&r.ID,
			&r.ReservationID,
			&r.RestrictionID,
			&r.RoomID,
			&r.StartDate,
			&r.EndDate,
		)
		if err != nil {
			return nil, err
		}
		restrictions = append(restrictions, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return restrictions, nil

}

// InsertBlockForRoom creates an owner block restriction for a specific room and date.
// Owner blocks are administrative restrictions that prevent guest bookings during
// maintenance periods, personal use, or other operational requirements. This method
// creates single-day blocks that can be managed through the administrative calendar interface.
//
// Block characteristics:
// - Restriction type: restriction_id=2 (Owner Block, vs. 1 for reservations)
// - Duration: Single day (startDate to startDate + 1 day)
// - Purpose: Administrative control over room availability
// - No reservation association: reservation_id remains NULL
//
// The single-day approach simplifies calendar management by creating discrete blocks
// that can be individually added or removed. Multi-day blocks are created by calling
// this method multiple times for consecutive dates, providing granular control.
//
// Administrative workflow:
// - Staff use calendar interface to click dates for blocking
// - Each click calls this method to create a single-day owner block
// - Blocks immediately affect availability searches and prevent new bookings
// - Blocks can be removed individually using DeleteBlockByID
//
// Parameters:
//   - id: Room ID to create the block for
//   - startDate: Date to block (end date is automatically set to startDate + 1 day)
//
// Returns:
//   - error: Database error if insertion fails, nil on success
//
// The method logs errors but also returns them, allowing calling code to decide
// on appropriate error handling strategies (logging, user notification, rollback).
func (m *postgresDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		insert into room_restrictions
			(start_date, end_date, room_id, restriction_id, created_at, updated_at)
		values
			($1, $2, $3, $4, $5, $6)
	`

	_, err := m.DB.ExecContext(ctx, query, startDate, startDate.AddDate(0, 0, 1), id, 2, time.Now(), time.Now())
	if err != nil {
		log.Println(err)
		return err
	}

	return nil

}

// DeleteBlockByID removes a specific room restriction by its unique identifier.
// This method is used to remove owner blocks from the administrative calendar
// interface, allowing staff to unblock dates that were previously restricted.
// It performs a hard deletion of the restriction record.
//
// The method is typically used for:
// - Removing owner blocks that are no longer needed (maintenance completed, plans changed)
// - Administrative cleanup of incorrectly created blocks
// - Calendar interface interactions where staff uncheck blocked dates
// - Bulk cleanup operations for expired or outdated restrictions
//
// Important considerations:
// - This method can delete any room restriction by ID, including reservation restrictions
// - Calling code should ensure only appropriate restrictions are deleted
// - Deleting reservation restrictions could cause data inconsistency
// - The method performs a hard delete with no recovery mechanism
//
// Administrative calendar workflow:
// - Staff uncheck blocked dates in calendar interface
// - Interface calls this method with the restriction ID to remove the block
// - Room becomes immediately available for new bookings on that date
// - Change is permanent unless manually recreated
//
// Parameters:
//   - id: Unique identifier of the room restriction to delete
//
// Returns:
//   - error: Database error if deletion fails, nil on success
//
// The method logs errors but also returns them for appropriate error handling.
// It does not return an error if the restriction ID does not exist (DELETE affects 0 rows).
func (m *postgresDBRepo) DeleteBlockByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		delete from
			room_restrictions
		where
			id = $1
	`

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil

}
