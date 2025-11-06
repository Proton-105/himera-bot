package domain

import "time"

// User represents an application user stored in the database.
type User struct {
	ID         int64
	TelegramID int64
	FirstName  string
	LastName   string
	Username   string
	Balance    int64
	CreatedAt  time.Time
}
