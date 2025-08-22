package models

// Reservation represents the visitor details collected by the reservation form.
type Reservation struct {
	FirstName string // visitor's given name for messages/confirmations
	LastName  string // visitor's family name
	Email     string // contact address for follow-ups
	Phone     string // optional direct contact number
	// TODO(data): when you add dates/room selections, extend this struct accordingly.
}
