package event

// AuthLogin is emitted when a user successfully authenticates.
type AuthLogin struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func (AuthLogin) EventName() string    { return "auth.login" }
func (e AuthLogin) EntityType() string { return "user" }
func (e AuthLogin) EntityID() string   { return e.UserID }

// AuthLogout is emitted when a user logs out.
type AuthLogout struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
}

func (AuthLogout) EventName() string    { return "auth.logout" }
func (e AuthLogout) EntityType() string { return "user" }
func (e AuthLogout) EntityID() string   { return e.UserID }

// SessionInvalidatedAll is emitted when all sessions for a user are
// invalidated (e.g., after password change or user deletion).
type SessionInvalidatedAll struct {
	UserID string `json:"user_id"`
}

func (SessionInvalidatedAll) EventName() string    { return "session.invalidated_all" }
func (e SessionInvalidatedAll) EntityType() string { return "user" }
func (e SessionInvalidatedAll) EntityID() string   { return e.UserID }
