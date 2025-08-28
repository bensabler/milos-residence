package models

import (
	"time"
)

// User implements the Domain Model pattern for user account representation and management.
// This struct encapsulates all the data and behavior related to user accounts within the
// reservation system, demonstrating how Go applications model business entities with
// appropriate data types, validation constraints, and relationships. It serves as both
// a data structure for persistence operations and a business entity that can contain
// behavior and validation logic.
//
// The Domain Model pattern is central to building maintainable business applications
// because it provides a single, authoritative representation of business concepts that
// can be used consistently across all application layers. This approach prevents the
// fragmentation and inconsistency that occurs when business concepts are represented
// differently in different parts of the system.
//
// Design Pattern: Domain Model - represents user business entity with data and behavior
// Design Pattern: Data Transfer Object - structured data for database and API operations
// Design Pattern: Value Object - immutable data representation for specific business concepts
type User struct {
	// ID serves as the primary key for database persistence and unique identification
	// Using int follows Go conventions for database primary keys and provides
	// efficient lookup performance. In production systems, consider using int64
	// for larger scale applications or UUID for distributed systems
	ID int

	// FirstName stores the user's given name for personalization and identification
	// This field supports user-friendly greetings, personalized communications,
	// and helps distinguish between users with similar email addresses or surnames
	FirstName string

	// LastName stores the user's family name for formal identification and record keeping
	// Combined with FirstName, this provides human-readable user identification
	// that's essential for customer service, booking confirmations, and legal records
	LastName string

	// Email serves as the unique identifier for user authentication and communication
	// This field must be unique across all users to prevent account conflicts and
	// enables password reset flows, booking confirmations, and marketing communications.
	// Validation should ensure proper email format and uniqueness constraints
	Email string

	// Password stores the hashed (never plaintext) password for user authentication
	// This field should always contain a cryptographically secure hash of the user's
	// password using algorithms like bcrypt. Never store plaintext passwords as this
	// creates massive security vulnerabilities and regulatory compliance issues
	Password string

	// AccessLevel implements Role-Based Access Control (RBAC) for authorization
	// This integer field determines what actions the user can perform within the system:
	// 1 = Regular user (can make reservations, view own bookings)
	// 2 = Staff user (can view all reservations, manage availability)
	// 3 = Administrator (full system access, user management, configuration)
	// Consider using constants or enums to make access levels more readable
	AccessLevel int

	// CreatedAt provides an audit trail of when the user account was established
	// This timestamp is essential for compliance, analytics, and customer service.
	// It helps track user acquisition patterns, account age for trust scoring,
	// and provides context for customer service interactions
	CreatedAt time.Time

	// UpdatedAt tracks when the user record was last modified for change management
	// This field enables auditing of account changes, detecting suspicious modifications,
	// and coordinating updates across distributed systems. It should be automatically
	// updated whenever any other field in the user record changes
	UpdatedAt time.Time
}

// Room implements the Domain Model pattern for accommodation representation and management.
// This struct represents the physical accommodations available for reservation, encapsulating
// the essential data needed for room identification, presentation, and booking operations.
// It demonstrates how business entities can be modeled simply while supporting complex
// business workflows through relationships with other entities.
//
// The Room entity is central to the reservation system's business logic because it represents
// the primary resource being booked. All availability checking, pricing, and reservation
// workflows ultimately revolve around room entities and their associated constraints.
//
// Design Pattern: Domain Model - represents room business entity with identification and metadata
// Design Pattern: Resource Entity - represents bookable resource in reservation system
type Room struct {
	// ID provides unique identification for database operations and referential integrity
	// This primary key enables efficient room lookups, foreign key relationships,
	// and ensures data consistency across all reservation-related operations
	ID int

	// RoomName provides human-readable identification for marketing and user interface
	// This descriptive name appears in booking interfaces, confirmation emails, and
	// customer communications. Examples: "Golden Haybeam Loft", "Window Perch Theater"
	// The name should be descriptive enough for users to understand the accommodation type
	RoomName string

	// CreatedAt tracks when this room was added to the system for audit purposes
	// This timestamp helps with inventory management, tracking accommodation additions,
	// and provides context for analytics about room performance and availability patterns
	CreatedAt time.Time

	// UpdatedAt records the last modification time for change management and synchronization
	// This field enables tracking of room detail updates, pricing changes, and other
	// modifications that might affect booking workflows or customer communications
	UpdatedAt time.Time
}

// Restriction implements the Domain Model pattern for availability constraint management.
// This struct represents the different types of restrictions that can be applied to rooms,
// providing a taxonomy of why accommodations might be unavailable. It demonstrates how
// business rules and constraints can be modeled as first-class entities that can be
// configured, managed, and referenced consistently across the application.
//
// The Restriction entity enables flexible availability management by categorizing different
// reasons for room unavailability. This allows the system to handle business logic like
// "maintenance blocks versus guest reservations" differently while using the same underlying
// data structure and workflow patterns.
//
// Design Pattern: Domain Model - represents restriction type business entity
// Design Pattern: Reference Data - master data for categorizing room restrictions
// Design Pattern: Strategy - enables different handling of different restriction types
type Restriction struct {
	// ID provides unique identification for restriction type references
	// This primary key enables foreign key relationships from room restrictions
	// and ensures referential integrity across the availability management system
	ID int

	// RestrictionName provides human-readable description of the restriction type
	// This name appears in administrative interfaces and helps staff understand
	// why specific time periods are blocked. Examples: "Reservation", "Maintenance",
	// "Owner Block", "Seasonal Closure". Clear naming supports operational clarity
	RestrictionName string

	// CreatedAt tracks when this restriction type was defined in the system
	// This timestamp supports auditing of configuration changes and helps track
	// the evolution of business rules and availability management policies
	CreatedAt time.Time

	// UpdatedAt records when this restriction type was last modified
	// This enables tracking changes to business rules and ensures consistency
	// across distributed systems that might cache restriction type information
	UpdatedAt time.Time
}

// Reservation implements the Aggregate Root pattern for booking transaction management.
// This struct represents a complete reservation transaction with all associated customer
// information, booking details, and business relationships. It serves as the central entity
// for the reservation workflow and demonstrates how complex business processes can be
// modeled as cohesive data structures with clear ownership and consistency boundaries.
//
// As an Aggregate Root, Reservation controls access to its associated data and enforces
// business invariants across the entire booking transaction. This pattern ensures that
// reservations remain consistent even as the system evolves and additional complexity
// is added to the booking workflow.
//
// Design Pattern: Aggregate Root - primary entity controlling booking transaction consistency
// Design Pattern: Domain Model - encapsulates reservation business logic and data
// Design Pattern: Transaction Script - represents complete business transaction state
type Reservation struct {
	// ID serves as the unique identifier for this reservation transaction
	// This primary key enables reservation lookups, modification tracking, and
	// provides the stable reference needed for confirmation emails, customer service,
	// and integration with external systems like payment processors
	ID int

	// FirstName stores the primary guest's given name for personalization
	// This information appears in booking confirmations, check-in processes, and
	// customer communications. It supports personalized service delivery and
	// helps staff provide appropriate greetings and service during the stay
	FirstName string

	// LastName stores the primary guest's family name for formal identification
	// Combined with FirstName, this provides the guest identification needed for
	// check-in processes, legal records, and customer service interactions.
	// This information may be required for regulatory compliance in some jurisdictions
	LastName string

	// Email provides the primary communication channel for booking confirmations and updates
	// This field enables automated confirmation emails, booking modification notifications,
	// last-minute updates, and serves as a backup identification method if other
	// guest information is unclear or disputed during check-in processes
	Email string

	// Phone provides direct contact capability for urgent communications and coordination
	// This information enables staff to contact guests about check-in procedures,
	// emergency situations, booking modifications, or other time-sensitive matters
	// that require immediate attention and cannot wait for email responses
	Phone string

	// StartDate defines when the guest's stay begins, controlling check-in timing
	// This date must be in the future (when reservation is created) and determines
	// room availability calculations, staff preparation schedules, and automated
	// communication timing for pre-arrival information and instructions
	StartDate time.Time

	// EndDate defines when the guest's stay concludes, controlling check-out timing
	// This date must be after StartDate and determines room availability for subsequent
	// bookings, housekeeping schedules, and departure procedures. The date range
	// (StartDate to EndDate) represents the complete reservation period
	EndDate time.Time

	// RoomID establishes the relationship between this reservation and the booked accommodation
	// This foreign key reference enables efficient lookups of room details, ensures
	// referential integrity, and supports availability checking algorithms that prevent
	// double-booking conflicts by tracking which rooms are reserved for which dates
	RoomID int

	// CreatedAt provides an audit trail of when this reservation was initially made
	// This timestamp is crucial for analytics (booking lead times), customer service
	// (understanding booking context), and business intelligence (peak booking periods).
	// It also supports fraud detection by identifying unusual booking patterns
	CreatedAt time.Time

	// UpdatedAt tracks when this reservation was last modified for change management
	// This field enables auditing of reservation changes, coordinating updates across
	// integrated systems, and provides context for customer service when handling
	// booking modifications, cancellations, or dispute resolution
	UpdatedAt time.Time

	// Room provides access to detailed accommodation information through object composition
	// This embedded Room instance eliminates the need for separate database queries
	// to retrieve room details when displaying reservation information. It demonstrates
	// how object composition can improve performance and simplify data access patterns
	Room Room
}

// RoomRestriction implements the Association Entity pattern for availability constraint management.
// This struct represents the complex many-to-many relationship between rooms, time periods,
// reservations, and restriction types. It serves as the central mechanism for preventing
// double-booking conflicts and managing all forms of room unavailability within the system.
//
// The Association Entity pattern is necessary here because the relationship between rooms
// and restrictions involves additional data (dates, associated reservations) that cannot
// be represented in a simple foreign key relationship. This pattern enables sophisticated
// availability management while maintaining referential integrity and data consistency.
//
// Design Pattern: Association Entity - manages complex many-to-many relationships with additional data
// Design Pattern: Business Rule - encodes availability constraints and booking conflict prevention
// Design Pattern: Temporal Data - handles time-based business rules and constraints
type RoomRestriction struct {
	// ID provides unique identification for this specific restriction instance
	// This primary key enables modification and deletion of individual restrictions
	// while maintaining referential integrity across the availability management system.
	// Each restriction instance represents a specific time period when a room is unavailable
	ID int

	// StartDate defines when this restriction period begins, blocking room availability
	// This date works in conjunction with EndDate to define a continuous time period
	// during which the associated room cannot be booked for new reservations.
	// The system uses these dates for availability checking and conflict detection
	StartDate time.Time

	// EndDate defines when this restriction period concludes, returning room to availability
	// This date must be after StartDate to create a valid restriction period.
	// The system automatically removes expired restrictions and makes rooms available
	// for new bookings once the EndDate has passed
	EndDate time.Time

	// RoomID identifies which room this restriction applies to via foreign key relationship
	// This reference ensures that availability checking algorithms can efficiently find
	// all restrictions affecting a specific room during any given time period.
	// The foreign key constraint ensures data integrity and prevents orphaned restrictions
	RoomID int

	// ReservationID links this restriction to a specific reservation transaction (nullable)
	// When a guest makes a reservation, a corresponding RoomRestriction is created to
	// block the room during the reservation period. If this field is null, the restriction
	// represents a non-reservation block (maintenance, owner use, seasonal closure, etc.)
	ReservationID int

	// RestrictionID categorizes the type of restriction via foreign key to Restriction entity
	// This reference enables different business logic for different restriction types:
	// Type 1 (Reservation): Created automatically when guests book, removed on checkout
	// Type 2 (Maintenance): Created by staff for repairs, cleaning, or maintenance work
	// Type 3 (Owner Block): Created for owner personal use or business purposes
	RestrictionID int

	// UpdatedAt tracks when this restriction was last modified for audit and synchronization
	// This timestamp enables change tracking, conflict resolution in distributed systems,
	// and provides audit trail for availability management decisions and modifications
	UpdatedAt time.Time

	// Room provides access to room details through object composition for display purposes
	// This embedded Room instance eliminates additional database queries when displaying
	// restriction information in administrative interfaces or availability reports
	Room Room

	// Reservation provides access to associated booking details when applicable
	// This embedded Reservation instance is populated when ReservationID is not null,
	// enabling rich display of restriction information with guest details and booking context
	// When ReservationID is null (non-reservation restrictions), this field remains empty
	Reservation Reservation

	// Restriction provides access to restriction type details through object composition
	// This embedded Restriction instance provides the human-readable restriction type
	// information needed for administrative interfaces, reports, and business logic
	// that needs to handle different restriction types with different rules
	Restriction Restriction
}

// MailData holds an email message
type MailData struct {
	To      string
	From    string
	Subject string
	Content string
}
