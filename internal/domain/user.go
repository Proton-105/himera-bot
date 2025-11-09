package domain

import "time"

// User represents an application user stored in the database.
type User struct {
	ID           int64
	TelegramID   int64
	FirstName    string
	LastName     string
	Username     string
	Balance      int64
	LastActiveAt time.Time
	IsBlocked    bool
	CreatedAt    time.Time
}

// UserSettings contains persisted per-user preferences.
type UserSettings struct {
	NotificationsEnabled bool
	Language             string
	Timezone             string
}
