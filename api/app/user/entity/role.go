package entity

// Role represents a row from the roles table.
type Role struct {
	ID   string `db:"id"   json:"id"`
	Slug string `db:"slug" json:"slug"`
	Name string `db:"name" json:"name"`
}

// UserWithRole is a joined view of a user and their role, used by admin
// list and detail endpoints. Fields use prefixed db tags to match the ORM's
// automatic column aliasing for dotted selects (e.g. "u.id" → "u.id AS \"u.id\"").
type UserWithRole struct {
	ID                  string  `db:"u.id"                    json:"id"`
	Email               string  `db:"u.email"                 json:"email"`
	Name                string  `db:"u.name"                  json:"name"`
	ForcePasswordChange bool    `db:"u.force_password_change" json:"force_password_change"`
	DeletedAt           *string `db:"u.deleted_at"            json:"deleted_at,omitempty"`
	CreatedAt           string  `db:"u.created_at"            json:"created_at"`
	UpdatedAt           string  `db:"u.updated_at"            json:"updated_at"`
	RoleSlug            string  `db:"r.slug"                  json:"role_slug"`
	RoleName            string  `db:"r.name"                  json:"role_name"`
}
