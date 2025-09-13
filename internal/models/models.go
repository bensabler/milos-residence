// Package models defines the core domain entities used across the application.
// These types represent users, rooms, reservations, and related constraints.
// They are intentionally light-weight and free of persistence concerns so they
// can be reused in handlers, repositories, and templates without side effects.
package models

import "time"

// User represents an application user with authorization context.
// Password is expected to be stored as a secure hash (never plaintext).
type User struct {
	ID          int       // Primary key
	FirstName   string    // Given name
	LastName    string    // Family name
	Email       string    // Unique email address for login/notifications
	Password    string    // Hashed password (implementation detail outside this package)
	AccessLevel int       // Authorization level/role; higher implies more privileges
	CreatedAt   time.Time // Creation timestamp (UTC recommended)
	UpdatedAt   time.Time // Last update timestamp
}

// Room represents a reservable unit (e.g., a named suite).
type Room struct {
	ID        int       // Primary key
	RoomName  string    // Human-readable name (unique display label)
	CreatedAt time.Time // Creation timestamp
	UpdatedAt time.Time // Last update timestamp
}

// Restriction captures a policy that limits availability (e.g., blackout).
type Restriction struct {
	ID              int       // Primary key
	RestrictionName string    // Human-readable label (e.g., "Owner Block", "Maintenance")
	CreatedAt       time.Time // Creation timestamp
	UpdatedAt       time.Time // Last update timestamp
}

// Reservation represents a booking request/record for a room across a date range.
type Reservation struct {
	ID        int       // Primary key
	FirstName string    // Guest given name
	LastName  string    // Guest family name
	Email     string    // Guest email for correspondence
	Phone     string    // Guest phone number
	StartDate time.Time // Check-in (inclusive)
	EndDate   time.Time // Check-out (exclusive by convention unless specified)
	RoomID    int       // Foreign key to Room
	CreatedAt time.Time // Creation timestamp
	UpdatedAt time.Time // Last update timestamp
	Processed int       // Processing status flag (0/1 or enum mapping)
	Room      Room      // Eager-loaded room details (optional; zero value if not set)
}

// RoomRestriction associates a restriction with a specific room (and optionally
// a reservation) across a date range, enforcing availability constraints.
type RoomRestriction struct {
	ID            int         // Primary key
	StartDate     time.Time   // Range start (inclusive)
	EndDate       time.Time   // Range end (exclusive by convention unless specified)
	RoomID        int         // Foreign key to Room
	ReservationID int         // Optional link to Reservation (0 if not tied)
	RestrictionID int         // Foreign key to Restriction
	UpdatedAt     time.Time   // Last update timestamp
	Room          Room        // Eager-loaded Room (optional)
	Reservation   Reservation // Eager-loaded Reservation (optional)
	Restriction   Restriction // Eager-loaded Restriction (optional)
}

// MailData contains information needed to send an email message, optionally
// referencing a template name for rendering the body.
type MailData struct {
	To       string // Recipient email address
	From     string // Sender email address
	Subject  string // Message subject line
	Content  string // Raw content; may be ignored if Template is used
	Template string // Template identifier for render pipeline (optional)
}
