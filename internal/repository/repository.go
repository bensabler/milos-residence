package repository

import (
	"time"

	"github.com/bensabler/milos-residence/internal/models"
)

// DatabaseRepo implements the Repository pattern interface for data persistence operations.
// This interface defines the contract for all data access operations in the reservation system,
// enabling different database implementations (PostgreSQL, MySQL, in-memory, mock) while
// maintaining consistent business logic in handlers and services. It demonstrates how Go
// interfaces enable dependency inversion and create clean architectural boundaries.
//
// The Repository pattern provides several architectural benefits that are crucial for
// maintainable business applications:
//
//  1. **Dependency Inversion**: Business logic depends on abstractions (this interface) rather
//     than concrete implementations (specific database code), making the system more flexible
//     and easier to modify as requirements change.
//
//  2. **Testability**: Mock implementations of this interface enable fast, isolated unit tests
//     of business logic without requiring actual database connections or complex test setup.
//
//  3. **Database Agnosticism**: The same business logic works with PostgreSQL, MySQL, SQLite,
//     or any other database by simply providing different implementations of this interface.
//
//  4. **Performance Optimization**: Different implementations can use database-specific
//     optimizations (stored procedures, advanced indexing, caching strategies) without
//     affecting business logic or requiring code changes in handlers.
//
//  5. **Environment Flexibility**: Development can use lightweight databases (SQLite) while
//     production uses enterprise databases (PostgreSQL) with the same application code.
//
// This interface follows the Interface Segregation Principle by focusing specifically on
// data access operations for the reservation domain, avoiding the temptation to create
// overly broad interfaces that mix unrelated concerns.
//
// Design Pattern: Repository - abstracts data persistence operations behind clean interface
// Design Pattern: Interface Segregation - focused interface for specific domain operations
// Design Pattern: Dependency Inversion - business logic depends on abstraction, not implementation
type DatabaseRepo interface {
	// User Management Operations
	// These methods handle basic user account operations and demonstrate how the repository
	// pattern can start simple and grow more sophisticated as business requirements evolve.

	// AllUsers demonstrates basic database connectivity and query operations for system health checking.
	// This method serves primarily as a database connection test and could be expanded to support
	// user management features like admin dashboards or user analytics. The boolean return value
	// provides a simple success/failure indication that can be used for health monitoring.
	//
	// In production systems, this method might evolve to return actual user data for administrative
	// interfaces, support pagination for large user bases, or include filtering criteria for
	// specific user segments. The current signature provides a foundation that can be extended
	// without breaking existing implementations or requiring changes to calling code.
	//
	// Design Pattern: Query - read-only data access operation
	// Design Pattern: Health Check - basic connectivity and functionality validation
	// Returns: true if query succeeds (indicating database connectivity), false on failure
	AllUsers() bool

	// Reservation Management Operations
	// These methods implement the core business functionality of the reservation system,
	// demonstrating how the Command Query Responsibility Segregation (CQRS) pattern can be
	// applied within a repository interface to clearly separate read and write operations.

	// InsertReservation implements the Command pattern for creating new reservation transactions.
	// This method handles the complex process of persisting reservation data while maintaining
	// referential integrity and business rule compliance. It demonstrates how repository methods
	// can encapsulate transaction management, validation, and error handling behind a clean interface.
	//
	// The method returns the newly created reservation ID, which is essential for subsequent
	// operations like creating room restrictions, sending confirmation emails, or updating
	// reservation details. This pattern ensures that calling code can immediately reference
	// the new reservation without requiring additional queries or complex coordination logic.
	//
	// Implementation considerations include:
	//   - Transaction management to ensure data consistency
	//   - Validation of business rules (dates, guest information, room availability)
	//   - Error handling for constraint violations and system failures
	//   - Audit trail creation for compliance and customer service
	//
	// Design Pattern: Command - state-changing operation with clear transaction boundaries
	// Design Pattern: Factory - creates new persistent entities and returns references
	// Parameters:
	//   res: Complete reservation information including guest details and booking dates
	// Returns: Database-generated ID for the new reservation, or error if creation fails
	InsertReservation(res models.Reservation) (int, error)

	// InsertRoomRestriction implements the Command pattern for availability management.
	// This method creates room restriction records that prevent double-booking by blocking
	// specific rooms during specific time periods. It works in conjunction with InsertReservation
	// to ensure that successful bookings immediately create corresponding availability blocks.
	//
	// The method handles different types of restrictions (guest reservations, maintenance blocks,
	// owner usage) through the restriction type system, demonstrating how the same data structure
	// and operations can support multiple business scenarios with different rules and behaviors.
	//
	// Critical business rules enforced by implementations include:
	//   - Prevention of overlapping restrictions that would create booking conflicts
	//   - Validation of date ranges to ensure logical consistency (start before end)
	//   - Referential integrity with rooms, reservations, and restriction types
	//   - Proper handling of restriction lifecycle (creation, modification, expiration)
	//
	// Design Pattern: Command - state-changing operation for availability management
	// Design Pattern: Business Rule Enforcement - validates and enforces booking constraints
	// Parameters:
	//   r: Room restriction details including dates, room, and restriction type
	// Returns: error if restriction cannot be created due to conflicts or validation failures
	InsertRoomRestriction(r models.RoomRestriction) error

	// Availability Query Operations
	// These methods implement sophisticated availability checking algorithms that form the
	// foundation of the booking system's user interface and business logic. They demonstrate
	// how complex business queries can be abstracted behind simple, intuitive interfaces.

	// SearchAvailabilityByDatesByRoomID implements the Query pattern for specific room availability checking.
	// This method performs targeted availability analysis for a single room during a specific
	// date range, supporting both user-facing booking interfaces and internal system operations
	// that need to verify room availability before processing reservations or modifications.
	//
	// The boolean return value provides immediate, actionable information for booking workflows:
	// true indicates the room is completely available and can be booked immediately, while false
	// indicates conflicts exist that prevent booking. This simple interface hides complex
	// database queries that check for overlapping reservations, maintenance blocks, and other
	// restrictions that might prevent booking.
	//
	// Implementation typically involves sophisticated SQL queries that:
	//   - Check for date range overlaps using interval mathematics
	//   - Consider all restriction types that might block availability
	//   - Handle edge cases like same-day checkout/check-in scenarios
	//   - Optimize query performance through proper indexing and query structure
	//
	// Design Pattern: Query - read-only availability checking operation
	// Design Pattern: Specification - encapsulates complex availability business rules
	// Parameters:
	//   start: Beginning of requested availability period (check-in date)
	//   end: End of requested availability period (check-out date)
	//   roomID: Specific room identifier to check for availability
	// Returns: true if room is available for entire period, false if any conflicts exist
	SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error)

	// SearchAvailabilityForAllRooms implements the Query pattern for comprehensive availability search.
	// This method performs system-wide availability analysis to find all rooms that are completely
	// available during a specified date range. It supports the primary user booking workflow by
	// enabling customers to see all available options for their desired dates.
	//
	// The method returns a slice of available rooms, providing rich information that enables
	// sophisticated user interfaces with room details, pricing, and availability status. This
	// approach scales efficiently because it performs one comprehensive query rather than
	// individual queries for each room, reducing database load and improving response times.
	//
	// Key implementation considerations include:
	//   - Performance optimization for queries across potentially large room inventories
	//   - Consistent ordering of results to provide predictable user experiences
	//   - Integration with room details to support rich presentation interfaces
	//   - Handling of business rules like minimum stay requirements or seasonal availability
	//
	// The returned room slice can be empty (no availability) without indicating an error,
	// distinguishing between "no rooms available" (valid business outcome) and "system failure"
	// (error condition requiring different handling and user feedback).
	//
	// Design Pattern: Query - comprehensive availability search across room inventory
	// Design Pattern: Collection Return - returns structured data for complex user interfaces
	// Parameters:
	//   start: Beginning of requested availability period
	//   end: End of requested availability period
	// Returns: Slice of available rooms with complete room information, or error for system failures
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)

	// Room Information Operations
	// These methods provide access to room master data that supports both user-facing
	// interfaces and internal business operations requiring room details and characteristics.

	// GetRoomByID implements the Query pattern for single room data retrieval.
	// This method provides efficient access to complete room information based on the room's
	// unique identifier, supporting booking confirmations, administrative interfaces, and
	// business logic that needs room details for processing or display purposes.
	//
	// The method demonstrates the Active Record pattern by returning a complete Room entity
	// with all associated information ready for immediate use, eliminating the need for
	// calling code to perform multiple queries or complex data assembly operations.
	//
	// Common usage scenarios include:
	//   - Displaying room details during booking confirmation workflows
	//   - Populating administrative interfaces with room information for management
	//   - Supporting business logic that makes decisions based on room characteristics
	//   - Generating reports and analytics that require room metadata
	//
	// Error handling distinguishes between "room not found" (which might be a user error
	// or stale reference) and "system error" (database connectivity problems or other
	// technical failures), enabling appropriate error handling and user feedback in
	// different scenarios.
	//
	// Design Pattern: Query - single entity retrieval by primary key
	// Design Pattern: Active Record - returns complete entity ready for immediate use
	// Parameters:
	//   id: Unique room identifier (primary key) for retrieval
	// Returns: Complete room entity with all details, or error if room not found or system failure
	GetRoomByID(id int) (models.Room, error)
}
