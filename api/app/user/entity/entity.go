package entity

import "localhost/app/core/sqlite/orm"

type User struct {
	orm.BaseModel
	RoleID              string  `db:"role_id"               json:"role_id"`
	Email               string  `db:"email"                 json:"email"`
	Name                string  `db:"name"                  json:"name"`
	PasswordHash        string  `db:"password_hash"         json:"-"`
	ForcePasswordChange bool    `db:"force_password_change" json:"force_password_change"`
	DeletedAt           *string `db:"deleted_at"            json:"deleted_at,omitempty"`
}
