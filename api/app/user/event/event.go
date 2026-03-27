package event

// UserCreated is emitted when a new user account is registered.
type UserCreated struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}

func (UserCreated) EventName() string    { return "user.created" }
func (e UserCreated) EntityType() string { return "user" }
func (e UserCreated) EntityID() string   { return e.UserID }

// UserUpdated is emitted when a user's profile fields are changed.
type UserUpdated struct {
	UserID  string         `json:"user_id"`
	Changes map[string]any `json:"changes"`
}

func (UserUpdated) EventName() string    { return "user.updated" }
func (e UserUpdated) EntityType() string { return "user" }
func (e UserUpdated) EntityID() string   { return e.UserID }

// UserDeleted is emitted when a user is soft-deleted.
type UserDeleted struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func (UserDeleted) EventName() string    { return "user.deleted" }
func (e UserDeleted) EntityType() string { return "user" }
func (e UserDeleted) EntityID() string   { return e.UserID }

// UserRestored is emitted when a soft-deleted user is restored.
type UserRestored struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func (UserRestored) EventName() string    { return "user.restored" }
func (e UserRestored) EntityType() string { return "user" }
func (e UserRestored) EntityID() string   { return e.UserID }

// UserPasswordChanged is emitted when a user changes their own password.
type UserPasswordChanged struct {
	UserID string `json:"user_id"`
}

func (UserPasswordChanged) EventName() string    { return "user.password_changed" }
func (e UserPasswordChanged) EntityType() string { return "user" }
func (e UserPasswordChanged) EntityID() string   { return e.UserID }

// UserPasswordSet is emitted when an admin sets a user's password.
type UserPasswordSet struct {
	UserID string `json:"user_id"`
	SetBy  string `json:"set_by"`
}

func (UserPasswordSet) EventName() string    { return "user.password_set" }
func (e UserPasswordSet) EntityType() string { return "user" }
func (e UserPasswordSet) EntityID() string   { return e.UserID }
